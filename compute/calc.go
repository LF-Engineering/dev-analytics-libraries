package compute

const (
	// Up label
	Up string = "UP"
	// Down label
	Down string = "DOWN"
	// Flat label
	Flat string = "SAME"
)

// GetStatsComparison returns the comparison of 2 metrics
func GetStatsComparison(currentStats float64, previousStats float64) string {
	if currentStats == previousStats {
		return Flat
	} else if currentStats < previousStats {
		return Down
	} else {
		return Up
	}
}

// CalculatePercent based on numerator & denominator values
func CalculatePercent(numerator float64, denominator float64) float64 {
	if denominator == 0 {
		return 0.0
	}
	return (numerator / denominator) * 100
}

// Subtract num2 from num1
func Subtract(num1 float64, num2 float64) float64 {
	if num2 > num1 {
		return 0.0
	}
	return (num1 - num2)
}

// Ratio of 2 numbers
func Ratio(numerator float64, denominator float64) float64 {
	if denominator == 0 {
		return 0.0
	}
	return (float64(numerator) / float64(denominator))
}

// SumStats sum each element of a given slice and return the result
func SumStats(stats []int64) float64 {
	result := 0.0
	for _, v := range stats {
		result += float64(v)
	}
	return result
}

// AggregateStats sum each element of the 2 different stats and returns the results.
func AggregateStats(stats1 []int64, stats2 []int64) []int64 {
	if len(stats1) > 0 && len(stats2) > 0 && len(stats2) == len(stats1) {
		for i := 0; i < len(stats1); i++ {
			stats1[i] = stats1[i] + stats2[i]
		}
	}

	return stats1
}
