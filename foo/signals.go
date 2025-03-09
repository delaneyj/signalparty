package foo

import "sync"

type Subscriber interface {
	markDirty()
}

type Dependency interface {
	value() any
	version() uint32
	addSubs(...Subscriber)
	removeSub(Subscriber)
}

type WriteableSignal[T comparable] struct {
	val  T
	ver  uint32
	subs []Subscriber
	mu   sync.RWMutex
}

func (s *WriteableSignal[T]) value() any {
	return s.val
}

func (s *WriteableSignal[T]) Value() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.val
}

func (s *WriteableSignal[T]) SetValue(val T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.val == val {
		return
	}
	s.val = val
	s.ver++
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *WriteableSignal[T]) version() uint32 {
	return s.ver
}

func (s *WriteableSignal[T]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *WriteableSignal[T]) removeSub(sub Subscriber) {
	for i, sub := range s.subs {
		if sub == sub {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

func Signal[T comparable](val T) *WriteableSignal[T] {
	return &WriteableSignal[T]{
		val: val,
		ver: 1,
		mu:  sync.RWMutex{},
	}
}

type ReadonlySignal[O comparable] struct {
	val            O
	isDirty        bool
	ver            uint32
	depVersionsSum uint32
	deps           []Dependency
	subs           []Subscriber
	fn             func(args ...any) O
	mu             sync.Mutex
}

func (s *ReadonlySignal[T]) value() any {
	if !s.isDirty {
		return s.val
	}
	s.isDirty = false

	args, sum := depHash(s.deps...)
	if s.depVersionsSum == sum {
		return s.val
	}
	s.depVersionsSum = sum
	newVal := s.fn(args...)
	if s.val == newVal {
		return s.val
	}
	s.val = newVal
	s.ver++
	return s.val
}

func depHash(deps ...Dependency) (args []any, sum uint32) {
	args = make([]any, len(deps))
	for i, dep := range deps {
		args[i] = dep.value()
		sum += dep.version()
	}
	return args, sum
}

func (s *ReadonlySignal[T]) Value() T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value().(T)
}

func (s *ReadonlySignal[T]) version() uint32 {
	return s.ver
}

func (s *ReadonlySignal[T]) markDirty() {
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal[T]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *ReadonlySignal[T]) removeSub(sub Subscriber) {
	for i, sub := range s.subs {
		if sub == sub {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

func newReadonlySignal[O comparable](
	fn func(...any) O,
	deps ...Dependency,
) *ReadonlySignal[O] {
	s := &ReadonlySignal[O]{
		isDirty: true,
		fn:      fn,
		ver:     1,
		deps:    deps,
		mu:      sync.Mutex{},
	}
	for _, dep := range deps {
		dep.addSubs(s)
	}
	return s
}

func Computed1[T0, O comparable](
	arg0 Dependency,
	fn func(T0) O,
) *ReadonlySignal[O] {
	anyFn := func(args ...any) O {
		return fn(args[0].(T0))
	}
	return newReadonlySignal(anyFn, arg0)
}

func Computed2[T0, T1, O comparable](
	arg0, arg1 Dependency,
	fn func(T0, T1) O,
) *ReadonlySignal[O] {
	anyFn := func(args ...any) O {
		return fn(
			args[0].(T0),
			args[1].(T1),
		)
	}
	return newReadonlySignal(anyFn, arg0, arg1)
}

func Computed3[T0, T1, T2, O comparable](
	arg0, arg1, arg2 Dependency,
	fn func(T0, T1, T2) O,
) *ReadonlySignal[O] {
	anyFn := func(args ...any) O {
		return fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
		)
	}
	return newReadonlySignal(anyFn, arg0, arg1, arg2)
}

func Computed4[T0, T1, T2, T3, O comparable](
	arg0, arg1, arg2, arg3 Dependency,
	fn func(T0, T1, T2, T3) O,
) *ReadonlySignal[O] {
	anyFn := func(args ...any) O {
		return fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
		)
	}
	return newReadonlySignal(anyFn, arg0, arg1, arg2, arg3)
}

func Computed5[T0, T1, T2, T3, T4, O comparable](
	arg0, arg1, arg2, arg3, arg4 Dependency,
	fn func(T0, T1, T2, T3, T4) O,
) *ReadonlySignal[O] {
	anyFn := func(args ...any) O {
		return fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
			args[4].(T4),
		)
	}
	return newReadonlySignal(anyFn, arg0, arg1, arg2, arg3, arg4)
}

func Computed6[T0, T1, T2, T3, T4, T5, O comparable](
	arg0, arg1, arg2, arg3, arg4, arg5 Dependency,
	fn func(T0, T1, T2, T3, T4, T5) O,
) *ReadonlySignal[O] {
	anyFn := func(args ...any) O {
		return fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
			args[4].(T4),
			args[5].(T5),
		)
	}
	return newReadonlySignal(anyFn, arg0, arg1, arg2, arg3, arg4, arg5)
}

func Computed7[T0, T1, T2, T3, T4, T5, T6, O comparable](
	arg0, arg1, arg2, arg3, arg4, arg5, arg6 Dependency,
	fn func(T0, T1, T2, T3, T4, T5, T6) O,
) *ReadonlySignal[O] {
	anyFn := func(args ...any) O {
		return fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
			args[4].(T4),
			args[5].(T5),
			args[6].(T6),
		)
	}
	return newReadonlySignal(anyFn, arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}

func Computed8[T0, T1, T2, T3, T4, T5, T6, T7, O comparable](
	arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 Dependency,
	fn func(T0, T1, T2, T3, T4, T5, T6, T7) O,
) *ReadonlySignal[O] {
	anyFn := func(args ...any) O {
		return fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
			args[4].(T4),
			args[5].(T5),
			args[6].(T6),
			args[7].(T7),
		)
	}
	return newReadonlySignal(anyFn, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

type SideEffect struct {
	fn             func(...any)
	depVersionsSum uint32
	deps           []Dependency
}

func (e *SideEffect) markDirty() {
	args, sum := depHash(e.deps...)
	if e.depVersionsSum == sum {
		return
	}
	e.depVersionsSum = sum
	e.fn(args...)
}

func newSideEffect(
	fn func(...any),
	deps ...Dependency,
) (stop func()) {
	e := &SideEffect{
		fn:   fn,
		deps: deps,
	}
	for _, dep := range deps {
		dep.addSubs(e)
	}
	return func() {
		for _, dep := range e.deps {
			dep.removeSub(e)
		}
	}
}

func Effect1[T0 comparable](
	arg0 Dependency,
	fn func(T0),
) (stop func()) {
	anyFn := func(args ...any) {
		fn(args[0].(T0))
	}
	return newSideEffect(anyFn, arg0)
}

func Effect2[T0, T1 comparable](
	arg0, arg1 Dependency,
	fn func(T0, T1),
) (stop func()) {
	anyFn := func(args ...any) {
		fn(
			args[0].(T0),
			args[1].(T1),
		)
	}
	return newSideEffect(anyFn, arg0, arg1)
}

func Effect3[T0, T1, T2 comparable](
	arg0, arg1, arg2 Dependency,
	fn func(T0, T1, T2),
) (stop func()) {
	anyFn := func(args ...any) {
		fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
		)
	}
	return newSideEffect(anyFn, arg0, arg1, arg2)
}

func Effect4[T0, T1, T2, T3 comparable](
	arg0, arg1, arg2, arg3 Dependency,
	fn func(T0, T1, T2, T3),
) (stop func()) {
	anyFn := func(args ...any) {
		fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
		)
	}
	return newSideEffect(anyFn, arg0, arg1, arg2, arg3)
}

