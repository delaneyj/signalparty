package reactively

type CacheState int

const (
	CacheClean CacheState = iota // reactive value is valid, no need to recompute
	CacheCheck                   // reactive value might be stale, check parent nodes to decide whether to recompute
	CacheDirty                   // reactive value is invalid, parents have changed, valueneeds to be recomputed
)

type CleanupFunc[T comparable] func(oldValue T)

type ReactiveContext struct {
	current         HasReactivity
	currentGets     []any
	currentGetIndex int
	effectQueue     []any
}

type reactiveOrCleanup interface {
	isReactiveOrCleanup()
}

func (c CleanupFunc[T]) isReactiveOrCleanup() {}
func (r *Reactive[T]) isReactiveOrCleanup()   {}

type Reactive[T comparable] struct {
	rctx      *ReactiveContext
	value     T
	fn        func() T
	sources   []HasReactivity
	observers []reactiveOrCleanup
	state     CacheState
	isEffect  bool
	cleanups  []CleanupFunc[T]
}

// For template literals
func (r *Reactive[T]) CanBeAttribute() {}
func (r *Reactive[T]) IsStatic() bool {
	return false
}

type HasReactivity interface {
	getState() CacheState
	setState(CacheState)
	getSources() []any
	stale(CacheState)
	getObservers() []reactiveOrCleanup
	setObservers([]reactiveOrCleanup)
	updateIfNecessary()
}

func (r *Reactive[T]) getState() CacheState {
	return r.state
}

func (r *Reactive[T]) setState(state CacheState) {
	r.state = state
}

func (r *Reactive[T]) getSources() []any {
	anySources := make([]any, len(r.sources))
	for i, source := range r.sources {
		anySources[i] = source
	}
	return anySources
}

func (r *Reactive[T]) setObservers(observers []reactiveOrCleanup) {
	r.observers = observers
}

func (r *Reactive[T]) getObservers() []reactiveOrCleanup {
	return r.observers
}

func Signal[T comparable](rctx *ReactiveContext, value T) *Reactive[T] {
	return &Reactive[T]{
		rctx:  rctx,
		value: value,
		state: CacheClean,
	}
}

func reactiveFunction[T comparable](rctx *ReactiveContext, fn func() T, isEffect bool) *Reactive[T] {
	r := &Reactive[T]{
		rctx:     rctx,
		fn:       fn,
		isEffect: isEffect,
		state:    CacheDirty,
	}
	if r.isEffect {
		r.update() // CONSIDER removing this?
	}
	return r
}

func Memo[T comparable](rctx *ReactiveContext, fn func() T) *Reactive[T] {
	return reactiveFunction(rctx, fn, false)
}

func Effect(rctx *ReactiveContext, fn func()) {
	fn2 := func() bool {
		fn()
		return false
	}
	reactiveFunction(rctx, fn2, true)
}

func (r *Reactive[T]) Read() T {
	if r.rctx.current != nil {
		sources := r.rctx.current.getSources()

		if len(r.rctx.currentGets) > 0 &&
			len(sources) > r.rctx.currentGetIndex &&
			sources[r.rctx.currentGetIndex] == r {
			r.rctx.currentGetIndex++
		} else {
			r.rctx.currentGets = append(r.rctx.currentGets, r)
		}
	}

	if r.fn != nil {
		r.updateIfNecessary()
	}

	return r.value
}

func (r *Reactive[T]) WriteFn(nextValueFn func() T) {
	// original JS has a check here to see if the function is the same
	// As far as I know Go doesn't have a way to do function equality checks
	// so I'm just going to assume that it's always different and pay the cost
	r.stale(CacheDirty) // TODO: is this correct? Naively I'd think it should be after assignment or both
	r.fn = nextValueFn
}

func (r *Reactive[T]) Write(nextValue T) {
	if r.fn != nil {
		r.removeParentObservers(0)
		r.sources = r.sources[:0]
		r.fn = nil
	}
	if r.value != nextValue {
		for _, ob := range r.observers {
			switch ob := ob.(type) {
			case HasReactivity:
				ob.stale(CacheDirty)
			case CleanupFunc[T]:
				// do nothing
			default:
				panic("not a valid observer")
			}
		}
		r.value = nextValue
	}

}

func (r *Reactive[T]) stale(state CacheState) {
	if state == CacheClean {
		panic("should not be called with CacheClean")
	}
	if r.state < state {
		// If we were previously clean, then we know that we may need to update to get the new value
		r.state = state
		for _, observerRaw := range r.observers {
			switch ob := observerRaw.(type) {
			case HasReactivity:
				ob.stale(CacheCheck)
			default:
				panic("not a reactive value")
			}
		}
	}
}

