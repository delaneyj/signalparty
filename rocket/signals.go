package rocket

import "sync"

type ReactiveSystem struct {
	mu sync.RWMutex
}

func NewReactiveSystem() *ReactiveSystem {
	return &ReactiveSystem{
		mu: sync.RWMutex{},
	}
}

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
	rs   *ReactiveSystem
	val  T
	subs []Subscriber
	ver  uint32
}

func (s *WriteableSignal[T]) Value() T {
	s.rs.mu.RLock()
	defer s.rs.mu.RUnlock()
	return s.val
}

func (s *WriteableSignal[T]) value() any {
	return s.val
}

func (s *WriteableSignal[T]) version() uint32 {
	return s.ver
}

func (s *WriteableSignal[T]) SetValue(value T) {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()

	if s.val == value {
		return
	}
	s.val = value

	if len(s.subs) == 0 {
		return
	}
	s.ver++
	s.markDirty()
}

func Signal[T comparable](rs *ReactiveSystem, value T) *WriteableSignal[T] {
	s := &WriteableSignal[T]{rs: rs, val: value, ver: 1}
	return s
}

func (s *WriteableSignal[T]) markDirty() {
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *WriteableSignal[T]) addSubs(sub ...Subscriber) {
	s.subs = append(s.subs, sub...)
}

