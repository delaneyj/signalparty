package flimsy

import (
	"fmt"
	"math/rand"
	"reflect"

	mapset "github.com/deckarep/golang-set/v2"
)

type callback[T any] func() (T, error)
type equalsFunction[T any] func(prev T, next T) bool
type errorFunction func(error error)
type rootFunction func(dispose func())
type updateFunction[T any] func(prev T) T

// Type for the getter function for signals
type Getter[T any] interface {
	Read() T
}

// Type for the setter function for signals
type Setter[T any] interface {
	// It can either be called with an update function, which will be called with the current value
	WriteFunc(update updateFunction[T]) T
	// Or the new value directly
	Write(value T) T
}

type Context[T any] struct {
	runtime *Runtime
	// Unique identifier for the context
	id int64
	// Default value for the context
	defaultValue T
}

func (c *Context[T]) Read() T {
	// If the getter finds null or undefined as the value then the default value is returned instead
	if t, ok := observerGet[T](c.runtime.observer, c.id); ok {
		return t
	}
	return c.defaultValue
}

func (c *Context[T]) Write(value T) {
	observerSet(c.runtime.observer, c.id, value)
}

type Runtime struct {
	// It says whether we are currently batching and where to keep the pending values
	batch map[*Signal[any]]any
	// It says what the current observer is, depending on the call stack, if any
	observer *Observer
	// Whether signals should register themselves as dependencies for the parent computation or not
	tracking bool
	// Unique symbol for errors, so that we can store them in the context and reuse the code for that
	symbolErrors int64
}

/* OBJECTS */

// Useless class wrapper
// Function that executes a function and sets the OBSERVER and TRACKING variables
// Basically it keeps track of what the previous OBSERVER and TRACKING values were, sets the new ones, and then restores the old ones back after the function has finished executing
func wrap[T any](runtime *Runtime, fn callback[T], observer *Observer, tracking bool) (T, error) {
	observerPrev := runtime.observer
	trackingPrev := runtime.tracking
	reset := func() {
		runtime.observer = observerPrev
		runtime.tracking = trackingPrev
	}
	defer reset()

	runtime.observer = observer
	runtime.tracking = tracking

	// Important to wrap this in a try..catch as the function may throw, messing up the restoration
	t, err := fn()
	if err != nil {
		// Catching the error, as the observer, or one of its ancestors, may be able to handle it via an error handler

		// Getting the closest error handlers
		var (
			fns []errorFunction
		)
		if runtime.observer != nil {
			fns, _ = observerGet[[]errorFunction](runtime.observer, runtime.symbolErrors)
		}

		// Some handlers, just calling them then
		if len(fns) > 0 {
			for _, fn := range fns {
				fn(err)
			}
		} else {
			// No handlers, throwing
			return t, err
		}

	}

	return t, nil
}

// Signals make values reactive, as going through function calls to get/set values for them enables the automatic dependency tracking and computation re-execution
type Signal[T any] struct {
	runtime *Runtime

	// It's important to keep track of the "parent" memo, if any, because we need to know when reading a signal if it belongs to a parent which isn't up to date, so that we can refresh it in that case
	parent *Observer
	// The current value of the signal
	value T
	// The equality function
	equals equalsFunction[T]
	// List of observers to notify when the value of the signal changes
	// It's a set because sooner or later we must deduplicate registrations
	// Like, if a signal is read multiple times inside an observer the observer must still be only refreshed once when that signal is updated
	observers mapset.Set[*Observer]
}

func newSignal[T any](runtime *Runtime, value T) *Signal[T] {
	return &Signal[T]{
		runtime:   runtime,
		value:     value,
		observers: mapset.NewSet[*Observer](),
	}
}

func (s *Signal[T]) asAny() *Signal[any] {
	var x interface{} = s
	y, ok := x.(*Signal[any])
	if !ok {
		panic("not a Signal[any]")
	}
	return y
}

// Getting the value from the signal
func (s *Signal[T]) Read() T {
	// Registering the signal as a dependency, if we are tracking and the parent is a computation (which can be re-executed, in contrast with roots for example)

	panic("OBSERVER instanceof Computation")
	if s.runtime.tracking && s.runtime.observer != nil {
		s.observers.Add(s.runtime.observer)

		s.runtime.observer.signals.Add(s.asAny())
	}

	// There is a parent and it's stale, we need to refresh it first
	// Refreshing the parent may cause other computations to be refreshed too, if needed
	// If we don't do this we get a "glitch", your code could simulaneously see values that don't make sense toghether, like "count" === 3 and "doubleCount" === 4 because it hasn't been updated yet maybe
	if s.parent != nil && s.parent.waiting() != 0 {
		s.parent.update()
	}

	return s.value
}

