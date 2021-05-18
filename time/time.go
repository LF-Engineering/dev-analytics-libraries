package time

import (
	"math"
	"time"
)

// ConvertTimeToFloat ...
func ConvertTimeToFloat(t time.Time) float64 {
	return math.Round(float64(t.UnixNano())/float64(time.Second)*1e6) / 1e6
}

// GetDaysBetweenDates calculate days between two dates
func GetDaysBetweenDates(t1 time.Time, t2 time.Time) float64 {
	res := t1.Sub(t2).Hours() / 24
	return res
}

// GetOldestDate get the older date between two nullable dates
func GetOldestDate(t1 *time.Time, t2 *time.Time) *time.Time {
	from, err := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	if err != nil {
		return nil
	}

	isT1Empty := t1 == nil || t1.IsZero()
	isT2Empty := t2 == nil || t2.IsZero()

	if isT1Empty && !isT2Empty {
		from = *t2
	} else if !isT1Empty && isT2Empty {
		from = *t1
	} else if !isT1Empty && !isT2Empty {
		from = *t2
		if t1.Before(*t2) {
			from = *t1
		}
	}

	return &from
}

// GetTimeStampValues returns formatted string of
// 	from & to timeframes
func GetTimeStampValues(val string) (from, to int64) {
	switch val {
	case "this month":
		from = formatTimestampEpochMillis(beginningOfMonth())
		to = formatTimestampEpochMillis(endOfMonth())
	case "this year":
		from = formatTimestampEpochMillis(beginningOfYear())
		to = formatTimestampEpochMillis(endOfYear())
	case "month to date":
		from = formatTimestampEpochMillis(beginningOfMonth())
		to = formatTimestampEpochMillis(today())
	case "year to date":
		from = formatTimestampEpochMillis(beginningOfYear())
		to = formatTimestampEpochMillis(today())
	case "last 7 days":
		from = formatTimestampEpochMillis(lastNumberOfDays(7))
		to = formatTimestampEpochMillis(today())
	case "last 30 days":
		from = formatTimestampEpochMillis(lastNumberOfDays(30))
		to = formatTimestampEpochMillis(today())
	case "last 60 days":
		from = formatTimestampEpochMillis(lastNumberOfDays(60))
		to = formatTimestampEpochMillis(today())
	case "last 90 days":
		from = formatTimestampEpochMillis(lastNumberOfDays(90))
		to = formatTimestampEpochMillis(today())
	case "last 6 months":
		from = formatTimestampEpochMillis(lastNumberOfDays(180))
		to = formatTimestampEpochMillis(today())
	case "last 1 year":
		from = formatTimestampEpochMillis(lastNumberOfYears(1))
		to = formatTimestampEpochMillis(today())
	case "last 2 years":
		from = formatTimestampEpochMillis(lastNumberOfYears(2))
		to = formatTimestampEpochMillis(today())
	case "last 3 years":
		from = formatTimestampEpochMillis(lastNumberOfYears(3))
		to = formatTimestampEpochMillis(today())
	case "last 5 years":
		from = formatTimestampEpochMillis(lastNumberOfYears(5))
		to = formatTimestampEpochMillis(today())
	case "all-time":
		from = formatTimestampEpochMillis(lastNumberOfYears(100))
		if from < 0 {
			fromFloat := math.Abs(float64(from))
			from = int64(fromFloat)
		}
		to = formatTimestampEpochMillis(today())
	default:
		from = formatTimestampEpochMillis(beginningOfMonth())
		to = formatTimestampEpochMillis(endOfMonth())
	}
	return
}

// beginningOfMonth returns the dateTime value of the first day of the current month
func beginningOfMonth() time.Time {
	now := time.Now()
	y, m, _ := now.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, now.Location())
}

// endOfMonth returns the dateTime value of the last day of the current month
func endOfMonth() time.Time {
	return beginningOfMonth().AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// today returns the datetime value of the current day
func today() time.Time {
	now := time.Now()
	y, m, d := now.Date()
	return time.Date(y, m, d, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
}

// beginningOfYear returns the dateTime value of the first day of the current year
func beginningOfYear() time.Time {
	now := time.Now()
	y, _, _ := now.Date()
	return time.Date(y, time.January, 1, 0, 0, 0, 0, now.Location())
}

// endOfYear returns the dateTime value of the last day of the current year
func endOfYear() time.Time {
	return beginningOfYear().AddDate(1, 0, 0).Add(-time.Nanosecond)
}

// lastNumberOfDays returns the dateTime value of the current day - passed number of days value
func lastNumberOfDays(days int) time.Time {
	now := time.Now()
	y, m, d := now.Date()
	return time.Date(y, m, d-days, 0, 0, 0, 0, now.Location())
}

// lastNumberOfYears returns the dateTime value of the current year - passed number of years value
func lastNumberOfYears(years int) time.Time {
	now := time.Now()
	return now.AddDate(-years, 0, 0)
}

// formatTimestampString returns a formatted RFC 33339 Datetime string
func formatTimestampString(t time.Time) string {
	layout := "2006-01-02T15:04:05.000Z"
	return t.Format(layout)
}

// formatTimestampEpochMillis returns epoch millis
func formatTimestampEpochMillis(t time.Time) int64 {
	return t.UnixNano() / 1000000
}
