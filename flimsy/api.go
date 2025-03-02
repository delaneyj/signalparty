package flimsy

import (
	"fmt"
	"math/rand"
)

type Callback[T any] func() (T, error)
type RootFunction[T any] func(dispose func()) T

type Getter[T any] func() T
type Setter[T any] func(value T)

func CreateSignal[T any](runtime *ReactiveContext, value T) (Getter[T], Setter[T]) {
	s := newSignal(runtime, value)

	return signalGetter[T](s), signalSetter[T](s)
}

// An effect is just a computation that doesn't return anything
func CreateEffect[T any](runtime *ReactiveContext, fn Callback[T]) {
	newComputation(runtime, func() (any, error) {
		return fn()
	})
}

// A memo is a computation that returns a getter to its internal signal, which holds the last return value of the function
func CreateMemo[T any](runtime *ReactiveContext, fn Callback[T]) (Getter[T], error) {
	getter, err := newComputation(runtime, func() (any, error) {
		return fn()
	})
	if err != nil {
		return nil, fmt.Errorf("error while setting up the computation: %w", err)
	}
	return signalGetter[T](getter.signal), nil
}

// A root is just a plain observer that exposes the "dispose" method and that will survive its parent observer being disposed
// Roots are essential for achieving great performance with things like <For> in Solid
func CreateRoot[T any](runtime *ReactiveContext, fn RootFunction[T]) (T, error) {
	r := &root{
		observer: newObserver(runtime),
	}

	fnAny := func(dispose func()) any {
		return fn(dispose)
	}

	var t T
	x, err := wrapRoot(r, fnAny)
	if err != nil {
		return t, fmt.Errorf("error while running the root: %w", err)
	}
	t, ok := x.(T)
	if !ok {
		panic("invalid type")
	}

	return t, nil
}

func CreateContext[T any](runtime *ReactiveContext, defaultValue T) *Context[T] {
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

func OnCleanup[T any](r *ReactiveContext, fn Callback[T]) {
	// If there's a current observer
	if r.observer != nil {
		// convert to callback
		cb := func() (any, error) {
			return fn()
		}

		// Let's add a cleanup function to it
		r.observer.cleanups = append(r.observer.cleanups, cb)
	}

}

// Batch is an important performance feature, it holds onto updates until the function has finished executing, so that computations are later re-executed the minimum amount of times possible
// Like if you change a signal in a loop, for some reason, then without batching its observers will be re-executed with each iteration
// With batching they are only executed at this end, potentially just 1 time instead of N times
// While batching is active the getter will give you the "old" value of the signal, as the new one hasn't actually been be set yet
func Batch[T any](r *ReactiveContext, fn Callback[T]) (T, error) {
	// Already batching? Nothing else to do then
	if r.batch != nil {
		return fn()
	}

	defer func() {
		// Turning off batching
		r.batch = nil

		// Marking all the signals as stale, all at once, or each update to each signal will cause its observers to be updated, but there might be observers listening to multiple of these signals, we want to execute them once still if possible
		// We don't know if something will change, so we set the "fresh" flag to "false"
		for signal := range r.batch {
			signal.stale(1, false)
		}
		// Updating values
		for signal, value := range r.batch {
			signal.set(value)
		}

		// Now making all those signals as not stale, allowing observers to finally update themselves
		// We don't know if something did change, so we set the "fresh" flag to "false"
		for signal := range r.batch {
			signal.stale(-1, false)
		}
	}()

	// New batch bucket where to store upcoming values for signals
	r.batch = map[*signal]any{}

	// Important to use a try..catch as the function may throw, messing up our flushing of updates later on
	t, err := fn()
	if err != nil {
		return t, fmt.Errorf("error while running batched callback: %w", err)
	}

	return t, nil
}

// Turn off tracking, the observer stays the same, but TRACKING is set to "false"
func Untrack[T any](runtime *ReactiveContext, fn Callback[T]) (T, error) {
	cb := func() (any, error) {
		return fn()
	}
	var t T
	x, err := wrap(runtime, cb, runtime.observer, false)
	if err != nil {
		return t, fmt.Errorf("error while running untracked callback: %w", err)
	}
	t, ok := x.(T)
	if !ok {
		panic("invalid type")
	}

	return t, nil
}