// Updating the value
func (s *Signal[T]) Write(next T) T {
	// Are they equal according to the equals function? If they are there's nothing to do, nothing changed, nothing to re-run
	if reflect.DeepEqual(s.value, next) {
		return s.value
	}

	// Are we batching? If so let's store this new value for later
	if s.runtime.batch != nil {
		s.runtime.batch[s.asAny()] = next
	} else {
		// Setting the new value for the signal
		s.value = next

		// Notifying observers now

		// First of all the observers and their observers and so on are marked as stale
		// We also tell them that something actually changed, so when it comes down to it they should update themselves
		s.stale(1, true)

		// Then they are marked as non-stale
		// We also tell them that something actually changed, so when it comes down to it they should update themselves
		s.stale(-1, true)

		// It looks silly but this is crucial
		// Basically if we don't do that computations might be executed multiple times
		// We want to execute them as few times as possible to get the best performance
		// Also while Flimsy doesn't care about performance notifying observers like this is easy and robust
	}

	return s.value
}

func (s *Signal[T]) WriteFn(fn func(prev T) T) T {
	return s.Write(fn(s.value))
}

// Propagating change of the "stale" status to every observer of this signal
// +1 means a signal you depend on is stale, wait for it
// -1 means a signal you depend on just became non-stale, maybe you can update yourself now if you are not waiting for anything else
// The "fresh" value tells observers whether something actually changed or not
// If nothing changed, not for this signal nor for any other signal that a computation is listening to, then the computation will just not be re-executed, for performance
// If at least one signal changed the computation will eventually be re-executed
func (s *Signal[T]) stale(change int, fresh bool) {
	// Propagating the change to every observer of this signal
	observers := s.observers.ToSlice()
	for _, observer := range observers {
		observer.stale(change, fresh)
	}
}

// An observer is something that can have signals as dependencies
type Observer struct {
	runtime *Runtime

	// The parent observer, if any, we need this because context reads and errors kind of bubble up
	parent *Computation[any]
	// List of custom cleanup functions to call
	cleanups []callback[any]
	// Object containg data for the context, plus error handlers, if any, since we are putting those there later in this file
	contexts map[int64]any
	// List of child observers, we need this because when this observer is disposed it has to tell its children to dispose themselves too
	observers mapset.Set[*Observer]
	// List of signals that this observer depends on, we need this because when this observer is disposed it has to tell signals to not refresh it anymore
	signals mapset.Set[*Signal[any]]

	waiting func() int
	update  func()
	stale   func(change int, fresh bool)
}

// Disposing, clearing everything
func (o *Observer) dispose() {
	// Clearing child observers, recursively
	for observer := range o.observers.Iter() {
		observer.dispose()
	}

	// Clearing signal dependencies
	for signal := range o.signals.Iter() {
		signal.observers.Remove(o)
	}

	// Calling custom cleanup functions
	for _, cleanup := range o.cleanups {
		cleanup()
	}

	// Actually emptying the intenral objects
	o.cleanups = o.cleanups[:0]
	o.contexts = map[int64]any{}
	o.observers.Clear()
	o.signals.Clear()

	// Unlinking it also from the parent, not doing this will cause memory leaks because this observer won't be garbage-collected as long as its parent is alive
	if o.parent != nil {
		o.parent.observers.Remove(o)
	}
}

// Getting something from the context
func observerGet[T any](o *Observer, id int64) (v T, ok bool) {
	// Do we have a value for this id?
	var x any
	if x, ok = o.contexts[id]; ok {
		v, ok = x.(T)
		return
	} else {
		// Does the parent have a value for this id?
		if o.parent != nil {
			return observerGet[T](o.parent.Observer, id)
		}
	}
	return
}

// Setting something in the context
func observerSet[T any](o *Observer, id int64, value T) {
	o.contexts[id] = value
}

// A root is a special kind of observer, the function passed to it receives the "dispose" function
// Plus in contrast to Computations the function here will not be re-executed
// Plus a root doesn't link itself with its parent, so the parent won't dispose of child roots simply because it doesn't know about them. As a consequence you'll have to eventually dispose of roots yourself manually
// Still the Root has to know about its parent, because contexts reads and errors bubble up
type Root struct {
	*Observer
}

