package rocket

import "sync"

type ReactiveSystem struct {
	mu            *sync.Mutex
	currentUpdate uint32
}

func NewReactiveSystem() *ReactiveSystem {
	rs := &ReactiveSystem{
		mu:            &sync.Mutex{},
		currentUpdate: 0,
	}
	return rs
}

type Cell interface {
	value() any
	markDirty()
	addSubs(...Cell)
	removeSub(Cell)
}

type WriteableSignal[T comparable] struct {
	v    T
	sys  *ReactiveSystem
	subs []Cell
}

func (s *WriteableSignal[T]) Value() T {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()
	return s.v
}

func (s *WriteableSignal[T]) value() any {
	return s.v
}

func (s *WriteableSignal[T]) SetValue(value T) {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()

	if s.v == value {
		return
	}
	s.v = value

	if len(s.subs) == 0 {
		return
	}
	s.sys.currentUpdate++
	s.markDirty()
}

func Signal[T comparable](sys *ReactiveSystem, value T) *WriteableSignal[T] {
	s := &WriteableSignal[T]{v: value, sys: sys}
	return s
}

func (s *WriteableSignal[T]) markDirty() {
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *WriteableSignal[T]) addSubs(sub ...Cell) {
	s.subs = append(s.subs, sub...)
}

func (s *WriteableSignal[T]) removeSub(cell Cell) {
	for i, sub := range s.subs {
		if sub == cell {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type ReadonlySignal1Func[T0, O comparable] func(T0) O

type ReadonlySignal1Args[T0 comparable] struct {
	Arg0 T0
}

type ReadonlySignal1[T0, O comparable] struct {
	sys        *ReactiveSystem
	subs       []Cell
	isDirty    bool
	lastUpdate uint32
	get        func(T0) O
	cell0      Cell
	cached0    T0
	v          O
}

func Computed1[T0, O comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	get ReadonlySignal1Func[T0, O],
) *ReadonlySignal1[T0, O] {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &ReadonlySignal1[T0, O]{
		isDirty: true,
		sys:     sys,
		get:     get,
		cell0:   arg0,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	s.v = s.get(
		s.cached0,
	)
	return s
}

func (s *ReadonlySignal1[T0, O]) Value() O {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal1[T0, O]) value() any {
	if !s.isDirty {
		return s.v
	}
	s.isDirty = false

	allArgsMatch := true
	depValue0 := s.cell0.value().(T0)
	if s.cached0 != depValue0 {
		allArgsMatch = false
	}

	if !allArgsMatch {
		currentValue := s.get(
			depValue0,
		)
		if s.v == currentValue {
			return s.v
		}
		s.v = currentValue
	}

	return s.v
}

func (s *ReadonlySignal1[T0, O]) markDirty() {
	if s.isDirty || s.lastUpdate == s.sys.currentUpdate {
		return
	}
	s.lastUpdate = s.sys.currentUpdate
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal1[T0, O]) addSubs(cells ...Cell) {
	s.subs = append(s.subs, cells...)
}

func (s *ReadonlySignal1[T0, O]) removeSub(cells Cell) {
	for i, sub := range s.subs {
		if sub == cells {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect1[T0 comparable] struct {
	system     *ReactiveSystem
	lastUpdate uint32
	fn         func(T0) error
	cell0      Cell
	cached0    T0
}

func Effect1[T0 comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	fn func(T0) error,
) (stop func()) {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &SideEffect1[T0]{
		system:     sys,
		fn:         fn,
		lastUpdate: 0,
		cell0:      arg0,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	s.fn(
		s.cached0,
	)

	return func() {
		arg0.removeSub(s)
	}
}

func (s *SideEffect1[T0]) value() any {
	current0, ok := s.cell0.value().(T0)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached0 == current0 {
		return nil
	}

	s.cached0 = current0
	s.fn(
		s.cached0,
	)
	return nil
}

func (s *SideEffect1[T0]) markDirty() {
	if s.lastUpdate == s.system.currentUpdate {
		return
	}
	s.lastUpdate = s.system.currentUpdate
	s.value()
}

func (s *SideEffect1[T0]) addSubs(sub ...Cell) {
	panic("not allowed")
}

func (s *SideEffect1[T0]) removeSub(sub Cell) {
	s.cell0 = nil
}

type ReadonlySignal2Func[T0, T1, O comparable] func(T0, T1) O

type ReadonlySignal2Args[T0, T1 comparable] struct {
	Arg0 T0
	Arg1 T1
}

type ReadonlySignal2[T0, T1, O comparable] struct {
	sys        *ReactiveSystem
	subs       []Cell
	isDirty    bool
	lastUpdate uint32
	get        func(T0, T1) O
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	v          O
}

func Computed2[T0, T1, O comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	get ReadonlySignal2Func[T0, T1, O],
) *ReadonlySignal2[T0, T1, O] {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &ReadonlySignal2[T0, T1, O]{
		isDirty: true,
		sys:     sys,
		get:     get,
		cell0:   arg0,
		cell1:   arg1,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	s.v = s.get(
		s.cached0,
		s.cached1,
	)
	return s
}

func (s *ReadonlySignal2[T0, T1, O]) Value() O {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal2[T0, T1, O]) value() any {
	if !s.isDirty {
		return s.v
	}
	s.isDirty = false

	allArgsMatch := true
	depValue0 := s.cell0.value().(T0)
	if s.cached0 != depValue0 {
		allArgsMatch = false
	}
	depValue1 := s.cell1.value().(T1)
	if s.cached1 != depValue1 {
		allArgsMatch = false
	}

	if !allArgsMatch {
		currentValue := s.get(
			depValue0,
			depValue1,
		)
		if s.v == currentValue {
			return s.v
		}
		s.v = currentValue
	}

	return s.v
}

func (s *ReadonlySignal2[T0, T1, O]) markDirty() {
	if s.isDirty || s.lastUpdate == s.sys.currentUpdate {
		return
	}
	s.lastUpdate = s.sys.currentUpdate
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal2[T0, T1, O]) addSubs(cells ...Cell) {
	s.subs = append(s.subs, cells...)
}

func (s *ReadonlySignal2[T0, T1, O]) removeSub(cells Cell) {
	for i, sub := range s.subs {
		if sub == cells {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect2[T0, T1 comparable] struct {
	system     *ReactiveSystem
	lastUpdate uint32
	fn         func(T0, T1) error
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
}

func Effect2[T0, T1 comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	fn func(T0, T1) error,
) (stop func()) {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &SideEffect2[T0, T1]{
		system:     sys,
		fn:         fn,
		lastUpdate: 0,
		cell0:      arg0,
		cell1:      arg1,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	s.fn(
		s.cached0,
		s.cached1,
	)

	return func() {
		arg0.removeSub(s)
		arg1.removeSub(s)
	}
}

func (s *SideEffect2[T0, T1]) value() any {
	current0, ok := s.cell0.value().(T0)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached0 == current0 {
		return nil
	}

	s.cached0 = current0
	current1, ok := s.cell1.value().(T1)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached1 == current1 {
		return nil
	}

	s.cached1 = current1
	s.fn(
		s.cached0,
		s.cached1,
	)
	return nil
}

func (s *SideEffect2[T0, T1]) markDirty() {
	if s.lastUpdate == s.system.currentUpdate {
		return
	}
	s.lastUpdate = s.system.currentUpdate
	s.value()
}

func (s *SideEffect2[T0, T1]) addSubs(sub ...Cell) {
	panic("not allowed")
}

func (s *SideEffect2[T0, T1]) removeSub(sub Cell) {
	s.cell0 = nil
	s.cell1 = nil
}

type ReadonlySignal3Func[T0, T1, T2, O comparable] func(T0, T1, T2) O

type ReadonlySignal3Args[T0, T1, T2 comparable] struct {
	Arg0 T0
	Arg1 T1
	Arg2 T2
}

type ReadonlySignal3[T0, T1, T2, O comparable] struct {
	sys        *ReactiveSystem
	subs       []Cell
	isDirty    bool
	lastUpdate uint32
	get        func(T0, T1, T2) O
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	v          O
}

func Computed3[T0, T1, T2, O comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	get ReadonlySignal3Func[T0, T1, T2, O],
) *ReadonlySignal3[T0, T1, T2, O] {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &ReadonlySignal3[T0, T1, T2, O]{
		isDirty: true,
		sys:     sys,
		get:     get,
		cell0:   arg0,
		cell1:   arg1,
		cell2:   arg2,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	s.v = s.get(
		s.cached0,
		s.cached1,
		s.cached2,
	)
	return s
}

func (s *ReadonlySignal3[T0, T1, T2, O]) Value() O {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal3[T0, T1, T2, O]) value() any {
	if !s.isDirty {
		return s.v
	}
	s.isDirty = false

	allArgsMatch := true
	depValue0 := s.cell0.value().(T0)
	if s.cached0 != depValue0 {
		allArgsMatch = false
	}
	depValue1 := s.cell1.value().(T1)
	if s.cached1 != depValue1 {
		allArgsMatch = false
	}
	depValue2 := s.cell2.value().(T2)
	if s.cached2 != depValue2 {
		allArgsMatch = false
	}

	if !allArgsMatch {
		currentValue := s.get(
			depValue0,
			depValue1,
			depValue2,
		)
		if s.v == currentValue {
			return s.v
		}
		s.v = currentValue
	}

	return s.v
}

func (s *ReadonlySignal3[T0, T1, T2, O]) markDirty() {
	if s.isDirty || s.lastUpdate == s.sys.currentUpdate {
		return
	}
	s.lastUpdate = s.sys.currentUpdate
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal3[T0, T1, T2, O]) addSubs(cells ...Cell) {
	s.subs = append(s.subs, cells...)
}

func (s *ReadonlySignal3[T0, T1, T2, O]) removeSub(cells Cell) {
	for i, sub := range s.subs {
		if sub == cells {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect3[T0, T1, T2 comparable] struct {
	system     *ReactiveSystem
	lastUpdate uint32
	fn         func(T0, T1, T2) error
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
}

func Effect3[T0, T1, T2 comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	fn func(T0, T1, T2) error,
) (stop func()) {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &SideEffect3[T0, T1, T2]{
		system:     sys,
		fn:         fn,
		lastUpdate: 0,
		cell0:      arg0,
		cell1:      arg1,
		cell2:      arg2,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
	)

	return func() {
		arg0.removeSub(s)
		arg1.removeSub(s)
		arg2.removeSub(s)
	}
}

func (s *SideEffect3[T0, T1, T2]) value() any {
	current0, ok := s.cell0.value().(T0)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached0 == current0 {
		return nil
	}

	s.cached0 = current0
	current1, ok := s.cell1.value().(T1)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached1 == current1 {
		return nil
	}

	s.cached1 = current1
	current2, ok := s.cell2.value().(T2)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached2 == current2 {
		return nil
	}

	s.cached2 = current2
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
	)
	return nil
}

func (s *SideEffect3[T0, T1, T2]) markDirty() {
	if s.lastUpdate == s.system.currentUpdate {
		return
	}
	s.lastUpdate = s.system.currentUpdate
	s.value()
}

func (s *SideEffect3[T0, T1, T2]) addSubs(sub ...Cell) {
	panic("not allowed")
}

func (s *SideEffect3[T0, T1, T2]) removeSub(sub Cell) {
	s.cell0 = nil
	s.cell1 = nil
	s.cell2 = nil
}

type ReadonlySignal4Func[T0, T1, T2, T3, O comparable] func(T0, T1, T2, T3) O

type ReadonlySignal4Args[T0, T1, T2, T3 comparable] struct {
	Arg0 T0
	Arg1 T1
	Arg2 T2
	Arg3 T3
}

type ReadonlySignal4[T0, T1, T2, T3, O comparable] struct {
	sys        *ReactiveSystem
	subs       []Cell
	isDirty    bool
	lastUpdate uint32
	get        func(T0, T1, T2, T3) O
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
	v          O
}

func Computed4[T0, T1, T2, T3, O comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	get ReadonlySignal4Func[T0, T1, T2, T3, O],
) *ReadonlySignal4[T0, T1, T2, T3, O] {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &ReadonlySignal4[T0, T1, T2, T3, O]{
		isDirty: true,
		sys:     sys,
		get:     get,
		cell0:   arg0,
		cell1:   arg1,
		cell2:   arg2,
		cell3:   arg3,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	s.v = s.get(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
	)
	return s
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) Value() O {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) value() any {
	if !s.isDirty {
		return s.v
	}
	s.isDirty = false

	allArgsMatch := true
	depValue0 := s.cell0.value().(T0)
	if s.cached0 != depValue0 {
		allArgsMatch = false
	}
	depValue1 := s.cell1.value().(T1)
	if s.cached1 != depValue1 {
		allArgsMatch = false
	}
	depValue2 := s.cell2.value().(T2)
	if s.cached2 != depValue2 {
		allArgsMatch = false
	}
	depValue3 := s.cell3.value().(T3)
	if s.cached3 != depValue3 {
		allArgsMatch = false
	}

	if !allArgsMatch {
		currentValue := s.get(
			depValue0,
			depValue1,
			depValue2,
			depValue3,
		)
		if s.v == currentValue {
			return s.v
		}
		s.v = currentValue
	}

	return s.v
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) markDirty() {
	if s.isDirty || s.lastUpdate == s.sys.currentUpdate {
		return
	}
	s.lastUpdate = s.sys.currentUpdate
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) addSubs(cells ...Cell) {
	s.subs = append(s.subs, cells...)
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) removeSub(cells Cell) {
	for i, sub := range s.subs {
		if sub == cells {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect4[T0, T1, T2, T3 comparable] struct {
	system     *ReactiveSystem
	lastUpdate uint32
	fn         func(T0, T1, T2, T3) error
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
}

func Effect4[T0, T1, T2, T3 comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	fn func(T0, T1, T2, T3) error,
) (stop func()) {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &SideEffect4[T0, T1, T2, T3]{
		system:     sys,
		fn:         fn,
		lastUpdate: 0,
		cell0:      arg0,
		cell1:      arg1,
		cell2:      arg2,
		cell3:      arg3,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
	)

	return func() {
		arg0.removeSub(s)
		arg1.removeSub(s)
		arg2.removeSub(s)
		arg3.removeSub(s)
	}
}

func (s *SideEffect4[T0, T1, T2, T3]) value() any {
	current0, ok := s.cell0.value().(T0)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached0 == current0 {
		return nil
	}

	s.cached0 = current0
	current1, ok := s.cell1.value().(T1)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached1 == current1 {
		return nil
	}

	s.cached1 = current1
	current2, ok := s.cell2.value().(T2)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached2 == current2 {
		return nil
	}

	s.cached2 = current2
	current3, ok := s.cell3.value().(T3)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached3 == current3 {
		return nil
	}

	s.cached3 = current3
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
	)
	return nil
}

func (s *SideEffect4[T0, T1, T2, T3]) markDirty() {
	if s.lastUpdate == s.system.currentUpdate {
		return
	}
	s.lastUpdate = s.system.currentUpdate
	s.value()
}

func (s *SideEffect4[T0, T1, T2, T3]) addSubs(sub ...Cell) {
	panic("not allowed")
}

func (s *SideEffect4[T0, T1, T2, T3]) removeSub(sub Cell) {
	s.cell0 = nil
	s.cell1 = nil
	s.cell2 = nil
	s.cell3 = nil
}

type ReadonlySignal5Func[T0, T1, T2, T3, T4, O comparable] func(T0, T1, T2, T3, T4) O

type ReadonlySignal5Args[T0, T1, T2, T3, T4 comparable] struct {
	Arg0 T0
	Arg1 T1
	Arg2 T2
	Arg3 T3
	Arg4 T4
}

type ReadonlySignal5[T0, T1, T2, T3, T4, O comparable] struct {
	sys        *ReactiveSystem
	subs       []Cell
	isDirty    bool
	lastUpdate uint32
	get        func(T0, T1, T2, T3, T4) O
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
	cell4      Cell
	cached4    T4
	v          O
}

func Computed5[T0, T1, T2, T3, T4, O comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	arg4 Cell,
	get ReadonlySignal5Func[T0, T1, T2, T3, T4, O],
) *ReadonlySignal5[T0, T1, T2, T3, T4, O] {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &ReadonlySignal5[T0, T1, T2, T3, T4, O]{
		isDirty: true,
		sys:     sys,
		get:     get,
		cell0:   arg0,
		cell1:   arg1,
		cell2:   arg2,
		cell3:   arg3,
		cell4:   arg4,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	arg4.addSubs(s)
	s.cached4 = arg4.value().(T4)
	s.v = s.get(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
	)
	return s
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) Value() O {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) value() any {
	if !s.isDirty {
		return s.v
	}
	s.isDirty = false

	allArgsMatch := true
	depValue0 := s.cell0.value().(T0)
	if s.cached0 != depValue0 {
		allArgsMatch = false
	}
	depValue1 := s.cell1.value().(T1)
	if s.cached1 != depValue1 {
		allArgsMatch = false
	}
	depValue2 := s.cell2.value().(T2)
	if s.cached2 != depValue2 {
		allArgsMatch = false
	}
	depValue3 := s.cell3.value().(T3)
	if s.cached3 != depValue3 {
		allArgsMatch = false
	}
	depValue4 := s.cell4.value().(T4)
	if s.cached4 != depValue4 {
		allArgsMatch = false
	}

	if !allArgsMatch {
		currentValue := s.get(
			depValue0,
			depValue1,
			depValue2,
			depValue3,
			depValue4,
		)
		if s.v == currentValue {
			return s.v
		}
		s.v = currentValue
	}

	return s.v
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) markDirty() {
	if s.isDirty || s.lastUpdate == s.sys.currentUpdate {
		return
	}
	s.lastUpdate = s.sys.currentUpdate
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) addSubs(cells ...Cell) {
	s.subs = append(s.subs, cells...)
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) removeSub(cells Cell) {
	for i, sub := range s.subs {
		if sub == cells {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect5[T0, T1, T2, T3, T4 comparable] struct {
	system     *ReactiveSystem
	lastUpdate uint32
	fn         func(T0, T1, T2, T3, T4) error
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
	cell4      Cell
	cached4    T4
}

func Effect5[T0, T1, T2, T3, T4 comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	arg4 Cell,
	fn func(T0, T1, T2, T3, T4) error,
) (stop func()) {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &SideEffect5[T0, T1, T2, T3, T4]{
		system:     sys,
		fn:         fn,
		lastUpdate: 0,
		cell0:      arg0,
		cell1:      arg1,
		cell2:      arg2,
		cell3:      arg3,
		cell4:      arg4,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	arg4.addSubs(s)
	s.cached4 = arg4.value().(T4)
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
	)

	return func() {
		arg0.removeSub(s)
		arg1.removeSub(s)
		arg2.removeSub(s)
		arg3.removeSub(s)
		arg4.removeSub(s)
	}
}

func (s *SideEffect5[T0, T1, T2, T3, T4]) value() any {
	current0, ok := s.cell0.value().(T0)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached0 == current0 {
		return nil
	}

	s.cached0 = current0
	current1, ok := s.cell1.value().(T1)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached1 == current1 {
		return nil
	}

	s.cached1 = current1
	current2, ok := s.cell2.value().(T2)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached2 == current2 {
		return nil
	}

	s.cached2 = current2
	current3, ok := s.cell3.value().(T3)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached3 == current3 {
		return nil
	}

	s.cached3 = current3
	current4, ok := s.cell4.value().(T4)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached4 == current4 {
		return nil
	}

	s.cached4 = current4
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
	)
	return nil
}

func (s *SideEffect5[T0, T1, T2, T3, T4]) markDirty() {
	if s.lastUpdate == s.system.currentUpdate {
		return
	}
	s.lastUpdate = s.system.currentUpdate
	s.value()
}

func (s *SideEffect5[T0, T1, T2, T3, T4]) addSubs(sub ...Cell) {
	panic("not allowed")
}

func (s *SideEffect5[T0, T1, T2, T3, T4]) removeSub(sub Cell) {
	s.cell0 = nil
	s.cell1 = nil
	s.cell2 = nil
	s.cell3 = nil
	s.cell4 = nil
}

type ReadonlySignal6Func[T0, T1, T2, T3, T4, T5, O comparable] func(T0, T1, T2, T3, T4, T5) O

type ReadonlySignal6Args[T0, T1, T2, T3, T4, T5 comparable] struct {
	Arg0 T0
	Arg1 T1
	Arg2 T2
	Arg3 T3
	Arg4 T4
	Arg5 T5
}

type ReadonlySignal6[T0, T1, T2, T3, T4, T5, O comparable] struct {
	sys        *ReactiveSystem
	subs       []Cell
	isDirty    bool
	lastUpdate uint32
	get        func(T0, T1, T2, T3, T4, T5) O
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
	cell4      Cell
	cached4    T4
	cell5      Cell
	cached5    T5
	v          O
}

func Computed6[T0, T1, T2, T3, T4, T5, O comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	arg4 Cell,
	arg5 Cell,
	get ReadonlySignal6Func[T0, T1, T2, T3, T4, T5, O],
) *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O] {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]{
		isDirty: true,
		sys:     sys,
		get:     get,
		cell0:   arg0,
		cell1:   arg1,
		cell2:   arg2,
		cell3:   arg3,
		cell4:   arg4,
		cell5:   arg5,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	arg4.addSubs(s)
	s.cached4 = arg4.value().(T4)
	arg5.addSubs(s)
	s.cached5 = arg5.value().(T5)
	s.v = s.get(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
	)
	return s
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) Value() O {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) value() any {
	if !s.isDirty {
		return s.v
	}
	s.isDirty = false

	allArgsMatch := true
	depValue0 := s.cell0.value().(T0)
	if s.cached0 != depValue0 {
		allArgsMatch = false
	}
	depValue1 := s.cell1.value().(T1)
	if s.cached1 != depValue1 {
		allArgsMatch = false
	}
	depValue2 := s.cell2.value().(T2)
	if s.cached2 != depValue2 {
		allArgsMatch = false
	}
	depValue3 := s.cell3.value().(T3)
	if s.cached3 != depValue3 {
		allArgsMatch = false
	}
	depValue4 := s.cell4.value().(T4)
	if s.cached4 != depValue4 {
		allArgsMatch = false
	}
	depValue5 := s.cell5.value().(T5)
	if s.cached5 != depValue5 {
		allArgsMatch = false
	}

	if !allArgsMatch {
		currentValue := s.get(
			depValue0,
			depValue1,
			depValue2,
			depValue3,
			depValue4,
			depValue5,
		)
		if s.v == currentValue {
			return s.v
		}
		s.v = currentValue
	}

	return s.v
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) markDirty() {
	if s.isDirty || s.lastUpdate == s.sys.currentUpdate {
		return
	}
	s.lastUpdate = s.sys.currentUpdate
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) addSubs(cells ...Cell) {
	s.subs = append(s.subs, cells...)
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) removeSub(cells Cell) {
	for i, sub := range s.subs {
		if sub == cells {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect6[T0, T1, T2, T3, T4, T5 comparable] struct {
	system     *ReactiveSystem
	lastUpdate uint32
	fn         func(T0, T1, T2, T3, T4, T5) error
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
	cell4      Cell
	cached4    T4
	cell5      Cell
	cached5    T5
}

func Effect6[T0, T1, T2, T3, T4, T5 comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	arg4 Cell,
	arg5 Cell,
	fn func(T0, T1, T2, T3, T4, T5) error,
) (stop func()) {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &SideEffect6[T0, T1, T2, T3, T4, T5]{
		system:     sys,
		fn:         fn,
		lastUpdate: 0,
		cell0:      arg0,
		cell1:      arg1,
		cell2:      arg2,
		cell3:      arg3,
		cell4:      arg4,
		cell5:      arg5,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	arg4.addSubs(s)
	s.cached4 = arg4.value().(T4)
	arg5.addSubs(s)
	s.cached5 = arg5.value().(T5)
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
	)

	return func() {
		arg0.removeSub(s)
		arg1.removeSub(s)
		arg2.removeSub(s)
		arg3.removeSub(s)
		arg4.removeSub(s)
		arg5.removeSub(s)
	}
}

func (s *SideEffect6[T0, T1, T2, T3, T4, T5]) value() any {
	current0, ok := s.cell0.value().(T0)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached0 == current0 {
		return nil
	}

	s.cached0 = current0
	current1, ok := s.cell1.value().(T1)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached1 == current1 {
		return nil
	}

	s.cached1 = current1
	current2, ok := s.cell2.value().(T2)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached2 == current2 {
		return nil
	}

	s.cached2 = current2
	current3, ok := s.cell3.value().(T3)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached3 == current3 {
		return nil
	}

	s.cached3 = current3
	current4, ok := s.cell4.value().(T4)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached4 == current4 {
		return nil
	}

	s.cached4 = current4
	current5, ok := s.cell5.value().(T5)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached5 == current5 {
		return nil
	}

	s.cached5 = current5
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
	)
	return nil
}

func (s *SideEffect6[T0, T1, T2, T3, T4, T5]) markDirty() {
	if s.lastUpdate == s.system.currentUpdate {
		return
	}
	s.lastUpdate = s.system.currentUpdate
	s.value()
}

func (s *SideEffect6[T0, T1, T2, T3, T4, T5]) addSubs(sub ...Cell) {
	panic("not allowed")
}

func (s *SideEffect6[T0, T1, T2, T3, T4, T5]) removeSub(sub Cell) {
	s.cell0 = nil
	s.cell1 = nil
	s.cell2 = nil
	s.cell3 = nil
	s.cell4 = nil
	s.cell5 = nil
}

type ReadonlySignal7Func[T0, T1, T2, T3, T4, T5, T6, O comparable] func(T0, T1, T2, T3, T4, T5, T6) O

type ReadonlySignal7Args[T0, T1, T2, T3, T4, T5, T6 comparable] struct {
	Arg0 T0
	Arg1 T1
	Arg2 T2
	Arg3 T3
	Arg4 T4
	Arg5 T5
	Arg6 T6
}

type ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O comparable] struct {
	sys        *ReactiveSystem
	subs       []Cell
	isDirty    bool
	lastUpdate uint32
	get        func(T0, T1, T2, T3, T4, T5, T6) O
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
	cell4      Cell
	cached4    T4
	cell5      Cell
	cached5    T5
	cell6      Cell
	cached6    T6
	v          O
}

func Computed7[T0, T1, T2, T3, T4, T5, T6, O comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	arg4 Cell,
	arg5 Cell,
	arg6 Cell,
	get ReadonlySignal7Func[T0, T1, T2, T3, T4, T5, T6, O],
) *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O] {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]{
		isDirty: true,
		sys:     sys,
		get:     get,
		cell0:   arg0,
		cell1:   arg1,
		cell2:   arg2,
		cell3:   arg3,
		cell4:   arg4,
		cell5:   arg5,
		cell6:   arg6,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	arg4.addSubs(s)
	s.cached4 = arg4.value().(T4)
	arg5.addSubs(s)
	s.cached5 = arg5.value().(T5)
	arg6.addSubs(s)
	s.cached6 = arg6.value().(T6)
	s.v = s.get(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
		s.cached6,
	)
	return s
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) Value() O {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) value() any {
	if !s.isDirty {
		return s.v
	}
	s.isDirty = false

	allArgsMatch := true
	depValue0 := s.cell0.value().(T0)
	if s.cached0 != depValue0 {
		allArgsMatch = false
	}
	depValue1 := s.cell1.value().(T1)
	if s.cached1 != depValue1 {
		allArgsMatch = false
	}
	depValue2 := s.cell2.value().(T2)
	if s.cached2 != depValue2 {
		allArgsMatch = false
	}
	depValue3 := s.cell3.value().(T3)
	if s.cached3 != depValue3 {
		allArgsMatch = false
	}
	depValue4 := s.cell4.value().(T4)
	if s.cached4 != depValue4 {
		allArgsMatch = false
	}
	depValue5 := s.cell5.value().(T5)
	if s.cached5 != depValue5 {
		allArgsMatch = false
	}
	depValue6 := s.cell6.value().(T6)
	if s.cached6 != depValue6 {
		allArgsMatch = false
	}

	if !allArgsMatch {
		currentValue := s.get(
			depValue0,
			depValue1,
			depValue2,
			depValue3,
			depValue4,
			depValue5,
			depValue6,
		)
		if s.v == currentValue {
			return s.v
		}
		s.v = currentValue
	}

	return s.v
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) markDirty() {
	if s.isDirty || s.lastUpdate == s.sys.currentUpdate {
		return
	}
	s.lastUpdate = s.sys.currentUpdate
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) addSubs(cells ...Cell) {
	s.subs = append(s.subs, cells...)
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) removeSub(cells Cell) {
	for i, sub := range s.subs {
		if sub == cells {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect7[T0, T1, T2, T3, T4, T5, T6 comparable] struct {
	system     *ReactiveSystem
	lastUpdate uint32
	fn         func(T0, T1, T2, T3, T4, T5, T6) error
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
	cell4      Cell
	cached4    T4
	cell5      Cell
	cached5    T5
	cell6      Cell
	cached6    T6
}

func Effect7[T0, T1, T2, T3, T4, T5, T6 comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	arg4 Cell,
	arg5 Cell,
	arg6 Cell,
	fn func(T0, T1, T2, T3, T4, T5, T6) error,
) (stop func()) {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &SideEffect7[T0, T1, T2, T3, T4, T5, T6]{
		system:     sys,
		fn:         fn,
		lastUpdate: 0,
		cell0:      arg0,
		cell1:      arg1,
		cell2:      arg2,
		cell3:      arg3,
		cell4:      arg4,
		cell5:      arg5,
		cell6:      arg6,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	arg4.addSubs(s)
	s.cached4 = arg4.value().(T4)
	arg5.addSubs(s)
	s.cached5 = arg5.value().(T5)
	arg6.addSubs(s)
	s.cached6 = arg6.value().(T6)
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
		s.cached6,
	)

	return func() {
		arg0.removeSub(s)
		arg1.removeSub(s)
		arg2.removeSub(s)
		arg3.removeSub(s)
		arg4.removeSub(s)
		arg5.removeSub(s)
		arg6.removeSub(s)
	}
}

func (s *SideEffect7[T0, T1, T2, T3, T4, T5, T6]) value() any {
	current0, ok := s.cell0.value().(T0)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached0 == current0 {
		return nil
	}

	s.cached0 = current0
	current1, ok := s.cell1.value().(T1)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached1 == current1 {
		return nil
	}

	s.cached1 = current1
	current2, ok := s.cell2.value().(T2)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached2 == current2 {
		return nil
	}

	s.cached2 = current2
	current3, ok := s.cell3.value().(T3)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached3 == current3 {
		return nil
	}

	s.cached3 = current3
	current4, ok := s.cell4.value().(T4)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached4 == current4 {
		return nil
	}

	s.cached4 = current4
	current5, ok := s.cell5.value().(T5)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached5 == current5 {
		return nil
	}

	s.cached5 = current5
	current6, ok := s.cell6.value().(T6)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached6 == current6 {
		return nil
	}

	s.cached6 = current6
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
		s.cached6,
	)
	return nil
}

func (s *SideEffect7[T0, T1, T2, T3, T4, T5, T6]) markDirty() {
	if s.lastUpdate == s.system.currentUpdate {
		return
	}
	s.lastUpdate = s.system.currentUpdate
	s.value()
}

func (s *SideEffect7[T0, T1, T2, T3, T4, T5, T6]) addSubs(sub ...Cell) {
	panic("not allowed")
}

func (s *SideEffect7[T0, T1, T2, T3, T4, T5, T6]) removeSub(sub Cell) {
	s.cell0 = nil
	s.cell1 = nil
	s.cell2 = nil
	s.cell3 = nil
	s.cell4 = nil
	s.cell5 = nil
	s.cell6 = nil
}

type ReadonlySignal8Func[T0, T1, T2, T3, T4, T5, T6, T7, O comparable] func(T0, T1, T2, T3, T4, T5, T6, T7) O

type ReadonlySignal8Args[T0, T1, T2, T3, T4, T5, T6, T7 comparable] struct {
	Arg0 T0
	Arg1 T1
	Arg2 T2
	Arg3 T3
	Arg4 T4
	Arg5 T5
	Arg6 T6
	Arg7 T7
}

type ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O comparable] struct {
	sys        *ReactiveSystem
	subs       []Cell
	isDirty    bool
	lastUpdate uint32
	get        func(T0, T1, T2, T3, T4, T5, T6, T7) O
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
	cell4      Cell
	cached4    T4
	cell5      Cell
	cached5    T5
	cell6      Cell
	cached6    T6
	cell7      Cell
	cached7    T7
	v          O
}

func Computed8[T0, T1, T2, T3, T4, T5, T6, T7, O comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	arg4 Cell,
	arg5 Cell,
	arg6 Cell,
	arg7 Cell,
	get ReadonlySignal8Func[T0, T1, T2, T3, T4, T5, T6, T7, O],
) *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O] {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]{
		isDirty: true,
		sys:     sys,
		get:     get,
		cell0:   arg0,
		cell1:   arg1,
		cell2:   arg2,
		cell3:   arg3,
		cell4:   arg4,
		cell5:   arg5,
		cell6:   arg6,
		cell7:   arg7,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	arg4.addSubs(s)
	s.cached4 = arg4.value().(T4)
	arg5.addSubs(s)
	s.cached5 = arg5.value().(T5)
	arg6.addSubs(s)
	s.cached6 = arg6.value().(T6)
	arg7.addSubs(s)
	s.cached7 = arg7.value().(T7)
	s.v = s.get(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
		s.cached6,
		s.cached7,
	)
	return s
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) Value() O {
	s.sys.mu.Lock()
	defer s.sys.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) value() any {
	if !s.isDirty {
		return s.v
	}
	s.isDirty = false

	allArgsMatch := true
	depValue0 := s.cell0.value().(T0)
	if s.cached0 != depValue0 {
		allArgsMatch = false
	}
	depValue1 := s.cell1.value().(T1)
	if s.cached1 != depValue1 {
		allArgsMatch = false
	}
	depValue2 := s.cell2.value().(T2)
	if s.cached2 != depValue2 {
		allArgsMatch = false
	}
	depValue3 := s.cell3.value().(T3)
	if s.cached3 != depValue3 {
		allArgsMatch = false
	}
	depValue4 := s.cell4.value().(T4)
	if s.cached4 != depValue4 {
		allArgsMatch = false
	}
	depValue5 := s.cell5.value().(T5)
	if s.cached5 != depValue5 {
		allArgsMatch = false
	}
	depValue6 := s.cell6.value().(T6)
	if s.cached6 != depValue6 {
		allArgsMatch = false
	}
	depValue7 := s.cell7.value().(T7)
	if s.cached7 != depValue7 {
		allArgsMatch = false
	}

	if !allArgsMatch {
		currentValue := s.get(
			depValue0,
			depValue1,
			depValue2,
			depValue3,
			depValue4,
			depValue5,
			depValue6,
			depValue7,
		)
		if s.v == currentValue {
			return s.v
		}
		s.v = currentValue
	}

	return s.v
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) markDirty() {
	if s.isDirty || s.lastUpdate == s.sys.currentUpdate {
		return
	}
	s.lastUpdate = s.sys.currentUpdate
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) addSubs(cells ...Cell) {
	s.subs = append(s.subs, cells...)
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) removeSub(cells Cell) {
	for i, sub := range s.subs {
		if sub == cells {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7 comparable] struct {
	system     *ReactiveSystem
	lastUpdate uint32
	fn         func(T0, T1, T2, T3, T4, T5, T6, T7) error
	cell0      Cell
	cached0    T0
	cell1      Cell
	cached1    T1
	cell2      Cell
	cached2    T2
	cell3      Cell
	cached3    T3
	cell4      Cell
	cached4    T4
	cell5      Cell
	cached5    T5
	cell6      Cell
	cached6    T6
	cell7      Cell
	cached7    T7
}

func Effect8[T0, T1, T2, T3, T4, T5, T6, T7 comparable](
	sys *ReactiveSystem,
	arg0 Cell,
	arg1 Cell,
	arg2 Cell,
	arg3 Cell,
	arg4 Cell,
	arg5 Cell,
	arg6 Cell,
	arg7 Cell,
	fn func(T0, T1, T2, T3, T4, T5, T6, T7) error,
) (stop func()) {
	sys.mu.Lock()
	defer sys.mu.Unlock()

	s := &SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]{
		system:     sys,
		fn:         fn,
		lastUpdate: 0,
		cell0:      arg0,
		cell1:      arg1,
		cell2:      arg2,
		cell3:      arg3,
		cell4:      arg4,
		cell5:      arg5,
		cell6:      arg6,
		cell7:      arg7,
	}
	arg0.addSubs(s)
	s.cached0 = arg0.value().(T0)
	arg1.addSubs(s)
	s.cached1 = arg1.value().(T1)
	arg2.addSubs(s)
	s.cached2 = arg2.value().(T2)
	arg3.addSubs(s)
	s.cached3 = arg3.value().(T3)
	arg4.addSubs(s)
	s.cached4 = arg4.value().(T4)
	arg5.addSubs(s)
	s.cached5 = arg5.value().(T5)
	arg6.addSubs(s)
	s.cached6 = arg6.value().(T6)
	arg7.addSubs(s)
	s.cached7 = arg7.value().(T7)
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
		s.cached6,
		s.cached7,
	)

	return func() {
		arg0.removeSub(s)
		arg1.removeSub(s)
		arg2.removeSub(s)
		arg3.removeSub(s)
		arg4.removeSub(s)
		arg5.removeSub(s)
		arg6.removeSub(s)
		arg7.removeSub(s)
	}
}

func (s *SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]) value() any {
	current0, ok := s.cell0.value().(T0)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached0 == current0 {
		return nil
	}

	s.cached0 = current0
	current1, ok := s.cell1.value().(T1)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached1 == current1 {
		return nil
	}

	s.cached1 = current1
	current2, ok := s.cell2.value().(T2)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached2 == current2 {
		return nil
	}

	s.cached2 = current2
	current3, ok := s.cell3.value().(T3)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached3 == current3 {
		return nil
	}

	s.cached3 = current3
	current4, ok := s.cell4.value().(T4)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached4 == current4 {
		return nil
	}

	s.cached4 = current4
	current5, ok := s.cell5.value().(T5)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached5 == current5 {
		return nil
	}

	s.cached5 = current5
	current6, ok := s.cell6.value().(T6)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached6 == current6 {
		return nil
	}

	s.cached6 = current6
	current7, ok := s.cell7.value().(T7)
	if !ok {
		panic("type assertion failed")
	}
	if s.cached7 == current7 {
		return nil
	}

	s.cached7 = current7
	s.fn(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
		s.cached6,
		s.cached7,
	)
	return nil
}

func (s *SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]) markDirty() {
	if s.lastUpdate == s.system.currentUpdate {
		return
	}
	s.lastUpdate = s.system.currentUpdate
	s.value()
}

func (s *SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]) addSubs(sub ...Cell) {
	panic("not allowed")
}

func (s *SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]) removeSub(sub Cell) {
	s.cell0 = nil
	s.cell1 = nil
	s.cell2 = nil
	s.cell3 = nil
	s.cell4 = nil
	s.cell5 = nil
	s.cell6 = nil
	s.cell7 = nil
}
