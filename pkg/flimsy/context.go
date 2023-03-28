package flimsy

type Context[T any] struct {
	runtime *ReactiveContext
	// Unique identifier for the context
	id int64
	// Default value for the context
	defaultValue T
}

func (c *Context[T]) Read() T {
	// If the getter finds null or undefined as the value then the default value is returned instead
	x, ok := c.runtime.observer.get(c.id)
	if !ok {
		return c.defaultValue
	}

	t, ok := x.(T)
	if !ok {
		panic("invalid type")
	}

	return t
}

func (c *Context[T]) Write(value T) {
	c.runtime.observer.set(c.id, value)
}
