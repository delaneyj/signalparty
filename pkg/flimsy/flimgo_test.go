package flimsy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCore(t *testing.T) {
	/*
	   a  b
	   | /
	   c
	*/
	t.Run("two CreateSignals", func(t *testing.T) {
		rctx := &ReactiveContext{}

		a, setA := CreateSignal(rctx, 7)
		b, setB := CreateSignal(rctx, 1)
		callCount := 0

		c, err := CreateMemo(rctx, func() (int, error) {
			callCount++
			return a() * b(), nil
		})
		assert.Nil(t, err)

		assert.Equal(t, 7, c())

		setA(2)
		assert.Equal(t, 2, c())

		setB(3)
		assert.Equal(t, 6, c())

		assert.Equal(t, 3, callCount)
		c()
		assert.Equal(t, 3, callCount)
	})

	/*
	   a  b
	   | /
	   c
	   |
	   d
	*/
	t.Run("dependent computed", func(t *testing.T) {
		rctx := &ReactiveContext{}
		a, setA := CreateSignal(rctx, 7)
		b, _ := CreateSignal(rctx, 1)

		callCount1 := 0
		c, err := CreateMemo(rctx, func() (int, error) {
			callCount1++
			return a() * b(), nil
		})
		assert.Nil(t, err)

		callCount2 := 0
		d, err := CreateMemo(rctx, func() (int, error) {
			callCount2++
			return c() + 1, nil
		})
		assert.Nil(t, err)

		assert.Equal(t, 8, d())
		assert.Equal(t, 1, callCount1)
		assert.Equal(t, 1, callCount2)
		setA(3)
		assert.Equal(t, 4, d())
		assert.Equal(t, 2, callCount1)
		assert.Equal(t, 2, callCount2)
	})

	/*
		a
		|
		c
	*/
	t.Run("equality check", func(t *testing.T) {
		callCount := 0
		rctx := &ReactiveContext{}
		a, setA := CreateSignal(rctx, 7)
		c, err := CreateMemo(rctx, func() (int, error) {
			callCount++
			return a() + 10, nil
		})
		assert.Nil(t, err)

		c()
		c()
		assert.Equal(t, 1, callCount)
		setA(7)
		assert.Equal(t, 1, callCount) // unchanged, equality check
	})

	/*
	   a     b
	   |     |
	   cA   cB
	   |   / (dynamically depends on cB)
	   cAB
	*/
	t.Run("dynamic computed", func(t *testing.T) {
		rctx := &ReactiveContext{}
		a, setA := CreateSignal(rctx, 1)
		b, setB := CreateSignal(rctx, 2)
		var callCountA, callCountB, callCountAB int

		cA, err := CreateMemo(rctx, func() (int, error) {
			callCountA++
			return a(), nil
		})
		assert.Nil(t, err)

		cB, err := CreateMemo(rctx, func() (int, error) {
			callCountB++
			return b(), nil
		})
		assert.Nil(t, err)

		cAB, err := CreateMemo(rctx, func() (int, error) {
			callCountAB++
			if av := cA(); av != 0 {
				return av, nil
			}
			return cB(), nil
		})
		assert.Nil(t, err)

		assert.Equal(t, 1, cAB())
		setA(2)
		setB(3)
		assert.Equal(t, 2, cAB())

		assert.Equal(t, 2, callCountA)
		assert.Equal(t, 2, callCountAB)
		assert.Equal(t, 0, callCountB)
		setA(0)
		assert.Equal(t, 3, cAB())
		assert.Equal(t, 3, callCountA)
		assert.Equal(t, 3, callCountAB)
		assert.Equal(t, 1, callCountB)
		setB(4)
		assert.Equal(t, 4, cAB())
		assert.Equal(t, 3, callCountA)
		assert.Equal(t, 4, callCountAB)
		assert.Equal(t, 2, callCountB)
	})

	// TBD test cleanup in api

	/*
	   a
	   |
	   b (=)
	   |
	   c
	*/
	t.Run("boolean equality check", func(t *testing.T) {
		rctx := &ReactiveContext{}
		a, setA := CreateSignal(rctx, 0)
		b, err := CreateMemo(rctx, func() (bool, error) {
			return a() > 0, nil
		})
		callCount := 0
		assert.Nil(t, err)

		c, err := CreateMemo(rctx, func() (int, error) {
			callCount++
			if b() {
				return 1, nil
			}
			return 0, nil
		})
		assert.Nil(t, err)

		assert.Equal(t, 0, c())
		assert.Equal(t, 1, callCount)

		setA(1)
		assert.Equal(t, 1, c())
		assert.Equal(t, 2, callCount)

		setA(2)
		assert.Equal(t, 1, c())
		assert.Equal(t, 2, callCount) // unchanged, oughtn't run because bool didn't change
	})

	/*
	   s
	   |
	   a
	   | \
	   b  c
	    \ |
	      d
	*/
	t.Run("diamond computeds", func(t *testing.T) {
		rctx := &ReactiveContext{}
		s, setS := CreateSignal(rctx, 1)
		a, err := CreateMemo(rctx, func() (int, error) {
			return s(), nil
		})
		assert.Nil(t, err)

		b, err := CreateMemo(rctx, func() (int, error) {
			return a() * 2, nil
		})
		assert.Nil(t, err)

		c, err := CreateMemo(rctx, func() (int, error) {
			return a() * 3, nil
		})
		assert.Nil(t, err)

		callCount := 0
		d, err := CreateMemo(rctx, func() (int, error) {
			callCount++
			return b() + c(), nil
		})
		assert.Nil(t, err)

		assert.Equal(t, 5, d())
		assert.Equal(t, 1, callCount)
		setS(2)
		assert.Equal(t, 10, d())
		assert.Equal(t, 2, callCount)
		setS(3)
		assert.Equal(t, 15, d())
		assert.Equal(t, 3, callCount)

	})

	/*
	   s
	   |
	   l  a (sets s)
	*/
	t.Run("set inside reaction", func(t *testing.T) {
		rctx := &ReactiveContext{}
		s, setS := CreateSignal(rctx, 1)
		a, err := CreateMemo(rctx, func() (bool, error) {
			setS(2)
			return true, nil
		})
		assert.Nil(t, err)

		l, err := CreateMemo(rctx, func() (int, error) {
			return s() + 100, nil
		})
		assert.Nil(t, err)

		a()
		assert.Equal(t, 102, l())
	})
}