func Effect5[T0, T1, T2, T3, T4 comparable](
	arg0, arg1, arg2, arg3, arg4 Dependency,
	fn func(T0, T1, T2, T3, T4),
) (stop func()) {
	anyFn := func(args ...any) {
		fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
			args[4].(T4),
		)
	}
	return newSideEffect(anyFn, arg0, arg1, arg2, arg3, arg4)
}

func Effect6[T0, T1, T2, T3, T4, T5 comparable](
	arg0, arg1, arg2, arg3, arg4, arg5 Dependency,
	fn func(T0, T1, T2, T3, T4, T5),
) (stop func()) {
	anyFn := func(args ...any) {
		fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
			args[4].(T4),
			args[5].(T5),
		)
	}
	return newSideEffect(anyFn, arg0, arg1, arg2, arg3, arg4, arg5)
}

func Effect7[T0, T1, T2, T3, T4, T5, T6 comparable](
	arg0, arg1, arg2, arg3, arg4, arg5, arg6 Dependency,
	fn func(T0, T1, T2, T3, T4, T5, T6),
) (stop func()) {
	anyFn := func(args ...any) {
		fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
			args[4].(T4),
			args[5].(T5),
			args[6].(T6),
		)
	}
	return newSideEffect(anyFn, arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}

func Effect8[T0, T1, T2, T3, T4, T5, T6, T7 comparable](
	arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 Dependency,
	fn func(T0, T1, T2, T3, T4, T5, T6, T7),
) (stop func()) {
	anyFn := func(args ...any) {
		fn(
			args[0].(T0),
			args[1].(T1),
			args[2].(T2),
			args[3].(T3),
			args[4].(T4),
			args[5].(T5),
			args[6].(T6),
			args[7].(T7),
		)
	}
	return newSideEffect(anyFn, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}
