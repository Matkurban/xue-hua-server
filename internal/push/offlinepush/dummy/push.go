package dummy

import (
	"context"
	"server/v3/pkg/common/log"
	"server/v3/internal/push/offlinepush/options"
	"sync/atomic"
)

func NewClient() *Dummy {
	return &Dummy{}
}

type Dummy struct {
	v atomic.Bool
}

func (d *Dummy) Push(ctx context.Context, userIDs []string, title, content string, opts *options.Opts) error {
	if d.v.CompareAndSwap(false, true) {
		log.ZWarn(ctx, "dummy push", nil, "ps", "the offline push is not configured. to configure it, please go to config/server-push.yml")
	}
	return nil
}
