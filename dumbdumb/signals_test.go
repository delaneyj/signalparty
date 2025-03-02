package dumbdumb_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/delaneyj/signalparty/dumbdumb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// from README
func TestBasicUsage(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()
	count := dumbdumb.Signal(rs, 1)
	doubleCount := dumbdumb.Computed1(rs, count, func(c int) int {
		return c * 2
	})

	dumbdumb.Effect1(rs, count, func(c int) error {
		log.Printf("Count is: %d", c)
		return nil
	})

	assert.Equal(t, 2, doubleCount.Value())
	count.SetValue(2)
	assert.Equal(t, 4, doubleCount.Value())
}

// from README
func TestBasicEffect(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()
	count := dumbdumb.Signal(rs, 1)

	callCount := 0
	dumbdumb.Effect1(rs, count, func(c int) error {
		callCount++
		return nil
	})
	// Console: Count in scope: 1
	assert.Equal(t, 1, callCount)
	count.SetValue(2) // Console: Count in scope: 2
	assert.Equal(t, 2, callCount)

	rs.Reset()
	count.SetValue(3) // No console output
	assert.Equal(t, 2, callCount)
}

// should clear subscriptions when untracked by all subscribers
func TestEffectClearSubsWhenUntracked(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()
	a := dumbdumb.Signal(rs, 1)
	b := dumbdumb.Computed1(rs, a, func(a int) int {
		return a * 2
	})
	cRunTimes := 0
	stopEffect := dumbdumb.Effect1(rs, b, func(b int) error {
		cRunTimes++
		return nil
	})

	assert.Equal(t, 1, cRunTimes)
	a.SetValue(2)
	assert.Equal(t, 2, cRunTimes)
	stopEffect()
	a.SetValue(3)
	assert.Equal(t, 2, cRunTimes)
}

func subOne[T int](a T) T {
	return a - 1
}

func sumTwo[T int](a, b T) T {
	return a + b
}

func identity[T any](a T) T {
	return a
}

func joinStrings(a, b string) string {
	return a + " " + b
}

func TestTopologyDropAbaUpdates(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()

	//     A
	//   / |
	//  B  | <- Looks like a flag doesn't it? :D
	//   \ |
	//     C
	//     |
	//     D
	a := dumbdumb.Signal(rs, 2)
	b := dumbdumb.Computed1(rs, a, subOne)
	c := dumbdumb.Computed2(rs, a, b, sumTwo)
	callCount := 0
	d := dumbdumb.Computed1(rs, c, func(c int) string {
		callCount++
		return string(fmt.Sprintf("d: %d", c))
	})

	// Trigger read
	dActual := d.Value()
	assert.Equal(t, "d: 3", dActual)
	assert.Equal(t, 1, callCount)

	a.SetValue(4)
	d.Value()
	assert.Equal(t, 2, callCount)
}

func TestShouldOnlyUpdateEverySignalOnceDiamond(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()

	// In this scenario "D" should only update once when "A" receives
	// an update. This is sometimes referred to as the "diamond" scenario.
	//     A
	//   /   \
	//  B     C
	//   \   /
	//     D

	a := dumbdumb.Signal(rs, "a")
	b := dumbdumb.Computed1[string](rs, a, identity)
	c := dumbdumb.Computed1[string](rs, a, identity)

	callCount := 0
	d := dumbdumb.Computed2(rs, b, c, func(b, c string) string {
		callCount++
		return b + " " + c
	})

	assert.Equal(t, "a a", d.Value())
	assert.Equal(t, 1, callCount)
	callCount = 0

	a.SetValue("aa")
	assert.Equal(t, "aa aa", d.Value())
	assert.Equal(t, 1, callCount)
}

func TestShouldOnlyUpdateEverySignalOnceDiamondTail(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()

	// "E" will be likely updated twice if our mark+sweep logic is buggy.
	//     A
	//   /   \
	//  B     C
	//   \   /
	//     D
	//     |
	//     E

	a := dumbdumb.Signal(rs, "a")
	b := dumbdumb.Computed1[string](rs, a, identity)
	c := dumbdumb.Computed1[string](rs, a, identity)
	d := dumbdumb.Computed2(rs, b, c, joinStrings)

	eCallCount := 0
	e := dumbdumb.Computed1(rs, d, func(d string) string {
		eCallCount++
		return d
	})

	assert.Equal(t, "a a", e.Value())
	assert.Equal(t, 1, eCallCount)

	a.SetValue("aa")
	assert.Equal(t, "aa aa", e.Value())
	assert.Equal(t, 2, eCallCount)
}

func TestBailOutIfResultIsTheSame(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()

	// Bail out if value of "B" never changes
	// A->B->C
	a := dumbdumb.Signal(rs, "a")
	b := dumbdumb.Computed1(rs, a, func(a string) string {
		return "foo"
	})

	callCount := 0
	c := dumbdumb.Computed1(rs, b, func(b string) string {
		callCount++
		return b
	})

	assert.Equal(t, "foo", c.Value())
	assert.Equal(t, 1, callCount)

	a.SetValue("aa")
	assert.Equal(t, "foo", c.Value())
	assert.Equal(t, 1, callCount)
}

