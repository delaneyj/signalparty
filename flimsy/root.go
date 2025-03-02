package flimsy

import "fmt"

// A root is a special kind of observer, the function passed to it receives the "dispose" function
// Plus in contrast to Computations the function here will not be re-executed
// Plus a root doesn't link itself with its parent, so the parent won't dispose of child roots simply because it doesn't know about them. As a consequence you'll have to eventually dispose of roots yourself manually
// Still the Root has to know about its parent, because contexts reads and errors bubble up
type root struct {
	*observer
}

func wrapRoot(r *root, rootFn rootFunction) (any, error) {
	// Making a customized function, so that we can reuse the Wrapper.wrap function, which doesn't pass anything to our function

	// Calling our function, with "this" as the current observer, and "false" as the value for TRACKING
	t, err := wrap(
		r.runtime,
		func() (any, error) {
			t := rootFn(r.dispose)
			return t, nil
		},
		r.observer,
		false,
	)
	if err != nil {
		return t, fmt.Errorf("can't wrap root: %w", err)
	}

	return t, nil
}

// Function that executes a function and sets the OBSERVER and TRACKING variables
// Basically it keeps track of what the previous OBSERVER and TRACKING values were, sets the new ones, and then restores the old ones back after the function has finished executing
func wrap(runtime *ReactiveContext, fn callback, observer *observer, tracking bool) (any, error) {
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
			fnsAny, ok := runtime.observer.get(runtime.symbolErrors)
			if ok {
				fns, ok = fnsAny.([]errorFunction)
				if !ok {
					panic("invalid type")
				}
			}
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