func (s *WriteableSignal[T]) removeSub(toRemove Subscriber) {
	for i, sub := range s.subs {
		if sub == toRemove {
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
	rs         *ReactiveSystem
	subs       []Subscriber
	isDirty    bool
	ver        uint32 // value of readoly signals changes atomic increment
	versionSum uint32 // value of dependencies changes
	get        func(T0) O
	dep0       Dependency
	val        O
}

func Computed1[T0, O comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	get ReadonlySignal1Func[T0, O],
) *ReadonlySignal1[T0, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal1[T0, O]{
		isDirty:    true,
		get:        get,
		ver:        1,
		versionSum: 0,
		dep0:       dep0,
	}
	dep0.addSubs(s)

	return s
}

func (s *ReadonlySignal1[T0, O]) Value() O {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal1[T0, O]) value() any {
	if !s.isDirty {
		return s.val
	}
	s.isDirty = false

	var allArgsSum uint32
	depValue0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		depValue0 = zeroT0
	}
	depVersion0 := s.dep0.version()
	allArgsSum += depVersion0
	if allArgsSum == s.versionSum {
		return s.val
	}

	s.versionSum = allArgsSum
	currentValue := s.get(
		depValue0,
	)
	if s.val == currentValue {
		return s.val
	}
	s.val = currentValue
	s.ver++
	return s.val
}

func (s *ReadonlySignal1[T0, O]) version() uint32 {
	return s.ver
}

func (s *ReadonlySignal1[T0, O]) markDirty() {
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal1[T0, O]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *ReadonlySignal1[T0, O]) removeSub(toRemove Subscriber) {
	for i, sub := range s.subs {
		if sub == toRemove {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect1[T0 comparable] struct {
	rs         *ReactiveSystem
	versionSum uint32
	fn         func(T0) error
	dep0       Dependency
}

func Effect1[T0 comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	fn func(T0) error,
) (stop func()) {
	rs.mu.Lock()

	s := &SideEffect1[T0]{
		rs:         rs,
		fn:         fn,
		versionSum: 0,
		dep0:       dep0,
	}
	dep0.addSubs(s)

	defer rs.mu.Unlock()

	s.fn(
		dep0.value().(T0),
	)

	return func() {
		dep0.removeSub(s)
	}
}

func (s *SideEffect1[T0]) value() any {
	var allArgsSum uint32

	current0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		current0 = zeroT0
	}
	currentVersion0 := s.dep0.version()
	allArgsSum += currentVersion0

	if allArgsSum == s.versionSum {
		return nil
	}
	s.versionSum = allArgsSum

	s.fn(
		current0,
	)
	return nil
}

func (s *SideEffect1[T0]) markDirty() {
	s.value()
}

type ReadonlySignal2Func[T0, T1, O comparable] func(T0, T1) O

type ReadonlySignal2Args[T0, T1 comparable] struct {
	Arg0 T0
	Arg1 T1
}

type ReadonlySignal2[T0, T1, O comparable] struct {
	rs         *ReactiveSystem
	subs       []Subscriber
	isDirty    bool
	ver        uint32 // value of readoly signals changes atomic increment
	versionSum uint32 // value of dependencies changes
	get        func(T0, T1) O
	dep0       Dependency
	dep1       Dependency
	val        O
}

func Computed2[T0, T1, O comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	get ReadonlySignal2Func[T0, T1, O],
) *ReadonlySignal2[T0, T1, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal2[T0, T1, O]{
		isDirty:    true,
		get:        get,
		ver:        1,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)

	return s
}

func (s *ReadonlySignal2[T0, T1, O]) Value() O {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal2[T0, T1, O]) value() any {
	if !s.isDirty {
		return s.val
	}
	s.isDirty = false

	var allArgsSum uint32
	depValue0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		depValue0 = zeroT0
	}
	depVersion0 := s.dep0.version()
	allArgsSum += depVersion0
	depValue1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		depValue1 = zeroT1
	}
	depVersion1 := s.dep1.version()
	allArgsSum += depVersion1
	if allArgsSum == s.versionSum {
		return s.val
	}

	s.versionSum = allArgsSum
	currentValue := s.get(
		depValue0,
		depValue1,
	)
	if s.val == currentValue {
		return s.val
	}
	s.val = currentValue
	s.ver++
	return s.val
}

func (s *ReadonlySignal2[T0, T1, O]) version() uint32 {
	return s.ver
}

func (s *ReadonlySignal2[T0, T1, O]) markDirty() {
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal2[T0, T1, O]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *ReadonlySignal2[T0, T1, O]) removeSub(toRemove Subscriber) {
	for i, sub := range s.subs {
		if sub == toRemove {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect2[T0, T1 comparable] struct {
	rs         *ReactiveSystem
	versionSum uint32
	fn         func(T0, T1) error
	dep0       Dependency
	dep1       Dependency
}

func Effect2[T0, T1 comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	fn func(T0, T1) error,
) (stop func()) {
	rs.mu.Lock()

	s := &SideEffect2[T0, T1]{
		rs:         rs,
		fn:         fn,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)

	defer rs.mu.Unlock()

	s.fn(
		dep0.value().(T0),
		dep1.value().(T1),
	)

	return func() {
		dep0.removeSub(s)
		dep1.removeSub(s)
	}
}

func (s *SideEffect2[T0, T1]) value() any {
	var allArgsSum uint32

	current0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		current0 = zeroT0
	}
	currentVersion0 := s.dep0.version()
	allArgsSum += currentVersion0

	current1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		current1 = zeroT1
	}
	currentVersion1 := s.dep1.version()
	allArgsSum += currentVersion1

	if allArgsSum == s.versionSum {
		return nil
	}
	s.versionSum = allArgsSum

	s.fn(
		current0,
		current1,
	)
	return nil
}

func (s *SideEffect2[T0, T1]) markDirty() {
	s.value()
}

type ReadonlySignal3Func[T0, T1, T2, O comparable] func(T0, T1, T2) O

type ReadonlySignal3Args[T0, T1, T2 comparable] struct {
	Arg0 T0
	Arg1 T1
	Arg2 T2
}

type ReadonlySignal3[T0, T1, T2, O comparable] struct {
	rs         *ReactiveSystem
	subs       []Subscriber
	isDirty    bool
	ver        uint32 // value of readoly signals changes atomic increment
	versionSum uint32 // value of dependencies changes
	get        func(T0, T1, T2) O
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	val        O
}

func Computed3[T0, T1, T2, O comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	get ReadonlySignal3Func[T0, T1, T2, O],
) *ReadonlySignal3[T0, T1, T2, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal3[T0, T1, T2, O]{
		isDirty:    true,
		get:        get,
		ver:        1,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)

	return s
}

func (s *ReadonlySignal3[T0, T1, T2, O]) Value() O {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal3[T0, T1, T2, O]) value() any {
	if !s.isDirty {
		return s.val
	}
	s.isDirty = false

	var allArgsSum uint32
	depValue0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		depValue0 = zeroT0
	}
	depVersion0 := s.dep0.version()
	allArgsSum += depVersion0
	depValue1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		depValue1 = zeroT1
	}
	depVersion1 := s.dep1.version()
	allArgsSum += depVersion1
	depValue2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		depValue2 = zeroT2
	}
	depVersion2 := s.dep2.version()
	allArgsSum += depVersion2
	if allArgsSum == s.versionSum {
		return s.val
	}

	s.versionSum = allArgsSum
	currentValue := s.get(
		depValue0,
		depValue1,
		depValue2,
	)
	if s.val == currentValue {
		return s.val
	}
	s.val = currentValue
	s.ver++
	return s.val
}

func (s *ReadonlySignal3[T0, T1, T2, O]) version() uint32 {
	return s.ver
}