// run the computation fn, updating the cached value
func (rf *Reactive[T]) update() {
	oldValue := rf.value

	// Evalute the reactive function body, dynamically capturing any other reactives used
	prevReaction := rf.rctx.current
	prevGets := make([]any, len(rf.rctx.currentGets))
	copy(prevGets, rf.rctx.currentGets)
	prevIndex := rf.rctx.currentGetIndex

	rf.rctx.current = rf
	rf.rctx.currentGets = nil
	rf.rctx.currentGetIndex = 0

	for _, cleanup := range rf.cleanups {
		cleanup(oldValue)
	}
	rf.cleanups = nil
	rf.value = rf.fn()

	// if the sources have changed, update source & observer links
	if len(rf.rctx.currentGets) > 0 {
		// remove all old sources' .observers links to us
		rf.removeParentObservers(rf.rctx.currentGetIndex)

		sources := make([]HasReactivity, 0, len(rf.rctx.currentGets))
		for _, source := range rf.rctx.currentGets {
			switch source := source.(type) {
			case HasReactivity:
				sources = append(sources, source)
			case CleanupFunc[T]:
				// do nothing
			default:
				panic("not a valid observer")
			}
		}

		// update source up links
		if len(rf.rctx.currentGets) > 0 && rf.rctx.currentGetIndex > 0 {
			rf.sources = append(rf.sources, sources...)
		} else {
			rf.sources = sources
		}

		// Add ourselves to the end of the parent observers array
		for _, source := range rf.sources[rf.rctx.currentGetIndex:] {
			source.setObservers(append(source.getObservers(), rf))
		}
	} else if len(rf.sources) > 0 && rf.rctx.currentGetIndex < len(rf.sources) {
		// remove all old sources' observers links to us
		rf.removeParentObservers(rf.rctx.currentGetIndex)
		rf.sources = rf.sources[:rf.rctx.currentGetIndex]
	}

	rf.rctx.currentGets = prevGets
	rf.rctx.current = prevReaction
	rf.rctx.currentGetIndex = prevIndex

	// handle diamond depenendencies if we're the parent of a diamond.
	if oldValue != rf.value && len(rf.observers) > 0 {
		// We've changed value, so mark our children as dirty so they'll reevaluate
		for _, obRaw := range rf.observers {
			switch ob := obRaw.(type) {
			case HasReactivity:
				ob.setState(CacheDirty)
			case CleanupFunc[T]:
				// noop
			default:
				panic("not a valid observer")
			}
		}
	}

	rf.state = CacheClean
}

// if dirty, or a parent turns out to be dirty.
func (rf *Reactive[T]) updateIfNecessary() {
	// If we are potentially dirty, see if we have a parent who has actually changed value
	if rf.state == CacheCheck {
		for _, source := range rf.sources {
			// can change this.state
			source.updateIfNecessary()
			if source.getState() == CacheDirty {
				rf.state = CacheDirty
				// Stop the loop here so we won't trigger updates on other parents unnecessarily
				// If our computation changes to no longer use some sources, we don't
				// want to update() a source we used last time, but now don't use.
				break
			}
		}
	}

	// If we were already dirty or marked dirty by the step above, update.
	if rf.state == CacheDirty {
		rf.update()
	}

	// By now, we're clean
	rf.state = CacheClean
}

func (rf *Reactive[T]) removeParentObservers(startIndex int) {
	for _, source := range rf.sources[startIndex:] {
		currentObservers := source.getObservers()
		revisedObservers := make([]reactiveOrCleanup, 0, len(currentObservers))
		for _, observer := range currentObservers {
			if observer != rf {
				revisedObservers = append(revisedObservers, observer)
			}
		}
		source.setObservers(revisedObservers)
	}
}

func OnCleanup[T comparable](rctx *ReactiveContext, fn CleanupFunc[T]) {
	if rctx.current == nil {
		panic("onCleanup must be called from within a @reactive function")
	}

	switch current := rctx.current.(type) {
	case *Reactive[T]:
		current.cleanups = append(current.cleanups, fn)
		current.observers = append(current.observers, fn)
	default:
		panic("not a reactive value")
	}
}

// run all non-clean effect nodes
func Stablize[T comparable](rctx *ReactiveContext) {
	for _, effect := range rctx.effectQueue {
		switch effect := effect.(type) {
		case *Reactive[T]:
			effect.stale(CacheCheck)
		default:
			panic("not a reactive value")
		}
	}
}