func wrapRoot[T any](r *Root, fn func(dispose func())) error {
	// Making a customized function, so that we can reuse the Wrapper.wrap function, which doesn't pass anything to our function

	// Calling our function, with "this" as the current observer, and "false" as the value for TRACKING
	if _, err := wrap[T](
		r.runtime,
		func() (T, error) {
			fn(r.dispose)
			var t T
			return t, nil
		},
		r.Observer,
		false,
	); err != nil {
		return fmt.Errorf("can't wrap root: %w", err)
	}

	return nil
}

// A computation is an observer like a root, but it can be re-executed and it can be disposed from its parent
type Computation[T any] struct {
	*Observer

	// Function to potentially re-execute
	fn callback[T]
	// Internal signal holding the last value returned by the function
	signal *Signal[T]
	// Little counter to keep track of the stale status of this computation
	// waiting > 0 means that number of our dependencies are stale, so we should wait for them if we can
	// waiting === 0 means this computation contains a fresh value, it's not waiting for anything, all of its dependencies are up-to-date
	// waiting < 0 doesn't make sense and never happens
	waiting int
	// The fresh flag tells the computation whether one of its dependencies changed or not, if some of its dependencies got re-executed but nothing really changed then we just don't re-execute this computation
	fresh bool
}

func newComputation[T any](runtime *Runtime, fn callback[T]) (*Computation[T], error) {
	c := &Computation[T]{
		Observer: &Observer{
			runtime: runtime,
		},
		fn: fn,
	}

	// Creating the internal signal, we have a dedicated "run" function because we don't want to call `signal.set` the first time, because if we did that we might have a bug if we are using a custom equality comparison as that would be called with "undefined" as the current value the first time
	t, err := c.run()
	if err != nil {
		return nil, fmt.Errorf("error while running the computation: %w", err)
	}
	c.signal = CreateSignal(runtime, t)

	// Linking this computation with the parent, so that we can get a reference to the computation from the signal when we want to check if the computation is stale or not
	c.signal.parent = c.Observer

	c.Observer.waiting = func() int {
		return c.waiting
	}

	return c, nil
}

/* API */

// Execute the computation
// It first disposes of itself basically
// Then it re-executes itself
// This way dynamic dependencies become possible also
func (c *Computation[T]) run() (T, error) {
	// Disposing
	c.dispose()

	// Linking with parent again
	if c.parent != nil {
		c.parent.observers.Add(c.Observer)
	}

	// Doing whatever the function does, "this" becomes the observer and "true" means we are tracking the function
	return wrap(c.runtime, c.fn, c.Observer, true)
}

// Same as run, but also update the signal
func (c *Computation[T]) update() error {
	// Resetting "waiting", as it may be > 0 here if the computation got forcefully refreshed
	c.waiting = 0

	// Resettings "fresh" for the next computation
	c.fresh = false

	// Doing whatever run does and updating the signal
	t, err := c.run()
	if err != nil {
		return fmt.Errorf("computation error: %w", err)
	}
	c.signal.Write(t)
	return nil
}

// Propagating change of the "stale" status to every observer of the internal signal
// Propagating a "false" "fresh" status too, it will be the signal itself that will propagate a "true" one when and if it will actually change
func (c *Computation[T]) stale(change int, fresh bool) error {
	// If this.waiting is already 0 but change is -1 it means the computation got forcefully refreshed already
	// So there's nothing to do here, refreshing again would be wasteful and setting this to -1 would be non-sensical
	if c.waiting == 0 && change < 0 {
		return nil
	}

	// Marking computations depending on us as stale
	// We only need to do this once, when the "waiting" counter goes from 0 to 1
	// We also tell them that nothing changed, becuase we don't know if something will change yet
	if c.waiting == 0 && change > 0 {
		c.signal.stale(1, false)
	}

	// Update the counter
	c.waiting += change

	// Internally we need to use the "fresh" status we recevied to understand if at least one of our dependencies changed
	if fresh {
		c.fresh = true
	}

	// Are we still waiting for something?
	if c.waiting != 0 {
		// Resetting the counter now, as maybe the update function won't be executed
		c.waiting = 0

		// Did something actually change? If so we actually update
		if c.fresh {
			if err := c.update(); err != nil {
				return fmt.Errorf("computation error: %w", err)
			}
		}

		// Now finally we mark computations depending on us as unstale
		// We still tell them that we don't know if something changed here
		// if something changed the signal itself will propagate its own true "fresh" status
		c.signal.stale(-1, false)
	}

	return nil
}

