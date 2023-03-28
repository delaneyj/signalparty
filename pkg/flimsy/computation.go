package flimsy

import "fmt"

// A computation is an observer like a root, but it can be re-executed and it can be disposed from its parent
type computation struct {
	*observer

	// Function to potentially re-execute
	fn callback
	// Internal signal holding the last value returned by the function
	signal *signal
	// Little counter to keep track of the stale status of this computation
	// waiting > 0 means that number of our dependencies are stale, so we should wait for them if we can
	// waiting === 0 means this computation contains a fresh value, it's not waiting for anything, all of its dependencies are up-to-date
	// waiting < 0 doesn't make sense and never happens
	waiting int
	// The fresh flag tells the computation whether one of its dependencies changed or not, if some of its dependencies got re-executed but nothing really changed then we just don't re-execute this computation
	fresh bool
}

func newComputation(runtime *ReactiveContext, fn callback) (*computation, error) {
	c := &computation{
		observer: newObserver(runtime),
		fn:       fn,
	}

	// Creating the internal signal, we have a dedicated "run" function because we don't want to call `signal.set` the first time, because if we did that we might have a bug if we are using a custom equality comparison as that would be called with "undefined" as the current value the first time
	t, err := c.run()
	if err != nil {
		return nil, fmt.Errorf("error while running the computation: %w", err)
	}

	c.signal = newSignal(runtime, t)

	// Linking this computation with the parent, so that we can get a reference to the computation from the signal when we want to check if the computation is stale or not
	c.signal.parent = c.observer

	c.observer.waiting = func() int {
		return c.waiting
	}

	c.observer.stale = c.stale

	return c, nil
}

/* API */

// Execute the computation
// It first disposes of itself basically
// Then it re-executes itself
// This way dynamic dependencies become possible also
func (c *computation) run() (any, error) {
	// Disposing
	c.dispose()

	// Linking with parent again
	if c.parent != nil {
		c.parent.observers.Add(c.observer)
	}

	// Doing whatever the function does, "this" becomes the observer and "true" means we are tracking the function
	return wrap(c.runtime, c.fn, c.observer, true)
}

// Same as run, but also update the signal
func (c *computation) update() error {
	// Resetting "waiting", as it may be > 0 here if the computation got forcefully refreshed
	c.waiting = 0

	// Resettings "fresh" for the next computation
	c.fresh = false

	// Doing whatever run does and updating the signal
	t, err := c.run()
	if err != nil {
		return fmt.Errorf("computation error: %w", err)
	}
	c.signal.set(t)
	return nil
}

// Propagating change of the "stale" status to every observer of the internal signal
// Propagating a "false" "fresh" status too, it will be the signal itself that will propagate a "true" one when and if it will actually change
func (c *computation) stale(change int, fresh bool) error {
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
