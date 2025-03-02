package dumbdumb

import "sync"

const (
	DefaultCellCacheSize = 4096
)

type state uint8

const (
	clean state = iota
	dirty
	computing
)

type Cell interface {
	setDirty()
	eval() any
}

type ReactiveSystem struct {
	mu       *sync.Mutex
	cells    []Cell
	anyDirty bool
}

func NewReactiveSystem() *ReactiveSystem {
	return &ReactiveSystem{
		mu:       &sync.Mutex{},
		cells:    make([]Cell, 0, DefaultCellCacheSize),
		anyDirty: false,
	}
}

func (rs *ReactiveSystem) Reset() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.anyDirty = false
	rs.cells = rs.cells[:0]
}

func (rs *ReactiveSystem) Remove(cells ...Cell) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.remove(cells...)
}

func (rs *ReactiveSystem) remove(cells ...Cell) {
	removeCount := 0
	for _, cell := range cells {
		for i, c := range rs.cells {
			if c == cell {
				// move the last element to the current position
				rs.cells[i] = rs.cells[len(rs.cells)-1]
				removeCount++
			}
		}
	}
	rs.cells = rs.cells[:len(rs.cells)-removeCount]
}

func (rs *ReactiveSystem) evalAll() {
	for _, c := range rs.cells {
		c.setDirty()
	}
	for _, c := range rs.cells {
		c.eval()
	}
}

type WriteableSignal[T comparable] struct {
	rs    *ReactiveSystem
	v     T
	state state
}

func (s *WriteableSignal[T]) Value() T {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()
	if s.rs.anyDirty {
		s.rs.evalAll()
	}
	return s.v
}

func (s *WriteableSignal[T]) SetValue(value T) {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()
	s.state = dirty
	s.rs.anyDirty = true
	s.v = value
	s.rs.evalAll()
}

func (s *WriteableSignal[T]) setDirty() {
	// noop
}

func (s *WriteableSignal[T]) eval() any {
	return s.v
}

func Signal[T comparable](rs *ReactiveSystem, value T) *WriteableSignal[T] {
	s := &WriteableSignal[T]{rs: rs, v: value}
	rs.cells = append(rs.cells, s)
	return s
}

func roSig[O any](rs *ReactiveSystem) ReadonlySignal[O] {
	rs.anyDirty = true
	return ReadonlySignal[O]{state: dirty, rs: rs}
}

type ReadonlySignal[O any] struct {
	state state
	rs    *ReactiveSystem
	value O
}

func (s *ReadonlySignal[O]) Value() O {
	s.rs.mu.Lock()
	defer s.rs.mu.Unlock()
	if s.rs.anyDirty {
		s.rs.evalAll()
	}
	return s.value
}

func (s *ReadonlySignal[O]) setDirty() {
	s.state = dirty
}

func (s *ReadonlySignal[O]) preEval() (o O, wasClean bool) {
	if s.state == computing {
		panic("circular dependency")
	} else if s.state == clean {
		return s.value, true
	}

	s.state = computing
	return o, false
}

func (s *ReadonlySignal[O]) postEval(v O) O {
	s.value = v
	s.state = clean
	return v
}

type SideEffect struct {
	rs    *ReactiveSystem
	state state
}

type ReadonlySignal1[T0, O comparable] struct {
	ReadonlySignal[O]
	cell0   Cell
	cached0 T0
	getter  func(T0) O
}

func (s *ReadonlySignal1[T0, O]) eval() any {
	v, wasClean := s.preEval()
	if wasClean {
		return v
	}

	allMatch := true
	arg0 := s.cell0.eval().(T0)
	if arg0 != s.cached0 {
		allMatch = false
		s.cached0 = arg0
	}
	if allMatch {
		return s.postEval(s.value)
	}

	v = s.getter(
		arg0,
	)
	return s.postEval(v)
}

func Computed1[T0, O comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	f func(T0) O,
) *ReadonlySignal1[T0, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal1[T0, O]{
		ReadonlySignal: roSig[O](rs),
		getter:         f,
		cell0:          cell0,
	}
	s.cached0 = cell0.eval().(T0)
	s.value = f(
		s.cached0,
	)
	rs.cells = append(rs.cells, s)
	return s
}

type SideEffect1[T0 comparable] struct {
	SideEffect
	fn      func(T0) error
	cell0   Cell
	cached0 T0
}

func (e *SideEffect1[T0]) setDirty() {
	e.state = dirty
}

