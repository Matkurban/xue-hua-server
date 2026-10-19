package mgo

import (
	"testing"
	"time"
)

func TestFillDayKeys(t *testing.T) {
	// 2026-10-01 00:00 +08:00 == 2026-09-30T16:00Z
	start := time.Date(2026, 9, 30, 16, 0, 0, 0, time.UTC)
	end := start.Add(3 * 24 * time.Hour) // exclusive: 10-01..10-03
	got := fillDayKeys(start, end, map[string]int64{"2026-10-02": 5})
	want := map[string]int64{"2026-10-01": 0, "2026-10-02": 5, "2026-10-03": 0}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	if m := fillDayKeys(start, start, nil); len(m) != 0 {
		t.Fatalf("expected empty, got %v", m)
	}
	if m := fillDayKeys(end, start, nil); len(m) != 0 {
		t.Fatalf("expected empty, got %v", m)
	}
	// UTC 2026-10-01T17:00Z is already 10-02 in Beijing.
	s2 := time.Date(2026, 10, 1, 17, 0, 0, 0, time.UTC)
	m := fillDayKeys(s2, s2.Add(time.Hour), nil)
	if _, ok := m["2026-10-02"]; !ok || len(m) != 1 {
		t.Fatalf("tz grouping wrong: %v", m)
	}
}
