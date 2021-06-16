package compute

import (
	"fmt"
	"math"
	"strconv"

	"github.com/labstack/gommon/log"
)

// NearestThousandFormat returns a string with %.2f<K/M/B>
func NearestThousandFormat(count float64) string {
	var suffix = []string{"K", "M", "B", "T", "Qd", "Qt"}
	var formattedStr string
	if count < 1000 {
		return fmt.Sprintf("%d", int64(count))
	}
	exp := (int64)(math.Log(count) / math.Log(1000))
	if exp <= int64(len(suffix)) {
		formattedStr = fmt.Sprintf("%.2f%s", count/math.Pow(1000, float64(exp)), suffix[exp-1])
	}
	return formattedStr
}

// RawNumberFormat returns a string with %.2f<K/M/B>
func RawNumberFormat(metricValue string) int64 {
	if len(metricValue) <= 0 {
		return 0
	}
	var numberVal int64
	if len(metricValue) == 1 {
		numberVal, err := strconv.ParseInt(metricValue, 0, 64)
		if err != nil {
			log.Warnf("Failed converting to raw number format: %+v", err)
		}
		return numberVal
	}
	var metricValueWithoutSuffix, err = strconv.ParseFloat(metricValue[:len(metricValue)-1], 64)
	if err != nil {
		log.Fatalf("Failed converting to raw number format %+v", err)
	}
	switch metricValue[len(metricValue)-1:] {
	case "K":
		numberVal = int64(float64(metricValueWithoutSuffix) * math.Pow(1000, 1.00))
	case "M":
		numberVal = int64(float64(metricValueWithoutSuffix) * math.Pow(1000, 2.00))
	case "B":
		numberVal = int64(float64(metricValueWithoutSuffix) * math.Pow(1000, 3.00))
	case "T":
		numberVal = int64(float64(metricValueWithoutSuffix) * math.Pow(1000, 4.00))
	case "Qd":
		numberVal = int64(float64(metricValueWithoutSuffix) * math.Pow(1000, 5.00))
	case "Qt":
		numberVal = int64(float64(metricValueWithoutSuffix) * math.Pow(1000, 6.00))
	default:
		numberVal, err := strconv.ParseInt(metricValue, 0, 64)
		if err != nil {
			log.Fatalf("Failed converting to raw number format: %+v", err)
		}
		return numberVal
	}

	return numberVal
}