func (e *SideEffect1[T0]) eval() any {
	if e.state == computing {
		panic("circular dependency")
	} else if e.state == clean {
		return nil
	}
	e.state = computing

	allMatch := true
	arg0 := e.cell0.eval().(T0)
	if arg0 != e.cached0 {
		allMatch = false
		e.cached0 = arg0
	}

	if allMatch {
		e.state = clean
		return nil
	}

	err := e.fn(
		arg0,
	)
	e.state = clean
	return err
}

func Effect1[T0 comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	fn func(T0) error,
) (stop func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	e := &SideEffect1[T0]{
		SideEffect: SideEffect{state: dirty, rs: rs},
		cell0:      cell0,
		fn:         fn,
	}
	rs.cells = append(rs.cells, e)
	rs.anyDirty = true
	rs.evalAll()

	return func() {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		rs.remove(e)
	}
}

type ReadonlySignal2[T0, T1, O comparable] struct {
	ReadonlySignal[O]
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	getter  func(T0, T1) O
}

func (s *ReadonlySignal2[T0, T1, O]) eval() any {
	v, wasClean := s.preEval()
	if wasClean {
		return v
	}

	allMatch := true
	arg0 := s.cell0.eval().(T0)
	if arg0 != s.cached0 {
		allMatch = false
		s.cached0 = arg0
	}
	arg1 := s.cell1.eval().(T1)
	if arg1 != s.cached1 {
		allMatch = false
		s.cached1 = arg1
	}
	if allMatch {
		return s.postEval(s.value)
	}

	v = s.getter(
		arg0,
		arg1,
	)
	return s.postEval(v)
}

func Computed2[T0, T1, O comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	f func(T0, T1) O,
) *ReadonlySignal2[T0, T1, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal2[T0, T1, O]{
		ReadonlySignal: roSig[O](rs),
		getter:         f,
		cell0:          cell0,
		cell1:          cell1,
	}
	s.cached0 = cell0.eval().(T0)
	s.cached1 = cell1.eval().(T1)
	s.value = f(
		s.cached0,
		s.cached1,
	)
	rs.cells = append(rs.cells, s)
	return s
}

type SideEffect2[T0, T1 comparable] struct {
	SideEffect
	fn      func(T0, T1) error
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
}

func (e *SideEffect2[T0, T1]) setDirty() {
	e.state = dirty
}

func (e *SideEffect2[T0, T1]) eval() any {
	if e.state == computing {
		panic("circular dependency")
	} else if e.state == clean {
		return nil
	}
	e.state = computing

	allMatch := true
	arg0 := e.cell0.eval().(T0)
	if arg0 != e.cached0 {
		allMatch = false
		e.cached0 = arg0
	}
	arg1 := e.cell1.eval().(T1)
	if arg1 != e.cached1 {
		allMatch = false
		e.cached1 = arg1
	}

	if allMatch {
		e.state = clean
		return nil
	}

	err := e.fn(
		arg0,
		arg1,
	)
	e.state = clean
	return err
}

func Effect2[T0, T1 comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	fn func(T0, T1) error,
) (stop func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	e := &SideEffect2[T0, T1]{
		SideEffect: SideEffect{state: dirty, rs: rs},
		cell0:      cell0,
		cell1:      cell1,
		fn:         fn,
	}
	rs.cells = append(rs.cells, e)
	rs.anyDirty = true
	rs.evalAll()

	return func() {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		rs.remove(e)
	}
}

type ReadonlySignal3[T0, T1, T2, O comparable] struct {
	ReadonlySignal[O]
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	getter  func(T0, T1, T2) O
}

func (s *ReadonlySignal3[T0, T1, T2, O]) eval() any {
	v, wasClean := s.preEval()
	if wasClean {
		return v
	}

	allMatch := true
	arg0 := s.cell0.eval().(T0)
	if arg0 != s.cached0 {
		allMatch = false
		s.cached0 = arg0
	}
	arg1 := s.cell1.eval().(T1)
	if arg1 != s.cached1 {
		allMatch = false
		s.cached1 = arg1
	}
	arg2 := s.cell2.eval().(T2)
	if arg2 != s.cached2 {
		allMatch = false
		s.cached2 = arg2
	}
	if allMatch {
		return s.postEval(s.value)
	}

	v = s.getter(
		arg0,
		arg1,
		arg2,
	)
	return s.postEval(v)
}