// /* METHODS */

func CreateSignal[T any](runtime *Runtime, value T) *Signal[T] {
	return newSignal[T](runtime, value)
}

// An effect is just a computation that doesn't return anything
func CreateEffect[T any](runtime *Runtime, fn callback[T]) {
	newComputation(runtime, fn)
}

// A memo is a computation that returns a getter to its internal signal, which holds the last return value of the function
func CreateMemo[T any](runtime *Runtime, fn callback[T]) (Getter[T], error) {
	getter, err := newComputation(runtime, fn)
	if err != nil {
		return nil, fmt.Errorf("error while setting up the computation: %w", err)
	}
	return getter.signal, nil
}

// A root is just a plain observer that exposes the "dispose" method and that will survive its parent observer being disposed
// Roots are essential for achieving great performance with things like <For> in Solid
func CreateRoot[T any](runtime *Runtime, fn callback[T]) T {
	r := &Root{
		Observer: &Observer{
			runtime: runtime,
		},
	}
	return wrapRoot(r, fn)
}

func CreateContext[T any](runtime *Runtime, defaultValue T) *Context[T] {
	// Making a new identifier for this context
	id := rand.Int63()

	// Making get/set functions dedicated to this context
	// If the getter finds null or undefined as the value then the default value is returned instead

	c := &Context[T]{
		runtime:      runtime,
		id:           id,
		defaultValue: defaultValue,
	}
	return c
}

func UseContext[T any](c *Context[T]) T {
	// Just calling the getter
	// this function is implemented for compatibility with Solid
	// Solid's implementation of this is a bit more interesting because it doesn't expose a "get" method on the context directly that can just be called like this
	return c.Read()
}

func OnCleanup[T any](fn callback[T]) {
	// If there's a current observer
}

// function onCleanup ( fn: Callback ): void {

//   // If there's a current observer let's add a cleanup function to it
//   OBSERVER?.cleanups.push ( fn );

// }

// function onError ( fn: ErrorFunction ): void {

//   if ( !OBSERVER ) return;

//   // If there's a current observer let's add an error handler function to it, ensuring the array containing these functions exists first though
//   OBSERVER.contexts[SYMBOL_ERRORS] ||= [];
//   OBSERVER.contexts[SYMBOL_ERRORS].push ( fn );

// }

// // Batching is an important performance feature, it holds onto updates until the function has finished executing, so that computations are later re-executed the minimum amount of times possible
// // Like if you change a signal in a loop, for some reason, then without batching its observers will be re-executed with each iteration
// // With batching they are only executed at this end, potentially just 1 time instead of N times
// // While batching is active the getter will give you the "old" value of the signal, as the new one hasn't actually been be set yet
// function batch <T> ( fn: Callback<T> ): T {

//   // Already batching? Nothing else to do then
//   if ( BATCH ) return fn ();

//   // New batch bucket where to store upcoming values for signals
//   const batch = BATCH = new Map<Signal, any> ();

//   // Important to use a try..catch as the function may throw, messing up our flushing of updates later on
//   try {

//     return fn ();

//   } finally {

//     // Turning batching off
//     BATCH = undefined;

//     // Marking all the signals as stale, all at once, or each update to each signal will cause its observers to be updated, but there might be observers listening to multiple of these signals, we want to execute them once still if possible
//     // We don't know if something will change, so we set the "fresh" flag to "false"
//     batch.forEach ( ( value, signal ) => signal.stale ( 1, false ) );
//     // Updating values
//     batch.forEach ( ( value, signal ) => signal.set ( () => value ) );
//     // Now making all those signals as not stale, allowing observers to finally update themselves
//     // We don't know if something did change, so we set the "fresh" flag to "false"
//     batch.forEach ( ( value, signal ) => signal.stale ( -1, false ) );

//   }

// }

// function untrack <T> ( fn: Callback<T> ): T {

//   // Turning off tracking
//   // The observer stays the same, but TRACKING is set to "false"
//   return Wrapper.wrap ( fn, OBSERVER, false );

// }

// /* EXPORT */

// export {createContext, createEffect, createMemo, createRoot, createSignal, onCleanup, onError, useContext, batch, untrack};
// export type {Getter, Setter, Context, Options};