func (s *ReadonlySignal3[T0, T1, T2, O]) markDirty() {
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal3[T0, T1, T2, O]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *ReadonlySignal3[T0, T1, T2, O]) removeSub(toRemove Subscriber) {
	for i, sub := range s.subs {
		if sub == toRemove {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect3[T0, T1, T2 comparable] struct {
	rs         *ReactiveSystem
	versionSum uint32
	fn         func(T0, T1, T2) error
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
}

func Effect3[T0, T1, T2 comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	fn func(T0, T1, T2) error,
) (stop func()) {
	rs.mu.Lock()

	s := &SideEffect3[T0, T1, T2]{
		rs:         rs,
		fn:         fn,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)

	defer rs.mu.Unlock()

	s.fn(
		dep0.value().(T0),
		dep1.value().(T1),
		dep2.value().(T2),
	)

	return func() {
		dep0.removeSub(s)
		dep1.removeSub(s)
		dep2.removeSub(s)
	}
}

func (s *SideEffect3[T0, T1, T2]) value() any {
	var allArgsSum uint32

	current0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		current0 = zeroT0
	}
	currentVersion0 := s.dep0.version()
	allArgsSum += currentVersion0

	current1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		current1 = zeroT1
	}
	currentVersion1 := s.dep1.version()
	allArgsSum += currentVersion1

	current2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		current2 = zeroT2
	}
	currentVersion2 := s.dep2.version()
	allArgsSum += currentVersion2

	if allArgsSum == s.versionSum {
		return nil
	}
	s.versionSum = allArgsSum

	s.fn(
		current0,
		current1,
		current2,
	)
	return nil
}

func (s *SideEffect3[T0, T1, T2]) markDirty() {
	s.value()
}

type ReadonlySignal4Func[T0, T1, T2, T3, O comparable] func(T0, T1, T2, T3) O

type ReadonlySignal4Args[T0, T1, T2, T3 comparable] struct {
	Arg0 T0
	Arg1 T1
	Arg2 T2
	Arg3 T3
}

type ReadonlySignal4[T0, T1, T2, T3, O comparable] struct {
	rs         *ReactiveSystem
	subs       []Subscriber
	isDirty    bool
	ver        uint32 // value of readoly signals changes atomic increment
	versionSum uint32 // value of dependencies changes
	get        func(T0, T1, T2, T3) O
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
	val        O
}

func Computed4[T0, T1, T2, T3, O comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	get ReadonlySignal4Func[T0, T1, T2, T3, O],
) *ReadonlySignal4[T0, T1, T2, T3, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal4[T0, T1, T2, T3, O]{
		isDirty:    true,
		get:        get,
		ver:        1,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)

	return s
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) Value() O {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) value() any {
	if !s.isDirty {
		return s.val
	}
	s.isDirty = false

	var allArgsSum uint32
	depValue0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		depValue0 = zeroT0
	}
	depVersion0 := s.dep0.version()
	allArgsSum += depVersion0
	depValue1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		depValue1 = zeroT1
	}
	depVersion1 := s.dep1.version()
	allArgsSum += depVersion1
	depValue2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		depValue2 = zeroT2
	}
	depVersion2 := s.dep2.version()
	allArgsSum += depVersion2
	depValue3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		depValue3 = zeroT3
	}
	depVersion3 := s.dep3.version()
	allArgsSum += depVersion3
	if allArgsSum == s.versionSum {
		return s.val
	}

	s.versionSum = allArgsSum
	currentValue := s.get(
		depValue0,
		depValue1,
		depValue2,
		depValue3,
	)
	if s.val == currentValue {
		return s.val
	}
	s.val = currentValue
	s.ver++
	return s.val
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) version() uint32 {
	return s.ver
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) markDirty() {
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) removeSub(toRemove Subscriber) {
	for i, sub := range s.subs {
		if sub == toRemove {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect4[T0, T1, T2, T3 comparable] struct {
	rs         *ReactiveSystem
	versionSum uint32
	fn         func(T0, T1, T2, T3) error
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
}

func Effect4[T0, T1, T2, T3 comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	fn func(T0, T1, T2, T3) error,
) (stop func()) {
	rs.mu.Lock()

	s := &SideEffect4[T0, T1, T2, T3]{
		rs:         rs,
		fn:         fn,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)

	defer rs.mu.Unlock()

	s.fn(
		dep0.value().(T0),
		dep1.value().(T1),
		dep2.value().(T2),
		dep3.value().(T3),
	)

	return func() {
		dep0.removeSub(s)
		dep1.removeSub(s)
		dep2.removeSub(s)
		dep3.removeSub(s)
	}
}

func (s *SideEffect4[T0, T1, T2, T3]) value() any {
	var allArgsSum uint32

	current0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		current0 = zeroT0
	}
	currentVersion0 := s.dep0.version()
	allArgsSum += currentVersion0

	current1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		current1 = zeroT1
	}
	currentVersion1 := s.dep1.version()
	allArgsSum += currentVersion1

	current2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		current2 = zeroT2
	}
	currentVersion2 := s.dep2.version()
	allArgsSum += currentVersion2

	current3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		current3 = zeroT3
	}
	currentVersion3 := s.dep3.version()
	allArgsSum += currentVersion3

	if allArgsSum == s.versionSum {
		return nil
	}
	s.versionSum = allArgsSum

	s.fn(
		current0,
		current1,
		current2,
		current3,
	)
	return nil
}

