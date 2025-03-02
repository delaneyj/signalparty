package reactively

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
	t.Run("two signals", func(t *testing.T) {
		rctx := &ReactiveContext{}

		a := Signal(rctx, 7)
		b := Signal(rctx, 1)
		callCount := 0

		c := Memo(rctx, func() int {
			callCount++
			return a.Read() * b.Read()
		})

		assert.Equal(t, 7, c.Read())

		a.Write(2)
		assert.Equal(t, 2, c.Read())

		b.Write(3)
		assert.Equal(t, 6, c.Read())

		assert.Equal(t, 3, callCount)
		c.Read()
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
		a := Signal(rctx, 7)
		b := Signal(rctx, 1)

		callCount1 := 0
		c := Memo(rctx, func() int {
			callCount1++
			return a.Read() * b.Read()
		})

		callCount2 := 0
		d := Memo(rctx, func() int {
			callCount2++
			return c.Read() + 1
		})

		assert.Equal(t, 8, d.Read())
		assert.Equal(t, 1, callCount1)
		assert.Equal(t, 1, callCount2)
		a.Write(3)
		assert.Equal(t, 4, d.Read())
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
		a := Signal(rctx, 7)
		c := Memo(rctx, func() int {
			callCount++
			return a.Read() + 10
		})

		c.Read()
		c.Read()
		assert.Equal(t, 1, callCount)
		a.Write(7)
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
		a := Signal(rctx, 1)
		b := Signal(rctx, 2)
		var callCountA, callCountB, callCountAB int

		cA := Memo(rctx, func() int {
			callCountA++
			return a.Read()
		})

		cB := Memo(rctx, func() int {
			callCountB++
			return b.Read()
		})

		cAB := Memo(rctx, func() int {
			callCountAB++
			if av := cA.Read(); av != 0 {
				return av
			}
			return cB.Read()
		})

		assert.Equal(t, 1, cAB.Read())
		a.Write(2)
		b.Write(3)
		assert.Equal(t, 2, cAB.Read())

		assert.Equal(t, 2, callCountA)
		assert.Equal(t, 2, callCountAB)
		assert.Equal(t, 0, callCountB)
		a.Write(0)
		assert.Equal(t, 3, cAB.Read())
		assert.Equal(t, 3, callCountA)
		assert.Equal(t, 3, callCountAB)
		assert.Equal(t, 1, callCountB)
		b.Write(4)
		assert.Equal(t, 4, cAB.Read())
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
		a := Signal(rctx, 0)
		b := Memo(rctx, func() bool {
			return a.Read() > 0
		})
		callCount := 0

		c := Memo(rctx, func() int {
			callCount++
			if b.Read() {
				return 1
			}
			return 0
		})

		assert.Equal(t, 0, c.Read())
		assert.Equal(t, 1, callCount)

		a.Write(1)
		assert.Equal(t, 1, c.Read())
		assert.Equal(t, 2, callCount)

		a.Write(2)
		assert.Equal(t, 1, c.Read())
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
		s := Signal(rctx, 1)
		a := Memo(rctx, func() int {
			return s.Read()
		})
		b := Memo(rctx, func() int {
			return a.Read() * 2
		})
		c := Memo(rctx, func() int {
			return a.Read() * 3
		})
		callCount := 0
		d := Memo(rctx, func() int {
			callCount++
			return b.Read() + c.Read()
		})

		assert.Equal(t, 5, d.Read())
		assert.Equal(t, 1, callCount)
		s.Write(2)
		assert.Equal(t, 10, d.Read())
		assert.Equal(t, 2, callCount)
		s.Write(3)
		assert.Equal(t, 15, d.Read())
		assert.Equal(t, 3, callCount)

	})

	/*
	   s
	   |
	   l  a (sets s)
	*/
	t.Run("set inside reaction", func(t *testing.T) {
		rctx := &ReactiveContext{}
		s := Signal(rctx, 1)
		a := Memo(rctx, func() bool {
			s.Write(2)
			return true
		})
		l := Memo(rctx, func() int {
			return s.Read() + 100
		})

		a.Read()
		assert.Equal(t, 102, l.Read())
	})
}