func Computed3[T0, T1, T2, O comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	f func(T0, T1, T2) O,
) *ReadonlySignal3[T0, T1, T2, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal3[T0, T1, T2, O]{
		ReadonlySignal: roSig[O](rs),
		getter:         f,
		cell0:          cell0,
		cell1:          cell1,
		cell2:          cell2,
	}
	s.cached0 = cell0.eval().(T0)
	s.cached1 = cell1.eval().(T1)
	s.cached2 = cell2.eval().(T2)
	s.value = f(
		s.cached0,
		s.cached1,
		s.cached2,
	)
	rs.cells = append(rs.cells, s)
	return s
}

type SideEffect3[T0, T1, T2 comparable] struct {
	SideEffect
	fn      func(T0, T1, T2) error
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
}

func (e *SideEffect3[T0, T1, T2]) setDirty() {
	e.state = dirty
}

func (e *SideEffect3[T0, T1, T2]) eval() any {
	if e.state == computing {
		panic("circular dependency")
	} else if e.state == clean {
		return nil
	}
	e.state = computing

	allMatch := true
	arg0 := e.cell0.eval().(T0)
	if arg0 != e.cached0 {
		allMatch = false
		e.cached0 = arg0
	}
	arg1 := e.cell1.eval().(T1)
	if arg1 != e.cached1 {
		allMatch = false
		e.cached1 = arg1
	}
	arg2 := e.cell2.eval().(T2)
	if arg2 != e.cached2 {
		allMatch = false
		e.cached2 = arg2
	}

	if allMatch {
		e.state = clean
		return nil
	}

	err := e.fn(
		arg0,
		arg1,
		arg2,
	)
	e.state = clean
	return err
}

func Effect3[T0, T1, T2 comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	fn func(T0, T1, T2) error,
) (stop func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	e := &SideEffect3[T0, T1, T2]{
		SideEffect: SideEffect{state: dirty, rs: rs},
		cell0:      cell0,
		cell1:      cell1,
		cell2:      cell2,
		fn:         fn,
	}
	rs.cells = append(rs.cells, e)
	rs.anyDirty = true
	rs.evalAll()

	return func() {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		rs.remove(e)
	}
}

type ReadonlySignal4[T0, T1, T2, T3, O comparable] struct {
	ReadonlySignal[O]
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
	getter  func(T0, T1, T2, T3) O
}

func (s *ReadonlySignal4[T0, T1, T2, T3, O]) eval() any {
	v, wasClean := s.preEval()
	if wasClean {
		return v
	}

	allMatch := true
	arg0 := s.cell0.eval().(T0)
	if arg0 != s.cached0 {
		allMatch = false
		s.cached0 = arg0
	}
	arg1 := s.cell1.eval().(T1)
	if arg1 != s.cached1 {
		allMatch = false
		s.cached1 = arg1
	}
	arg2 := s.cell2.eval().(T2)
	if arg2 != s.cached2 {
		allMatch = false
		s.cached2 = arg2
	}
	arg3 := s.cell3.eval().(T3)
	if arg3 != s.cached3 {
		allMatch = false
		s.cached3 = arg3
	}
	if allMatch {
		return s.postEval(s.value)
	}

	v = s.getter(
		arg0,
		arg1,
		arg2,
		arg3,
	)
	return s.postEval(v)
}

func Computed4[T0, T1, T2, T3, O comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	f func(T0, T1, T2, T3) O,
) *ReadonlySignal4[T0, T1, T2, T3, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal4[T0, T1, T2, T3, O]{
		ReadonlySignal: roSig[O](rs),
		getter:         f,
		cell0:          cell0,
		cell1:          cell1,
		cell2:          cell2,
		cell3:          cell3,
	}
	s.cached0 = cell0.eval().(T0)
	s.cached1 = cell1.eval().(T1)
	s.cached2 = cell2.eval().(T2)
	s.cached3 = cell3.eval().(T3)
	s.value = f(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
	)
	rs.cells = append(rs.cells, s)
	return s
}

type SideEffect4[T0, T1, T2, T3 comparable] struct {
	SideEffect
	fn      func(T0, T1, T2, T3) error
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
}

func (e *SideEffect4[T0, T1, T2, T3]) setDirty() {
	e.state = dirty
}