func (s *SideEffect4[T0, T1, T2, T3]) markDirty() {
	s.value()
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
	rs         *ReactiveSystem
	subs       []Subscriber
	isDirty    bool
	ver        uint32 // value of readoly signals changes atomic increment
	versionSum uint32 // value of dependencies changes
	get        func(T0, T1, T2, T3, T4) O
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
	dep4       Dependency
	val        O
}

func Computed5[T0, T1, T2, T3, T4, O comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	dep4 Dependency,
	get ReadonlySignal5Func[T0, T1, T2, T3, T4, O],
) *ReadonlySignal5[T0, T1, T2, T3, T4, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal5[T0, T1, T2, T3, T4, O]{
		isDirty:    true,
		get:        get,
		ver:        1,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
		dep4:       dep4,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)
	dep4.addSubs(s)

	return s
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) Value() O {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) value() any {
	if !s.isDirty {
		return s.val
	}
	s.isDirty = false

	var allArgsSum uint32
	depValue0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		depValue0 = zeroT0
	}
	depVersion0 := s.dep0.version()
	allArgsSum += depVersion0
	depValue1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		depValue1 = zeroT1
	}
	depVersion1 := s.dep1.version()
	allArgsSum += depVersion1
	depValue2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		depValue2 = zeroT2
	}
	depVersion2 := s.dep2.version()
	allArgsSum += depVersion2
	depValue3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		depValue3 = zeroT3
	}
	depVersion3 := s.dep3.version()
	allArgsSum += depVersion3
	depValue4, ok4 := s.dep4.value().(T4)
	if !ok4 {
		var zeroT4 T4
		depValue4 = zeroT4
	}
	depVersion4 := s.dep4.version()
	allArgsSum += depVersion4
	if allArgsSum == s.versionSum {
		return s.val
	}

	s.versionSum = allArgsSum
	currentValue := s.get(
		depValue0,
		depValue1,
		depValue2,
		depValue3,
		depValue4,
	)
	if s.val == currentValue {
		return s.val
	}
	s.val = currentValue
	s.ver++
	return s.val
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) version() uint32 {
	return s.ver
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) markDirty() {
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) removeSub(toRemove Subscriber) {
	for i, sub := range s.subs {
		if sub == toRemove {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect5[T0, T1, T2, T3, T4 comparable] struct {
	rs         *ReactiveSystem
	versionSum uint32
	fn         func(T0, T1, T2, T3, T4) error
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
	dep4       Dependency
}

func Effect5[T0, T1, T2, T3, T4 comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	dep4 Dependency,
	fn func(T0, T1, T2, T3, T4) error,
) (stop func()) {
	rs.mu.Lock()

	s := &SideEffect5[T0, T1, T2, T3, T4]{
		rs:         rs,
		fn:         fn,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
		dep4:       dep4,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)
	dep4.addSubs(s)

	defer rs.mu.Unlock()

	s.fn(
		dep0.value().(T0),
		dep1.value().(T1),
		dep2.value().(T2),
		dep3.value().(T3),
		dep4.value().(T4),
	)

	return func() {
		dep0.removeSub(s)
		dep1.removeSub(s)
		dep2.removeSub(s)
		dep3.removeSub(s)
		dep4.removeSub(s)
	}
}

func (s *SideEffect5[T0, T1, T2, T3, T4]) value() any {
	var allArgsSum uint32

	current0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		current0 = zeroT0
	}
	currentVersion0 := s.dep0.version()
	allArgsSum += currentVersion0

	current1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		current1 = zeroT1
	}
	currentVersion1 := s.dep1.version()
	allArgsSum += currentVersion1

	current2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		current2 = zeroT2
	}
	currentVersion2 := s.dep2.version()
	allArgsSum += currentVersion2

	current3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		current3 = zeroT3
	}
	currentVersion3 := s.dep3.version()
	allArgsSum += currentVersion3

	current4, ok4 := s.dep4.value().(T4)
	if !ok4 {
		var zeroT4 T4
		current4 = zeroT4
	}
	currentVersion4 := s.dep4.version()
	allArgsSum += currentVersion4

	if allArgsSum == s.versionSum {
		return nil
	}
	s.versionSum = allArgsSum

	s.fn(
		current0,
		current1,
		current2,
		current3,
		current4,
	)
	return nil
}

