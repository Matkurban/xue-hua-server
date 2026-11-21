package statistics

import (
	"context"
	"time"

	"server/v3/pkg/common/log"
)

type Statistics struct {
	AllCount   *uint64
	ModuleName string
	PrintArgs  string
	SleepTime  uint64
}

func (s *Statistics) output() {
	var intervalCount uint64
	t := time.NewTicker(time.Duration(s.SleepTime) * time.Second)
	defer t.Stop()
	var sum uint64
	var timeIntervalNum uint64
	for {
		sum = *s.AllCount
		<-t.C
		if *s.AllCount-sum <= 0 {
			intervalCount = 0
		} else {
			intervalCount = *s.AllCount - sum
		}
		timeIntervalNum++
		log.ZWarn(
			context.Background(),
			" system stat ",
			nil,
			"args",
			s.PrintArgs,
			"intervalCount",
			intervalCount,
			"total:",
			*s.AllCount,
			"intervalNum",
			timeIntervalNum,
			"avg",
			(*s.AllCount)/(timeIntervalNum)/s.SleepTime,
		)
	}
}