func (e *SideEffect4[T0, T1, T2, T3]) eval() any {
	if e.state == computing {
		panic("circular dependency")
	} else if e.state == clean {
		return nil
	}
	e.state = computing

	allMatch := true
	arg0 := e.cell0.eval().(T0)
	if arg0 != e.cached0 {
		allMatch = false
		e.cached0 = arg0
	}
	arg1 := e.cell1.eval().(T1)
	if arg1 != e.cached1 {
		allMatch = false
		e.cached1 = arg1
	}
	arg2 := e.cell2.eval().(T2)
	if arg2 != e.cached2 {
		allMatch = false
		e.cached2 = arg2
	}
	arg3 := e.cell3.eval().(T3)
	if arg3 != e.cached3 {
		allMatch = false
		e.cached3 = arg3
	}

	if allMatch {
		e.state = clean
		return nil
	}

	err := e.fn(
		arg0,
		arg1,
		arg2,
		arg3,
	)
	e.state = clean
	return err
}

func Effect4[T0, T1, T2, T3 comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	fn func(T0, T1, T2, T3) error,
) (stop func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	e := &SideEffect4[T0, T1, T2, T3]{
		SideEffect: SideEffect{state: dirty, rs: rs},
		cell0:      cell0,
		cell1:      cell1,
		cell2:      cell2,
		cell3:      cell3,
		fn:         fn,
	}
	rs.cells = append(rs.cells, e)
	rs.anyDirty = true
	rs.evalAll()

	return func() {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		rs.remove(e)
	}
}

type ReadonlySignal5[T0, T1, T2, T3, T4, O comparable] struct {
	ReadonlySignal[O]
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
	cell4   Cell
	cached4 T4
	getter  func(T0, T1, T2, T3, T4) O
}

func (s *ReadonlySignal5[T0, T1, T2, T3, T4, O]) eval() any {
	v, wasClean := s.preEval()
	if wasClean {
		return v
	}

	allMatch := true
	arg0 := s.cell0.eval().(T0)
	if arg0 != s.cached0 {
		allMatch = false
		s.cached0 = arg0
	}
	arg1 := s.cell1.eval().(T1)
	if arg1 != s.cached1 {
		allMatch = false
		s.cached1 = arg1
	}
	arg2 := s.cell2.eval().(T2)
	if arg2 != s.cached2 {
		allMatch = false
		s.cached2 = arg2
	}
	arg3 := s.cell3.eval().(T3)
	if arg3 != s.cached3 {
		allMatch = false
		s.cached3 = arg3
	}
	arg4 := s.cell4.eval().(T4)
	if arg4 != s.cached4 {
		allMatch = false
		s.cached4 = arg4
	}
	if allMatch {
		return s.postEval(s.value)
	}

	v = s.getter(
		arg0,
		arg1,
		arg2,
		arg3,
		arg4,
	)
	return s.postEval(v)
}

func Computed5[T0, T1, T2, T3, T4, O comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	cell4 Cell,
	f func(T0, T1, T2, T3, T4) O,
) *ReadonlySignal5[T0, T1, T2, T3, T4, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal5[T0, T1, T2, T3, T4, O]{
		ReadonlySignal: roSig[O](rs),
		getter:         f,
		cell0:          cell0,
		cell1:          cell1,
		cell2:          cell2,
		cell3:          cell3,
		cell4:          cell4,
	}
	s.cached0 = cell0.eval().(T0)
	s.cached1 = cell1.eval().(T1)
	s.cached2 = cell2.eval().(T2)
	s.cached3 = cell3.eval().(T3)
	s.cached4 = cell4.eval().(T4)
	s.value = f(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
	)
	rs.cells = append(rs.cells, s)
	return s
}

type SideEffect5[T0, T1, T2, T3, T4 comparable] struct {
	SideEffect
	fn      func(T0, T1, T2, T3, T4) error
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
	cell4   Cell
	cached4 T4
}

func (e *SideEffect5[T0, T1, T2, T3, T4]) setDirty() {
	e.state = dirty
}

func (e *SideEffect5[T0, T1, T2, T3, T4]) eval() any {
	if e.state == computing {
		panic("circular dependency")
	} else if e.state == clean {
		return nil
	}
	e.state = computing

	allMatch := true
	arg0 := e.cell0.eval().(T0)
	if arg0 != e.cached0 {
		allMatch = false
		e.cached0 = arg0
	}
	arg1 := e.cell1.eval().(T1)
	if arg1 != e.cached1 {
		allMatch = false
		e.cached1 = arg1
	}
	arg2 := e.cell2.eval().(T2)
	if arg2 != e.cached2 {
		allMatch = false
		e.cached2 = arg2
	}
	arg3 := e.cell3.eval().(T3)
	if arg3 != e.cached3 {
		allMatch = false
		e.cached3 = arg3
	}
	arg4 := e.cell4.eval().(T4)
	if arg4 != e.cached4 {
		allMatch = false
		e.cached4 = arg4
	}

	if allMatch {
		e.state = clean
		return nil
	}

	err := e.fn(
		arg0,
		arg1,
		arg2,
		arg3,
		arg4,
	)
	e.state = clean
	return err
}

