package api

import (
	"testing"
	"time"
)

func TestNormalizeAdminPage(t *testing.T) {
	cases := []struct{ in, want PageReq }{
		{PageReq{0, 0}, PageReq{1, 20}},
		{PageReq{-3, -1}, PageReq{1, 20}},
		{PageReq{2, 500}, PageReq{2, 100}},
		{PageReq{3, 10}, PageReq{3, 10}},
	}
	for _, c := range cases {
		if got := normalizeAdminPage(c.in); got != c.want {
			t.Errorf("%v => %v, want %v", c.in, got, c.want)
		}
	}
}

func TestNormalizeMomentIDs(t *testing.T) {
	got, err := normalizeMomentIDs([]string{" a ", "a", "", "b"})
	if err != nil || len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %v %v", got, err)
	}
	if _, err := normalizeMomentIDs(nil); err == nil {
		t.Error("empty should fail")
	}
	if _, err := normalizeMomentIDs([]string{" ", ""}); err == nil {
		t.Error("blank should fail")
	}
	big := make([]string, 101)
	for i := range big {
		big[i] = string(rune('A'+i/26)) + string(rune('a'+i%26))
	}
	if _, err := normalizeMomentIDs(big); err == nil {
		t.Error("101 should fail")
	}
	if _, err := normalizeMomentIDs(big[:100]); err != nil {
		t.Errorf("100 ok: %v", err)
	}
}

func TestShanghaiDayStart(t *testing.T) {
	// 2026-10-16 17:00 UTC == 2026-10-17 01:00 CST -> day start 2026-10-16 16:00 UTC
	now := time.Date(2026, 10, 16, 17, 0, 0, 0, time.UTC)
	want := time.Date(2026, 10, 16, 16, 0, 0, 0, time.UTC)
	if got := shanghaiDayStart(now); !got.Equal(want) {
		t.Errorf("got %v want %v", got.UTC(), want)
	}
	now = time.Date(2026, 10, 16, 15, 59, 0, 0, time.UTC) // 23:59 CST on the 16th
	want = time.Date(2026, 10, 15, 16, 0, 0, 0, time.UTC)
	if got := shanghaiDayStart(now); !got.Equal(want) {
		t.Errorf("got %v want %v", got.UTC(), want)
	}
}
