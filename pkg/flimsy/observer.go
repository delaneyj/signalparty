package flimsy

import mapset "github.com/deckarep/golang-set/v2"

// An observer is something that can have signals as dependencies
type observer struct {
	runtime *ReactiveContext

	// The parent observer, if any, we need this because context reads and errors kind of bubble up
	parent *computation
	// List of custom cleanup functions to call
	cleanups []callback
	// Object containg data for the context, plus error handlers, if any, since we are putting those there later in this file
	contexts map[int64]any
	// List of child observers, we need this because when this observer is disposed it has to tell its children to dispose themselves too
	observers mapset.Set[*observer]
	// List of signals that this observer depends on, we need this because when this observer is disposed it has to tell signals to not refresh it anymore
	signals mapset.Set[*signal]

	waiting func() int
	update  func()
	stale   func(change int, fresh bool) error
}

func newObserver(runtime *ReactiveContext) *observer {
	return &observer{
		runtime:   runtime,
		contexts:  map[int64]any{},
		observers: mapset.NewSet[*observer](),
		signals:   mapset.NewSet[*signal](),
	}
}

// Disposing, clearing everything
func (o *observer) dispose() {
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
func (o *observer) get(id int64) (v any, ok bool) {
	// Do we have a value for this id?
	if v, ok = o.contexts[id]; ok {
		return
	} else {
		// Does the parent have a value for this id?
		if o.parent != nil {
			return o.parent.observer.get(id)
		}
	}
	return
}

// Setting something in the context
func (o *observer) set(id int64, value any) {
	o.contexts[id] = value
}
