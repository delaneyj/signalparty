package flimsy

import (
	"reflect"

	mapset "github.com/deckarep/golang-set/v2"
)

// Signals make values reactive, as going through function calls to get/set values for them enables the automatic dependency tracking and computation re-execution
type signal struct {
	runtime *ReactiveContext

	// It's important to keep track of the "parent" memo, if any, because we need to know when reading a signal if it belongs to a parent which isn't up to date, so that we can refresh it in that case
	parent *observer
	// The current value of the signal
	value any

	// List of observers to notify when the value of the signal changes
	// It's a set because sooner or later we must deduplicate registrations
	// Like, if a signal is read multiple times inside an observer the observer must still be only refreshed once when that signal is updated
	observers mapset.Set[*observer]
}

func newSignal(runtime *ReactiveContext, value any) *signal {
	return &signal{
		runtime:   runtime,
		value:     value,
		observers: mapset.NewSet[*observer](),
	}
}

// Propagating change of the "stale" status to every observer of this signal
// +1 means a signal you depend on is stale, wait for it
// -1 means a signal you depend on just became non-stale, maybe you can update yourself now if you are not waiting for anything else
// The "fresh" value tells observers whether something actually changed or not
// If nothing changed, not for this signal nor for any other signal that a computation is listening to, then the computation will just not be re-executed, for performance
// If at least one signal changed the computation will eventually be re-executed
func (s *signal) stale(change int, fresh bool) {
	// Propagating the change to every observer of this signal
	observers := s.observers.ToSlice()
	for _, observer := range observers {
		observer.stale(change, fresh)
	}
}

// Getting the value from the signal
func (s *signal) get() any {
	// Registering the signal as a dependency, if we are tracking and the parent is a computation (which can be re-executed, in contrast with roots for example)

	if s.runtime.tracking && s.runtime.observer != nil {
		s.observers.Add(s.runtime.observer)

		s.runtime.observer.signals.Add(s)
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
func (s *signal) set(next any) any {
	// Are they equal according to the equals function? If they are there's nothing to do, nothing changed, nothing to re-run
	if reflect.DeepEqual(s.value, next) {
		return s.value
	}

	// Are we batching? If so let's store this new value for later
	if s.runtime.batch != nil {
		s.runtime.batch[s] = next
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

func signalGetter[T any](s *signal) Getter[T] {
	return func() T {
		return s.get().(T)
	}
}

func signalSetter[T any](s *signal) Setter[T] {
	return func(next T) {
		s.set(next)
	}
}