func Effect5[T0, T1, T2, T3, T4 comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	cell4 Cell,
	fn func(T0, T1, T2, T3, T4) error,
) (stop func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	e := &SideEffect5[T0, T1, T2, T3, T4]{
		SideEffect: SideEffect{state: dirty, rs: rs},
		cell0:      cell0,
		cell1:      cell1,
		cell2:      cell2,
		cell3:      cell3,
		cell4:      cell4,
		fn:         fn,
	}
	rs.cells = append(rs.cells, e)
	rs.anyDirty = true
	rs.evalAll()

	return func() {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		rs.remove(e)
	}
}

type ReadonlySignal6[T0, T1, T2, T3, T4, T5, O comparable] struct {
	ReadonlySignal[O]
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
	cell4   Cell
	cached4 T4
	cell5   Cell
	cached5 T5
	getter  func(T0, T1, T2, T3, T4, T5) O
}

func (s *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]) eval() any {
	v, wasClean := s.preEval()
	if wasClean {
		return v
	}

	allMatch := true
	arg0 := s.cell0.eval().(T0)
	if arg0 != s.cached0 {
		allMatch = false
		s.cached0 = arg0
	}
	arg1 := s.cell1.eval().(T1)
	if arg1 != s.cached1 {
		allMatch = false
		s.cached1 = arg1
	}
	arg2 := s.cell2.eval().(T2)
	if arg2 != s.cached2 {
		allMatch = false
		s.cached2 = arg2
	}
	arg3 := s.cell3.eval().(T3)
	if arg3 != s.cached3 {
		allMatch = false
		s.cached3 = arg3
	}
	arg4 := s.cell4.eval().(T4)
	if arg4 != s.cached4 {
		allMatch = false
		s.cached4 = arg4
	}
	arg5 := s.cell5.eval().(T5)
	if arg5 != s.cached5 {
		allMatch = false
		s.cached5 = arg5
	}
	if allMatch {
		return s.postEval(s.value)
	}

	v = s.getter(
		arg0,
		arg1,
		arg2,
		arg3,
		arg4,
		arg5,
	)
	return s.postEval(v)
}

func Computed6[T0, T1, T2, T3, T4, T5, O comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	cell4 Cell,
	cell5 Cell,
	f func(T0, T1, T2, T3, T4, T5) O,
) *ReadonlySignal6[T0, T1, T2, T3, T4, T5, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal6[T0, T1, T2, T3, T4, T5, O]{
		ReadonlySignal: roSig[O](rs),
		getter:         f,
		cell0:          cell0,
		cell1:          cell1,
		cell2:          cell2,
		cell3:          cell3,
		cell4:          cell4,
		cell5:          cell5,
	}
	s.cached0 = cell0.eval().(T0)
	s.cached1 = cell1.eval().(T1)
	s.cached2 = cell2.eval().(T2)
	s.cached3 = cell3.eval().(T3)
	s.cached4 = cell4.eval().(T4)
	s.cached5 = cell5.eval().(T5)
	s.value = f(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
	)
	rs.cells = append(rs.cells, s)
	return s
}

type SideEffect6[T0, T1, T2, T3, T4, T5 comparable] struct {
	SideEffect
	fn      func(T0, T1, T2, T3, T4, T5) error
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
	cell4   Cell
	cached4 T4
	cell5   Cell
	cached5 T5
}

func (e *SideEffect6[T0, T1, T2, T3, T4, T5]) setDirty() {
	e.state = dirty
}

