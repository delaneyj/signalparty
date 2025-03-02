package alien_test

import (
	"testing"

	"github.com/delaneyj/signalparty/alien"
	"github.com/stretchr/testify/assert"
)

// should pause tracking
func TestShouldPauseTracking(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
		t.FailNow()
	})

	src := alien.Signal(rs, 0)
	c := alien.Computed(rs, func(oldValue int) int {
		rs.PauseTracking()
		value := src.Value()
		rs.ResumeTracking()
		return value
	})
	actualC := c.Value()
	assert.Equal(t, 0, actualC)

	src.SetValue(1)
	actualC = c.Value()
	assert.Equal(t, 0, actualC)
}
