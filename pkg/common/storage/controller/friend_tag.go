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

package controller

import (
	"context"
	"time"

	"server/v3/pkg/common/storage/pagination"
	"server/v3/pkg/common/errs"
	"server/v3/pkg/util/idutil"
	"server/v3/pkg/common/storage/database"
	"server/v3/pkg/common/storage/database/mgo"
	"server/v3/pkg/common/storage/model"
)

type FriendTagDatabase interface {
	CreateTag(ctx context.Context, ownerUserID, name string) (*model.FriendTag, error)
	RenameTag(ctx context.Context, ownerUserID, tagID, name string) error
	DeleteTag(ctx context.Context, ownerUserID, tagID string) error
	AddMembers(ctx context.Context, ownerUserID, tagID string, friendUserIDs []string) error
	RemoveMembers(ctx context.Context, ownerUserID, tagID string, friendUserIDs []string) error
	ListTags(ctx context.Context, ownerUserID string, pagination pagination.Pagination) (int64, []*model.FriendTag, error)
	ListTagMembers(ctx context.Context, ownerUserID, tagID string, pagination pagination.Pagination) (int64, []*model.FriendTagMember, error)
	FindTagIDsByFriend(ctx context.Context, ownerUserID, friendUserID string) ([]string, error)
	FindMembersByOwner(ctx context.Context, ownerUserID string, tagIDs []string) ([]*model.FriendTagMember, error)
	FindMembersByFriend(ctx context.Context, friendUserID string) ([]*model.FriendTagMember, error)
}

type friendTagDatabase struct {
	tagDB    database.FriendTag
	memberDB database.FriendTagMember
	friendDB database.Friend
}

func NewFriendTagDatabase(tagDB database.FriendTag, memberDB database.FriendTagMember, friendDB database.Friend) FriendTagDatabase {
	return &friendTagDatabase{tagDB: tagDB, memberDB: memberDB, friendDB: friendDB}
}

func (f *friendTagDatabase) CreateTag(ctx context.Context, ownerUserID, name string) (*model.FriendTag, error) {
	if ownerUserID == "" || name == "" {
		return nil, errs.ErrArgs.WrapMsg("ownerUserID or name is empty")
	}
	if _, err := f.tagDB.TakeByName(ctx, ownerUserID, name); err == nil {
		return nil, errs.ErrArgs.WrapMsg("tag name already exists")
	} else if !mgo.IsNotFound(err) {
		return nil, err
	}
	m := &model.FriendTag{
		TagID:       idutil.OperationIDGenerator(),
		OwnerUserID: ownerUserID,
		Name:        name,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	if err := f.tagDB.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (f *friendTagDatabase) RenameTag(ctx context.Context, ownerUserID, tagID, name string) error {
	if ownerUserID == "" || tagID == "" || name == "" {
		return errs.ErrArgs.WrapMsg("ownerUserID/tagID/name is empty")
	}
	return f.tagDB.UpdateName(ctx, ownerUserID, tagID, name)
}

func (f *friendTagDatabase) DeleteTag(ctx context.Context, ownerUserID, tagID string) error {
	if ownerUserID == "" || tagID == "" {
		return errs.ErrArgs.WrapMsg("ownerUserID/tagID is empty")
	}
	if err := f.tagDB.Delete(ctx, ownerUserID, tagID); err != nil {
		return err
	}
	return f.memberDB.RemoveByTag(ctx, ownerUserID, tagID)
}

func (f *friendTagDatabase) AddMembers(ctx context.Context, ownerUserID, tagID string, friendUserIDs []string) error {
	if ownerUserID == "" || tagID == "" || len(friendUserIDs) == 0 {
		return errs.ErrArgs.WrapMsg("ownerUserID/tagID/friendUserIDs is empty")
	}
	friends, err := f.friendDB.FindFriends(ctx, ownerUserID, friendUserIDs)
	if err != nil {
		return err
	}
	if len(friends) != len(friendUserIDs) {
		return errs.ErrArgs.WrapMsg("some friend user ids are invalid")
	}
	members := make([]*model.FriendTagMember, 0, len(friendUserIDs))
	for _, id := range friendUserIDs {
		members = append(members, &model.FriendTagMember{
			TagID:        tagID,
			OwnerUserID:  ownerUserID,
			FriendUserID: id,
			CreateTime:   time.Now(),
		})
	}
	return f.memberDB.AddMembers(ctx, members)
}

func (f *friendTagDatabase) RemoveMembers(ctx context.Context, ownerUserID, tagID string, friendUserIDs []string) error {
	if ownerUserID == "" || tagID == "" || len(friendUserIDs) == 0 {
		return errs.ErrArgs.WrapMsg("ownerUserID/tagID/friendUserIDs is empty")
	}
	return f.memberDB.RemoveMembers(ctx, ownerUserID, tagID, friendUserIDs)
}

func (f *friendTagDatabase) ListTags(ctx context.Context, ownerUserID string, pagination pagination.Pagination) (int64, []*model.FriendTag, error) {
	if ownerUserID == "" {
		return 0, nil, errs.ErrArgs.WrapMsg("ownerUserID is empty")
	}
	return f.tagDB.FindOwnerTags(ctx, ownerUserID, pagination)
}

func (f *friendTagDatabase) ListTagMembers(ctx context.Context, ownerUserID, tagID string, pagination pagination.Pagination) (int64, []*model.FriendTagMember, error) {
	if ownerUserID == "" || tagID == "" {
		return 0, nil, errs.ErrArgs.WrapMsg("ownerUserID/tagID is empty")
	}
	return f.memberDB.FindByTag(ctx, ownerUserID, tagID, pagination)
}

func (f *friendTagDatabase) FindTagIDsByFriend(ctx context.Context, ownerUserID, friendUserID string) ([]string, error) {
	return f.memberDB.FindTagIDsByFriend(ctx, ownerUserID, friendUserID)
}

func (f *friendTagDatabase) FindMembersByOwner(ctx context.Context, ownerUserID string, tagIDs []string) ([]*model.FriendTagMember, error) {
	return f.memberDB.FindMembersByOwner(ctx, ownerUserID, tagIDs)
}

func (f *friendTagDatabase) FindMembersByFriend(ctx context.Context, friendUserID string) ([]*model.FriendTagMember, error) {
	return f.memberDB.FindByFriend(ctx, friendUserID)
}
