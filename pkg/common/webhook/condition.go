package webhook

import (
	"context"
	"server/v3/pkg/common/config"
)

func WithCondition(ctx context.Context, before *config.BeforeConfig, callback func(context.Context) error) error {
	if !before.Enable {
		return nil
	}
	return callback(ctx)
}