func TestShouldOnlyUpdateEverySignalOnceJaggedDiamondTails(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()

	// "F" and "G" will be likely updated twice if our mark+sweep logic is buggy.
	//     A
	//   /   \
	//  B     C
	//  |     |
	//  |     D
	//   \   /
	//     E
	//   /   \
	//  F     G

	a := dumbdumb.Signal(rs, "a")
	b := dumbdumb.Computed1[string](rs, a, identity)
	c := dumbdumb.Computed1[string](rs, a, identity)
	d := dumbdumb.Computed1[string](rs, c, identity)

	eCallCount, eTime := 0, time.Time{}
	e := dumbdumb.Computed2(rs, b, d, func(bV, dV string) string {
		eV := bV + " " + dV
		eCallCount++
		eTime = time.Now()
		return eV
	})

	fCallCount, fTime := 0, time.Time{}
	f := dumbdumb.Computed1(rs, e, func(ev string) string {
		fCallCount++
		fTime = time.Now()
		return ev
	})

	gCallCount, gTime := 0, time.Time{}
	g := dumbdumb.Computed1(rs, e, func(ev string) string {
		gCallCount++
		gTime = time.Now()
		return ev
	})

	require.Equal(t, "a a", f.Value())
	require.Equal(t, 1, fCallCount)
	require.Equal(t, "a a", g.Value())
	require.Equal(t, 1, gCallCount)
	eCallCount, fCallCount, gCallCount = 0, 0, 0

	a.SetValue("b")
	require.Equal(t, "b b", e.Value())
	require.Equal(t, 1, eCallCount)
	require.Equal(t, "b b", f.Value())
	require.Equal(t, 1, fCallCount)
	require.Equal(t, "b b", g.Value())
	require.Equal(t, 1, gCallCount)
	eCallCount, fCallCount, gCallCount = 0, 0, 0

	a.SetValue("c")
	require.Equal(t, "c c", e.Value())
	require.Equal(t, 1, eCallCount)
	require.Equal(t, "c c", f.Value())
	require.Equal(t, 1, fCallCount)
	require.Equal(t, "c c", g.Value())
	require.Equal(t, 1, gCallCount)

	// top to bottom
	assert.True(t, eTime.Before(fTime))
	// left to right
	assert.True(t, fTime.Before(gTime))

}

func TestShouldEnsureSubsUpdate(t *testing.T) {
	// In this scenario "C" always returns the same value. When "A"
	// changes, "B" will update, then "C" at which point its update
	// to "D" will be unmarked. But "D" must still update because
	// "B" marked it. If "D" isn't updated, then we have a bug.
	//     A
	//   /   \
	//  B     *C <- returns same value every time
	//   \   /
	//     D
	rs := dumbdumb.NewReactiveSystem()
	a := dumbdumb.Signal(rs, "a")
	b := dumbdumb.Computed1(rs, a, func(a string) string {
		return a
	})
	c := dumbdumb.Computed1(rs, a, func(a string) string {
		return "c"
	})
	dCallCount := 0
	d := dumbdumb.Computed2(rs, b, c, func(b, c string) string {
		dCallCount++
		return b + " " + c
	})

	assert.Equal(t, "a c", d.Value())
	assert.Equal(t, 1, dCallCount)

	a.SetValue("aa")
	assert.Equal(t, "aa c", d.Value())
}

func TestShouldEnsureSubsUpdateEvenIfTwoDepsUnmarkIt(t *testing.T) {
	// In this scenario both "C" and "D" always return the same
	// value. But "E" must still update because "A" marked it.
	// If "E" isn't updated, then we have a bug.
	//     A
	//   / | \
	//  B *C *D
	//   \ | /
	//     E
	rs := dumbdumb.NewReactiveSystem()
	a := dumbdumb.Signal(rs, "a")
	b := dumbdumb.Computed1[string](rs, a, identity)
	c := dumbdumb.Computed1(rs, a, func(a string) string {
		return "c"
	})
	d := dumbdumb.Computed1(rs, a, func(a string) string {
		return "d"
	})
	eCallCount := 0
	e := dumbdumb.Computed3(rs, b, c, d, func(b, c, d string) string {
		eCallCount++
		return b + " " + c + " " + d
	})

	assert.Equal(t, "a c d", e.Value())
	assert.Equal(t, 1, eCallCount)

	a.SetValue("aa")
	assert.Equal(t, "aa c d", e.Value())
	assert.Equal(t, 2, eCallCount)
}

func TestShouldEnsureSubsUpdateEvenIfAllDepsUnmarkIt(t *testing.T) {
	// In this scenario "B" and "C" always return the same value. When "A"
	// changes, "D" should not update.
	//     A
	//   /   \
	// *B     *C
	//   \   /
	//     D
	rs := dumbdumb.NewReactiveSystem()
	a := dumbdumb.Signal(rs, "a")
	b := dumbdumb.Computed1(rs, a, func(a string) string {
		return "b"
	})
	c := dumbdumb.Computed1(rs, a, func(a string) string {
		return "c"
	})
	dCallCount := 0
	d := dumbdumb.Computed2(rs, b, c, func(b, c string) string {
		dCallCount++
		return b + " " + c
	})

	assert.Equal(t, "b c", d.Value())
	assert.Equal(t, 1, dCallCount)
	dCallCount = 0

	a.SetValue("aa")
	assert.Equal(t, 0, dCallCount)
}

func fail[T any](a T) T {
	panic("fail")
}

func TestShouldKeepGraphConsistentOnActivationErrors(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()

	a := dumbdumb.Signal(rs, 0)

	assert.Panics(t, func() {
		dumbdumb.Computed1[int](rs, a, fail)
	})

	a.SetValue(1)
	assert.Equal(t, 1, a.Value())
}

func TestShouldKeepGraphConsistentOnComputedErrors(t *testing.T) {
	rs := dumbdumb.NewReactiveSystem()

	a := dumbdumb.Signal(rs, 0)

	c := dumbdumb.Computed1[int](rs, a, identity)

	assert.Panics(t, func() {
		dumbdumb.Computed1[int](rs, a, fail)
	})

	a.SetValue(1)
	assert.Equal(t, 1, c.Value())
}
