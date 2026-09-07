package group

import (
	"context"

	relationtb "server/v3/pkg/common/storage/model"
)

func (g *groupServer) PopulateGroupMember(ctx context.Context, members ...*relationtb.GroupMember) error {
	return g.notification.PopulateGroupMember(ctx, members...)
}
