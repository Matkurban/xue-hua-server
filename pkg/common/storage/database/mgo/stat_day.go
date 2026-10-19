package mgo

import "time"

// statTimezone is the timezone used to group daily statistics (Beijing time, UTC+8).
const statTimezone = "+08:00"

var statLocation = time.FixedZone("CST", 8*3600)

// fillDayKeys returns a map containing every Beijing-time "YYYY-MM-DD" day
// that intersects the half-open range [start, end), with values taken from
// counts (0 if absent). If end <= start it returns an empty map.
func fillDayKeys(start, end time.Time, counts map[string]int64) map[string]int64 {
	res := make(map[string]int64)
	if !end.After(start) {
		return res
	}
	last := end.Add(-time.Millisecond).In(statLocation)
	lastDay := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, statLocation)
	s := start.In(statLocation)
	for d := time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, statLocation); !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		key := d.Format(time.DateOnly)
		res[key] = counts[key]
	}
	return res
}
