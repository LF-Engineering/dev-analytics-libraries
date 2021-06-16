package compute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Check the subtract util method returns expected computed value
func TestSubtract(t *testing.T) {
	result := Subtract(540.00, 230.00)
	assert.Equal(t, result, 310.00)

	result = Subtract(140.00, 230.00)
	assert.Equal(t, result, 0.00)

	result = Subtract(340.00, 230.00)
	assert.NotEqual(t, result, 111.00)

}

// Check the calculate percent util method returns expected computed value
func TestCalculatePercent(t *testing.T) {
	result := CalculatePercent(500.00, 1000.00)
	assert.Equal(t, result, 50.00)

	result = CalculatePercent(500.00, 0.00)
	assert.Equal(t, result, 0.00)

}

// Check the ratio util method returns expected computed value
func TestRatio(t *testing.T) {
	result := Ratio(500.00, 1000.00)
	assert.Equal(t, result, 0.50)

	result = Ratio(750.00, 1000.00)
	assert.Equal(t, result, 0.75)

	result = Ratio(1500.00, 1000.00)
	assert.Equal(t, result, 1.5)

	result = Ratio(500.00, 0.00)
	assert.Equal(t, result, 0.00)

}

// Check the stats comparison util method returns expected trend for different values
func TestStatsComparison(t *testing.T) {
	assert.Equal(t, Up, "UP")
	assert.Equal(t, Down, "DOWN")
	assert.Equal(t, Flat, "SAME")

	result := GetStatsComparison(1200.00, 1000.00)
	assert.Equal(t, result, Up)

	result = GetStatsComparison(750.00, 1000.00)
	assert.Equal(t, result, Down)

	result = GetStatsComparison(1500.00, 1500.00)
	assert.Equal(t, Flat, result)

}
