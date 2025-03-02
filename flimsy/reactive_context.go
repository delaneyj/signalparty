package flimsy

type ReactiveContext struct {
	// It says whether we are currently batching and where to keep the pending values
	batch map[*signal]any
	// It says what the current observer is, depending on the call stack, if any
	observer *observer
	// Whether signals should register themselves as dependencies for the parent computation or not
	tracking bool
	// Unique symbol for errors, so that we can store them in the context and reuse the code for that
	symbolErrors int64
}