func (s *SideEffect5[T0, T1, T2, T3, T4]) markDirty() {
	s.value()
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
	rs         *ReactiveSystem
	subs       []Subscriber
	isDirty    bool
	ver        uint32 // value of readoly signals changes atomic increment
	versionSum uint32 // value of dependencies changes
	get        func(T0, T1, T2, T3, T4, T5) O
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
	dep4       Dependency
	dep5       Dependency
	val        O
}

func Computed6[T0, T1, T2, T3, T4, T5, O comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	dep4 Dependency,
	dep5 Dependency,
	get ReadonlySignal6Func[T0, T1, T2, T3, T4, T5, O],
) *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]{
		isDirty:    true,
		get:        get,
		ver:        1,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
		dep4:       dep4,
		dep5:       dep5,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)
	dep4.addSubs(s)
	dep5.addSubs(s)

	return s
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) Value() O {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) value() any {
	if !s.isDirty {
		return s.val
	}
	s.isDirty = false

	var allArgsSum uint32
	depValue0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		depValue0 = zeroT0
	}
	depVersion0 := s.dep0.version()
	allArgsSum += depVersion0
	depValue1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		depValue1 = zeroT1
	}
	depVersion1 := s.dep1.version()
	allArgsSum += depVersion1
	depValue2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		depValue2 = zeroT2
	}
	depVersion2 := s.dep2.version()
	allArgsSum += depVersion2
	depValue3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		depValue3 = zeroT3
	}
	depVersion3 := s.dep3.version()
	allArgsSum += depVersion3
	depValue4, ok4 := s.dep4.value().(T4)
	if !ok4 {
		var zeroT4 T4
		depValue4 = zeroT4
	}
	depVersion4 := s.dep4.version()
	allArgsSum += depVersion4
	depValue5, ok5 := s.dep5.value().(T5)
	if !ok5 {
		var zeroT5 T5
		depValue5 = zeroT5
	}
	depVersion5 := s.dep5.version()
	allArgsSum += depVersion5
	if allArgsSum == s.versionSum {
		return s.val
	}

	s.versionSum = allArgsSum
	currentValue := s.get(
		depValue0,
		depValue1,
		depValue2,
		depValue3,
		depValue4,
		depValue5,
	)
	if s.val == currentValue {
		return s.val
	}
	s.val = currentValue
	s.ver++
	return s.val
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) version() uint32 {
	return s.ver
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) markDirty() {
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) removeSub(toRemove Subscriber) {
	for i, sub := range s.subs {
		if sub == toRemove {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect6[T0, T1, T2, T3, T4, T5 comparable] struct {
	rs         *ReactiveSystem
	versionSum uint32
	fn         func(T0, T1, T2, T3, T4, T5) error
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
	dep4       Dependency
	dep5       Dependency
}

func Effect6[T0, T1, T2, T3, T4, T5 comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	dep4 Dependency,
	dep5 Dependency,
	fn func(T0, T1, T2, T3, T4, T5) error,
) (stop func()) {
	rs.mu.Lock()

	s := &SideEffect6[T0, T1, T2, T3, T4, T5]{
		rs:         rs,
		fn:         fn,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
		dep4:       dep4,
		dep5:       dep5,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)
	dep4.addSubs(s)
	dep5.addSubs(s)

	defer rs.mu.Unlock()

	s.fn(
		dep0.value().(T0),
		dep1.value().(T1),
		dep2.value().(T2),
		dep3.value().(T3),
		dep4.value().(T4),
		dep5.value().(T5),
	)

	return func() {
		dep0.removeSub(s)
		dep1.removeSub(s)
		dep2.removeSub(s)
		dep3.removeSub(s)
		dep4.removeSub(s)
		dep5.removeSub(s)
	}
}

func (s *SideEffect6[T0, T1, T2, T3, T4, T5]) value() any {
	var allArgsSum uint32

	current0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		current0 = zeroT0
	}
	currentVersion0 := s.dep0.version()
	allArgsSum += currentVersion0

	current1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		current1 = zeroT1
	}
	currentVersion1 := s.dep1.version()
	allArgsSum += currentVersion1

	current2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		current2 = zeroT2
	}
	currentVersion2 := s.dep2.version()
	allArgsSum += currentVersion2

	current3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		current3 = zeroT3
	}
	currentVersion3 := s.dep3.version()
	allArgsSum += currentVersion3

	current4, ok4 := s.dep4.value().(T4)
	if !ok4 {
		var zeroT4 T4
		current4 = zeroT4
	}
	currentVersion4 := s.dep4.version()
	allArgsSum += currentVersion4

	current5, ok5 := s.dep5.value().(T5)
	if !ok5 {
		var zeroT5 T5
		current5 = zeroT5
	}
	currentVersion5 := s.dep5.version()
	allArgsSum += currentVersion5

	if allArgsSum == s.versionSum {
		return nil
	}
	s.versionSum = allArgsSum

	s.fn(
		current0,
		current1,
		current2,
		current3,
		current4,
		current5,
	)
	return nil
}