func (e *SideEffect6[T0, T1, T2, T3, T4, T5]) eval() any {
	if e.state == computing {
		panic("circular dependency")
	} else if e.state == clean {
		return nil
	}
	e.state = computing

	allMatch := true
	arg0 := e.cell0.eval().(T0)
	if arg0 != e.cached0 {
		allMatch = false
		e.cached0 = arg0
	}
	arg1 := e.cell1.eval().(T1)
	if arg1 != e.cached1 {
		allMatch = false
		e.cached1 = arg1
	}
	arg2 := e.cell2.eval().(T2)
	if arg2 != e.cached2 {
		allMatch = false
		e.cached2 = arg2
	}
	arg3 := e.cell3.eval().(T3)
	if arg3 != e.cached3 {
		allMatch = false
		e.cached3 = arg3
	}
	arg4 := e.cell4.eval().(T4)
	if arg4 != e.cached4 {
		allMatch = false
		e.cached4 = arg4
	}
	arg5 := e.cell5.eval().(T5)
	if arg5 != e.cached5 {
		allMatch = false
		e.cached5 = arg5
	}

	if allMatch {
		e.state = clean
		return nil
	}

	err := e.fn(
		arg0,
		arg1,
		arg2,
		arg3,
		arg4,
		arg5,
	)
	e.state = clean
	return err
}

func Effect6[T0, T1, T2, T3, T4, T5 comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	cell4 Cell,
	cell5 Cell,
	fn func(T0, T1, T2, T3, T4, T5) error,
) (stop func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	e := &SideEffect6[T0, T1, T2, T3, T4, T5]{
		SideEffect: SideEffect{state: dirty, rs: rs},
		cell0:      cell0,
		cell1:      cell1,
		cell2:      cell2,
		cell3:      cell3,
		cell4:      cell4,
		cell5:      cell5,
		fn:         fn,
	}
	rs.cells = append(rs.cells, e)
	rs.anyDirty = true
	rs.evalAll()

	return func() {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		rs.remove(e)
	}
}

type ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O comparable] struct {
	ReadonlySignal[O]
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
	cell4   Cell
	cached4 T4
	cell5   Cell
	cached5 T5
	cell6   Cell
	cached6 T6
	getter  func(T0, T1, T2, T3, T4, T5, T6) O
}

func (s *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]) eval() any {
	v, wasClean := s.preEval()
	if wasClean {
		return v
	}

	allMatch := true
	arg0 := s.cell0.eval().(T0)
	if arg0 != s.cached0 {
		allMatch = false
		s.cached0 = arg0
	}
	arg1 := s.cell1.eval().(T1)
	if arg1 != s.cached1 {
		allMatch = false
		s.cached1 = arg1
	}
	arg2 := s.cell2.eval().(T2)
	if arg2 != s.cached2 {
		allMatch = false
		s.cached2 = arg2
	}
	arg3 := s.cell3.eval().(T3)
	if arg3 != s.cached3 {
		allMatch = false
		s.cached3 = arg3
	}
	arg4 := s.cell4.eval().(T4)
	if arg4 != s.cached4 {
		allMatch = false
		s.cached4 = arg4
	}
	arg5 := s.cell5.eval().(T5)
	if arg5 != s.cached5 {
		allMatch = false
		s.cached5 = arg5
	}
	arg6 := s.cell6.eval().(T6)
	if arg6 != s.cached6 {
		allMatch = false
		s.cached6 = arg6
	}
	if allMatch {
		return s.postEval(s.value)
	}

	v = s.getter(
		arg0,
		arg1,
		arg2,
		arg3,
		arg4,
		arg5,
		arg6,
	)
	return s.postEval(v)
}

func Computed7[T0, T1, T2, T3, T4, T5, T6, O comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	cell4 Cell,
	cell5 Cell,
	cell6 Cell,
	f func(T0, T1, T2, T3, T4, T5, T6) O,
) *ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal7[T0, T1, T2, T3, T4, T5, T6, O]{
		ReadonlySignal: roSig[O](rs),
		getter:         f,
		cell0:          cell0,
		cell1:          cell1,
		cell2:          cell2,
		cell3:          cell3,
		cell4:          cell4,
		cell5:          cell5,
		cell6:          cell6,
	}
	s.cached0 = cell0.eval().(T0)
	s.cached1 = cell1.eval().(T1)
	s.cached2 = cell2.eval().(T2)
	s.cached3 = cell3.eval().(T3)
	s.cached4 = cell4.eval().(T4)
	s.cached5 = cell5.eval().(T5)
	s.cached6 = cell6.eval().(T6)
	s.value = f(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
		s.cached6,
	)
	rs.cells = append(rs.cells, s)
	return s
}

type SideEffect7[T0, T1, T2, T3, T4, T5, T6 comparable] struct {
	SideEffect
	fn      func(T0, T1, T2, T3, T4, T5, T6) error
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
	cell4   Cell
	cached4 T4
	cell5   Cell
	cached5 T5
	cell6   Cell
	cached6 T6
}

