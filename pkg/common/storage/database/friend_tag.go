// Copyright © 2026 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package database

import (
	"context"

	"gitee.com/matkurban/openimsdk-tools/db/pagination"
	"server/v3/pkg/common/storage/model"
)

type FriendTag interface {
	Create(ctx context.Context, tag *model.FriendTag) error
	UpdateName(ctx context.Context, ownerUserID, tagID, name string) error
	Delete(ctx context.Context, ownerUserID, tagID string) error
	Take(ctx context.Context, ownerUserID, tagID string) (*model.FriendTag, error)
	TakeByName(ctx context.Context, ownerUserID, name string) (*model.FriendTag, error)
	FindByIDs(ctx context.Context, ownerUserID string, tagIDs []string) ([]*model.FriendTag, error)
	FindOwnerTags(ctx context.Context, ownerUserID string, pagination pagination.Pagination) (int64, []*model.FriendTag, error)
}

type FriendTagMember interface {
	AddMembers(ctx context.Context, members []*model.FriendTagMember) error
	RemoveMembers(ctx context.Context, ownerUserID, tagID string, friendUserIDs []string) error
	RemoveByTag(ctx context.Context, ownerUserID, tagID string) error
	FindByTag(ctx context.Context, ownerUserID, tagID string, pagination pagination.Pagination) (int64, []*model.FriendTagMember, error)
	FindByFriend(ctx context.Context, friendUserID string) ([]*model.FriendTagMember, error)
	FindTagIDsByFriend(ctx context.Context, ownerUserID, friendUserID string) ([]string, error)
	FindMembersByOwner(ctx context.Context, ownerUserID string, tagIDs []string) ([]*model.FriendTagMember, error)
}
