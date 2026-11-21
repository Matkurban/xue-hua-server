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

package mgo

import (
	"context"
	"time"

	"server/v3/pkg/common/storage/mongoutil"
	"server/v3/pkg/common/storage/pagination"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"server/v3/pkg/common/storage/database"
	"server/v3/pkg/common/storage/model"
)

type FriendTagMemberMgo struct {
	coll *mongo.Collection
}

func NewFriendTagMemberMongo(db *mongo.Database) (database.FriendTagMember, error) {
	coll := db.Collection(database.FriendTagMemberName)
	_, err := coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "owner_user_id", Value: 1}, {Key: "tag_id", Value: 1}, {Key: "friend_user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "owner_user_id", Value: 1}, {Key: "tag_id", Value: 1}}},
		{Keys: bson.D{{Key: "owner_user_id", Value: 1}, {Key: "friend_user_id", Value: 1}}},
	})
	if err != nil {
		return nil, err
	}
	return &FriendTagMemberMgo{coll: coll}, nil
}

func (f *FriendTagMemberMgo) AddMembers(ctx context.Context, members []*model.FriendTagMember) error {
	for i := range members {
		if members[i].CreateTime.IsZero() {
			members[i].CreateTime = time.Now()
		}
	}
	return mongoutil.InsertMany(ctx, f.coll, members)
}

func (f *FriendTagMemberMgo) RemoveMembers(ctx context.Context, ownerUserID, tagID string, friendUserIDs []string) error {
	filter := bson.M{"owner_user_id": ownerUserID, "tag_id": tagID, "friend_user_id": bson.M{"$in": friendUserIDs}}
	return mongoutil.DeleteMany(ctx, f.coll, filter)
}

func (f *FriendTagMemberMgo) RemoveByTag(ctx context.Context, ownerUserID, tagID string) error {
	filter := bson.M{"owner_user_id": ownerUserID, "tag_id": tagID}
	return mongoutil.DeleteMany(ctx, f.coll, filter)
}

func (f *FriendTagMemberMgo) FindByTag(ctx context.Context, ownerUserID, tagID string, pagination pagination.Pagination) (int64, []*model.FriendTagMember, error) {
	filter := bson.M{"owner_user_id": ownerUserID, "tag_id": tagID}
	opt := options.Find().SetSort(bson.D{{Key: "create_time", Value: -1}})
	return mongoutil.FindPage[*model.FriendTagMember](ctx, f.coll, filter, pagination, opt)
}

func (f *FriendTagMemberMgo) FindTagIDsByFriend(ctx context.Context, ownerUserID, friendUserID string) ([]string, error) {
	filter := bson.M{"owner_user_id": ownerUserID, "friend_user_id": friendUserID}
	return mongoutil.Find[string](ctx, f.coll, filter, options.Find().SetProjection(bson.M{"_id": 0, "tag_id": 1}))
}

func (f *FriendTagMemberMgo) FindByFriend(ctx context.Context, friendUserID string) ([]*model.FriendTagMember, error) {
	filter := bson.M{"friend_user_id": friendUserID}
	return mongoutil.Find[*model.FriendTagMember](ctx, f.coll, filter)
}

func (f *FriendTagMemberMgo) FindMembersByOwner(ctx context.Context, ownerUserID string, tagIDs []string) ([]*model.FriendTagMember, error) {
	filter := bson.M{"owner_user_id": ownerUserID, "tag_id": bson.M{"$in": tagIDs}}
	return mongoutil.Find[*model.FriendTagMember](ctx, f.coll, filter)
}