func (e *SideEffect7[T0, T1, T2, T3, T4, T5, T6]) setDirty() {
	e.state = dirty
}

func (e *SideEffect7[T0, T1, T2, T3, T4, T5, T6]) eval() any {
	if e.state == computing {
		panic("circular dependency")
	} else if e.state == clean {
		return nil
	}
	e.state = computing

	allMatch := true
	arg0 := e.cell0.eval().(T0)
	if arg0 != e.cached0 {
		allMatch = false
		e.cached0 = arg0
	}
	arg1 := e.cell1.eval().(T1)
	if arg1 != e.cached1 {
		allMatch = false
		e.cached1 = arg1
	}
	arg2 := e.cell2.eval().(T2)
	if arg2 != e.cached2 {
		allMatch = false
		e.cached2 = arg2
	}
	arg3 := e.cell3.eval().(T3)
	if arg3 != e.cached3 {
		allMatch = false
		e.cached3 = arg3
	}
	arg4 := e.cell4.eval().(T4)
	if arg4 != e.cached4 {
		allMatch = false
		e.cached4 = arg4
	}
	arg5 := e.cell5.eval().(T5)
	if arg5 != e.cached5 {
		allMatch = false
		e.cached5 = arg5
	}
	arg6 := e.cell6.eval().(T6)
	if arg6 != e.cached6 {
		allMatch = false
		e.cached6 = arg6
	}

	if allMatch {
		e.state = clean
		return nil
	}

	err := e.fn(
		arg0,
		arg1,
		arg2,
		arg3,
		arg4,
		arg5,
		arg6,
	)
	e.state = clean
	return err
}

func Effect7[T0, T1, T2, T3, T4, T5, T6 comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	cell4 Cell,
	cell5 Cell,
	cell6 Cell,
	fn func(T0, T1, T2, T3, T4, T5, T6) error,
) (stop func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	e := &SideEffect7[T0, T1, T2, T3, T4, T5, T6]{
		SideEffect: SideEffect{state: dirty, rs: rs},
		cell0:      cell0,
		cell1:      cell1,
		cell2:      cell2,
		cell3:      cell3,
		cell4:      cell4,
		cell5:      cell5,
		cell6:      cell6,
		fn:         fn,
	}
	rs.cells = append(rs.cells, e)
	rs.anyDirty = true
	rs.evalAll()

	return func() {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		rs.remove(e)
	}
}

type ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O comparable] struct {
	ReadonlySignal[O]
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
	cell4   Cell
	cached4 T4
	cell5   Cell
	cached5 T5
	cell6   Cell
	cached6 T6
	cell7   Cell
	cached7 T7
	getter  func(T0, T1, T2, T3, T4, T5, T6, T7) O
}

func (s *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]) eval() any {
	v, wasClean := s.preEval()
	if wasClean {
		return v
	}

	allMatch := true
	arg0 := s.cell0.eval().(T0)
	if arg0 != s.cached0 {
		allMatch = false
		s.cached0 = arg0
	}
	arg1 := s.cell1.eval().(T1)
	if arg1 != s.cached1 {
		allMatch = false
		s.cached1 = arg1
	}
	arg2 := s.cell2.eval().(T2)
	if arg2 != s.cached2 {
		allMatch = false
		s.cached2 = arg2
	}
	arg3 := s.cell3.eval().(T3)
	if arg3 != s.cached3 {
		allMatch = false
		s.cached3 = arg3
	}
	arg4 := s.cell4.eval().(T4)
	if arg4 != s.cached4 {
		allMatch = false
		s.cached4 = arg4
	}
	arg5 := s.cell5.eval().(T5)
	if arg5 != s.cached5 {
		allMatch = false
		s.cached5 = arg5
	}
	arg6 := s.cell6.eval().(T6)
	if arg6 != s.cached6 {
		allMatch = false
		s.cached6 = arg6
	}
	arg7 := s.cell7.eval().(T7)
	if arg7 != s.cached7 {
		allMatch = false
		s.cached7 = arg7
	}
	if allMatch {
		return s.postEval(s.value)
	}

	v = s.getter(
		arg0,
		arg1,
		arg2,
		arg3,
		arg4,
		arg5,
		arg6,
		arg7,
	)
	return s.postEval(v)
}