func (s *SideEffect6[T0, T1, T2, T3, T4, T5]) markDirty() {
	s.value()
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
	rs         *ReactiveSystem
	subs       []Subscriber
	isDirty    bool
	ver        uint32 // value of readoly signals changes atomic increment
	versionSum uint32 // value of dependencies changes
	get        func(T0, T1, T2, T3, T4, T5, T6) O
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
	dep4       Dependency
	dep5       Dependency
	dep6       Dependency
	val        O
}

func Computed7[T0, T1, T2, T3, T4, T5, T6, O comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	dep4 Dependency,
	dep5 Dependency,
	dep6 Dependency,
	get ReadonlySignal7Func[T0, T1, T2, T3, T4, T5, T6, O],
) *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]{
		isDirty:    true,
		get:        get,
		ver:        1,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
		dep4:       dep4,
		dep5:       dep5,
		dep6:       dep6,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)
	dep4.addSubs(s)
	dep5.addSubs(s)
	dep6.addSubs(s)

	return s
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) Value() O {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) value() any {
	if !s.isDirty {
		return s.val
	}
	s.isDirty = false

	var allArgsSum uint32
	depValue0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		depValue0 = zeroT0
	}
	depVersion0 := s.dep0.version()
	allArgsSum += depVersion0
	depValue1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		depValue1 = zeroT1
	}
	depVersion1 := s.dep1.version()
	allArgsSum += depVersion1
	depValue2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		depValue2 = zeroT2
	}
	depVersion2 := s.dep2.version()
	allArgsSum += depVersion2
	depValue3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		depValue3 = zeroT3
	}
	depVersion3 := s.dep3.version()
	allArgsSum += depVersion3
	depValue4, ok4 := s.dep4.value().(T4)
	if !ok4 {
		var zeroT4 T4
		depValue4 = zeroT4
	}
	depVersion4 := s.dep4.version()
	allArgsSum += depVersion4
	depValue5, ok5 := s.dep5.value().(T5)
	if !ok5 {
		var zeroT5 T5
		depValue5 = zeroT5
	}
	depVersion5 := s.dep5.version()
	allArgsSum += depVersion5
	depValue6, ok6 := s.dep6.value().(T6)
	if !ok6 {
		var zeroT6 T6
		depValue6 = zeroT6
	}
	depVersion6 := s.dep6.version()
	allArgsSum += depVersion6
	if allArgsSum == s.versionSum {
		return s.val
	}

	s.versionSum = allArgsSum
	currentValue := s.get(
		depValue0,
		depValue1,
		depValue2,
		depValue3,
		depValue4,
		depValue5,
		depValue6,
	)
	if s.val == currentValue {
		return s.val
	}
	s.val = currentValue
	s.ver++
	return s.val
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) version() uint32 {
	return s.ver
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) markDirty() {
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) removeSub(toRemove Subscriber) {
	for i, sub := range s.subs {
		if sub == toRemove {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect7[T0, T1, T2, T3, T4, T5, T6 comparable] struct {
	rs         *ReactiveSystem
	versionSum uint32
	fn         func(T0, T1, T2, T3, T4, T5, T6) error
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
	dep4       Dependency
	dep5       Dependency
	dep6       Dependency
}

func Effect7[T0, T1, T2, T3, T4, T5, T6 comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	dep4 Dependency,
	dep5 Dependency,
	dep6 Dependency,
	fn func(T0, T1, T2, T3, T4, T5, T6) error,
) (stop func()) {
	rs.mu.Lock()

	s := &SideEffect7[T0, T1, T2, T3, T4, T5, T6]{
		rs:         rs,
		fn:         fn,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
		dep4:       dep4,
		dep5:       dep5,
		dep6:       dep6,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)
	dep4.addSubs(s)
	dep5.addSubs(s)
	dep6.addSubs(s)

	defer rs.mu.Unlock()

	s.fn(
		dep0.value().(T0),
		dep1.value().(T1),
		dep2.value().(T2),
		dep3.value().(T3),
		dep4.value().(T4),
		dep5.value().(T5),
		dep6.value().(T6),
	)

	return func() {
		dep0.removeSub(s)
		dep1.removeSub(s)
		dep2.removeSub(s)
		dep3.removeSub(s)
		dep4.removeSub(s)
		dep5.removeSub(s)
		dep6.removeSub(s)
	}
}

func (s *SideEffect7[T0, T1, T2, T3, T4, T5, T6]) value() any {
	var allArgsSum uint32

	current0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		current0 = zeroT0
	}
	currentVersion0 := s.dep0.version()
	allArgsSum += currentVersion0

	current1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		current1 = zeroT1
	}
	currentVersion1 := s.dep1.version()
	allArgsSum += currentVersion1

	current2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		current2 = zeroT2
	}
	currentVersion2 := s.dep2.version()
	allArgsSum += currentVersion2

	current3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		current3 = zeroT3
	}
	currentVersion3 := s.dep3.version()
	allArgsSum += currentVersion3

	current4, ok4 := s.dep4.value().(T4)
	if !ok4 {
		var zeroT4 T4
		current4 = zeroT4
	}
	currentVersion4 := s.dep4.version()
	allArgsSum += currentVersion4

	current5, ok5 := s.dep5.value().(T5)
	if !ok5 {
		var zeroT5 T5
		current5 = zeroT5
	}
	currentVersion5 := s.dep5.version()
	allArgsSum += currentVersion5

	current6, ok6 := s.dep6.value().(T6)
	if !ok6 {
		var zeroT6 T6
		current6 = zeroT6
	}
	currentVersion6 := s.dep6.version()
	allArgsSum += currentVersion6

	if allArgsSum == s.versionSum {
		return nil
	}
	s.versionSum = allArgsSum

	s.fn(
		current0,
		current1,
		current2,
		current3,
		current4,
		current5,
		current6,
	)
	return nil
}