func Computed8[T0, T1, T2, T3, T4, T5, T6, T7, O comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	cell4 Cell,
	cell5 Cell,
	cell6 Cell,
	cell7 Cell,
	f func(T0, T1, T2, T3, T4, T5, T6, T7) O,
) *ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O] {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	s := &ReadonlySignal8[T0, T1, T2, T3, T4, T5, T6, T7, O]{
		ReadonlySignal: roSig[O](rs),
		getter:         f,
		cell0:          cell0,
		cell1:          cell1,
		cell2:          cell2,
		cell3:          cell3,
		cell4:          cell4,
		cell5:          cell5,
		cell6:          cell6,
		cell7:          cell7,
	}
	s.cached0 = cell0.eval().(T0)
	s.cached1 = cell1.eval().(T1)
	s.cached2 = cell2.eval().(T2)
	s.cached3 = cell3.eval().(T3)
	s.cached4 = cell4.eval().(T4)
	s.cached5 = cell5.eval().(T5)
	s.cached6 = cell6.eval().(T6)
	s.cached7 = cell7.eval().(T7)
	s.value = f(
		s.cached0,
		s.cached1,
		s.cached2,
		s.cached3,
		s.cached4,
		s.cached5,
		s.cached6,
		s.cached7,
	)
	rs.cells = append(rs.cells, s)
	return s
}

type SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7 comparable] struct {
	SideEffect
	fn      func(T0, T1, T2, T3, T4, T5, T6, T7) error
	cell0   Cell
	cached0 T0
	cell1   Cell
	cached1 T1
	cell2   Cell
	cached2 T2
	cell3   Cell
	cached3 T3
	cell4   Cell
	cached4 T4
	cell5   Cell
	cached5 T5
	cell6   Cell
	cached6 T6
	cell7   Cell
	cached7 T7
}

func (e *SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]) setDirty() {
	e.state = dirty
}

func (e *SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]) eval() any {
	if e.state == computing {
		panic("circular dependency")
	} else if e.state == clean {
		return nil
	}
	e.state = computing

	allMatch := true
	arg0 := e.cell0.eval().(T0)
	if arg0 != e.cached0 {
		allMatch = false
		e.cached0 = arg0
	}
	arg1 := e.cell1.eval().(T1)
	if arg1 != e.cached1 {
		allMatch = false
		e.cached1 = arg1
	}
	arg2 := e.cell2.eval().(T2)
	if arg2 != e.cached2 {
		allMatch = false
		e.cached2 = arg2
	}
	arg3 := e.cell3.eval().(T3)
	if arg3 != e.cached3 {
		allMatch = false
		e.cached3 = arg3
	}
	arg4 := e.cell4.eval().(T4)
	if arg4 != e.cached4 {
		allMatch = false
		e.cached4 = arg4
	}
	arg5 := e.cell5.eval().(T5)
	if arg5 != e.cached5 {
		allMatch = false
		e.cached5 = arg5
	}
	arg6 := e.cell6.eval().(T6)
	if arg6 != e.cached6 {
		allMatch = false
		e.cached6 = arg6
	}
	arg7 := e.cell7.eval().(T7)
	if arg7 != e.cached7 {
		allMatch = false
		e.cached7 = arg7
	}

	if allMatch {
		e.state = clean
		return nil
	}

	err := e.fn(
		arg0,
		arg1,
		arg2,
		arg3,
		arg4,
		arg5,
		arg6,
		arg7,
	)
	e.state = clean
	return err
}

func Effect8[T0, T1, T2, T3, T4, T5, T6, T7 comparable](
	rs *ReactiveSystem,
	cell0 Cell,
	cell1 Cell,
	cell2 Cell,
	cell3 Cell,
	cell4 Cell,
	cell5 Cell,
	cell6 Cell,
	cell7 Cell,
	fn func(T0, T1, T2, T3, T4, T5, T6, T7) error,
) (stop func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	e := &SideEffect8[T0, T1, T2, T3, T4, T5, T6, T7]{
		SideEffect: SideEffect{state: dirty, rs: rs},
		cell0:      cell0,
		cell1:      cell1,
		cell2:      cell2,
		cell3:      cell3,
		cell4:      cell4,
		cell5:      cell5,
		cell6:      cell6,
		cell7:      cell7,
		fn:         fn,
	}
	rs.cells = append(rs.cells, e)
	rs.anyDirty = true
	rs.evalAll()

	return func() {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		rs.remove(e)
	}
}