func (s *SideEffect7[T0, T1, T2, T3, T4, T5, T6]) markDirty() {
	s.value()
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
	rs         *ReactiveSystem
	subs       []Subscriber
	isDirty    bool
	ver        uint32 // value of readoly signals changes atomic increment
	versionSum uint32 // value of dependencies changes
	get        func(T0, T1, T2, T3, T4, T5, T6, T7) O
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
	dep4       Dependency
	dep5       Dependency
	dep6       Dependency
	dep7       Dependency
	val        O
}

func Computed8[T0, T1, T2, T3, T4, T5, T6, T7, O comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	dep4 Dependency,
	dep5 Dependency,
	dep6 Dependency,
	dep7 Dependency,
	get ReadonlySignal8Func[T0, T1, T2, T3, T4, T5, T6, T7, O],
) *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]{
		isDirty:    true,
		get:        get,
		ver:        1,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
		dep4:       dep4,
		dep5:       dep5,
		dep6:       dep6,
		dep7:       dep7,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)
	dep4.addSubs(s)
	dep5.addSubs(s)
	dep6.addSubs(s)
	dep7.addSubs(s)

	return s
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) Value() O {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()

	return s.value().(O)
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) value() any {
	if !s.isDirty {
		return s.val
	}
	s.isDirty = false

	var allArgsSum uint32
	depValue0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		depValue0 = zeroT0
	}
	depVersion0 := s.dep0.version()
	allArgsSum += depVersion0
	depValue1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		depValue1 = zeroT1
	}
	depVersion1 := s.dep1.version()
	allArgsSum += depVersion1
	depValue2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		depValue2 = zeroT2
	}
	depVersion2 := s.dep2.version()
	allArgsSum += depVersion2
	depValue3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		depValue3 = zeroT3
	}
	depVersion3 := s.dep3.version()
	allArgsSum += depVersion3
	depValue4, ok4 := s.dep4.value().(T4)
	if !ok4 {
		var zeroT4 T4
		depValue4 = zeroT4
	}
	depVersion4 := s.dep4.version()
	allArgsSum += depVersion4
	depValue5, ok5 := s.dep5.value().(T5)
	if !ok5 {
		var zeroT5 T5
		depValue5 = zeroT5
	}
	depVersion5 := s.dep5.version()
	allArgsSum += depVersion5
	depValue6, ok6 := s.dep6.value().(T6)
	if !ok6 {
		var zeroT6 T6
		depValue6 = zeroT6
	}
	depVersion6 := s.dep6.version()
	allArgsSum += depVersion6
	depValue7, ok7 := s.dep7.value().(T7)
	if !ok7 {
		var zeroT7 T7
		depValue7 = zeroT7
	}
	depVersion7 := s.dep7.version()
	allArgsSum += depVersion7
	if allArgsSum == s.versionSum {
		return s.val
	}

	s.versionSum = allArgsSum
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
	if s.val == currentValue {
		return s.val
	}
	s.val = currentValue
	s.ver++
	return s.val
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) version() uint32 {
	return s.ver
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) markDirty() {
	s.isDirty = true
	for _, sub := range s.subs {
		sub.markDirty()
	}
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) addSubs(subs ...Subscriber) {
	s.subs = append(s.subs, subs...)
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) removeSub(toRemove Subscriber) {
	for i, sub := range s.subs {
		if sub == toRemove {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

type SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7 comparable] struct {
	rs         *ReactiveSystem
	versionSum uint32
	fn         func(T0, T1, T2, T3, T4, T5, T6, T7) error
	dep0       Dependency
	dep1       Dependency
	dep2       Dependency
	dep3       Dependency
	dep4       Dependency
	dep5       Dependency
	dep6       Dependency
	dep7       Dependency
}

func Effect8[T0, T1, T2, T3, T4, T5, T6, T7 comparable](
	rs *ReactiveSystem,
	dep0 Dependency,
	dep1 Dependency,
	dep2 Dependency,
	dep3 Dependency,
	dep4 Dependency,
	dep5 Dependency,
	dep6 Dependency,
	dep7 Dependency,
	fn func(T0, T1, T2, T3, T4, T5, T6, T7) error,
) (stop func()) {
	rs.mu.Lock()

	s := &SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]{
		rs:         rs,
		fn:         fn,
		versionSum: 0,
		dep0:       dep0,
		dep1:       dep1,
		dep2:       dep2,
		dep3:       dep3,
		dep4:       dep4,
		dep5:       dep5,
		dep6:       dep6,
		dep7:       dep7,
	}
	dep0.addSubs(s)
	dep1.addSubs(s)
	dep2.addSubs(s)
	dep3.addSubs(s)
	dep4.addSubs(s)
	dep5.addSubs(s)
	dep6.addSubs(s)
	dep7.addSubs(s)

	defer rs.mu.Unlock()

	s.fn(
		dep0.value().(T0),
		dep1.value().(T1),
		dep2.value().(T2),
		dep3.value().(T3),
		dep4.value().(T4),
		dep5.value().(T5),
		dep6.value().(T6),
		dep7.value().(T7),
	)

	return func() {
		dep0.removeSub(s)
		dep1.removeSub(s)
		dep2.removeSub(s)
		dep3.removeSub(s)
		dep4.removeSub(s)
		dep5.removeSub(s)
		dep6.removeSub(s)
		dep7.removeSub(s)
	}
}

func (s *SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]) value() any {
	var allArgsSum uint32

	current0, ok0 := s.dep0.value().(T0)
	if !ok0 {
		var zeroT0 T0
		current0 = zeroT0
	}
	currentVersion0 := s.dep0.version()
	allArgsSum += currentVersion0

	current1, ok1 := s.dep1.value().(T1)
	if !ok1 {
		var zeroT1 T1
		current1 = zeroT1
	}
	currentVersion1 := s.dep1.version()
	allArgsSum += currentVersion1

	current2, ok2 := s.dep2.value().(T2)
	if !ok2 {
		var zeroT2 T2
		current2 = zeroT2
	}
	currentVersion2 := s.dep2.version()
	allArgsSum += currentVersion2

	current3, ok3 := s.dep3.value().(T3)
	if !ok3 {
		var zeroT3 T3
		current3 = zeroT3
	}
	currentVersion3 := s.dep3.version()
	allArgsSum += currentVersion3

	current4, ok4 := s.dep4.value().(T4)
	if !ok4 {
		var zeroT4 T4
		current4 = zeroT4
	}
	currentVersion4 := s.dep4.version()
	allArgsSum += currentVersion4

	current5, ok5 := s.dep5.value().(T5)
	if !ok5 {
		var zeroT5 T5
		current5 = zeroT5
	}
	currentVersion5 := s.dep5.version()
	allArgsSum += currentVersion5

	current6, ok6 := s.dep6.value().(T6)
	if !ok6 {
		var zeroT6 T6
		current6 = zeroT6
	}
	currentVersion6 := s.dep6.version()
	allArgsSum += currentVersion6

	current7, ok7 := s.dep7.value().(T7)
	if !ok7 {
		var zeroT7 T7
		current7 = zeroT7
	}
	currentVersion7 := s.dep7.version()
	allArgsSum += currentVersion7

	if allArgsSum == s.versionSum {
		return nil
	}
	s.versionSum = allArgsSum

	s.fn(
		current0,
		current1,
		current2,
		current3,
		current4,
		current5,
		current6,
		current7,
	)
	return nil
}

func (s *SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]) markDirty() {
	s.value()
}
