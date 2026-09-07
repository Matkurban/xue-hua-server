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

	"gitee.com/matkurban/openimsdk-tools/db/mongoutil"
	"gitee.com/matkurban/openimsdk-tools/db/pagination"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"server/v3/pkg/common/storage/database"
	"server/v3/pkg/common/storage/model"
)

type FriendTagMgo struct {
	coll *mongo.Collection
}

func NewFriendTagMongo(db *mongo.Database) (database.FriendTag, error) {
	coll := db.Collection(database.FriendTagName)
	_, err := coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tag_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "owner_user_id", Value: 1}, {Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "owner_user_id", Value: 1}, {Key: "create_time", Value: -1}}},
	})
	if err != nil {
		return nil, err
	}
	return &FriendTagMgo{coll: coll}, nil
}

func (f *FriendTagMgo) Create(ctx context.Context, tag *model.FriendTag) error {
	if tag.CreateTime.IsZero() {
		tag.CreateTime = time.Now()
	}
	tag.UpdateTime = tag.CreateTime
	return mongoutil.InsertOne(ctx, f.coll, tag)
}

func (f *FriendTagMgo) UpdateName(ctx context.Context, ownerUserID, tagID, name string) error {
	filter := bson.M{"owner_user_id": ownerUserID, "tag_id": tagID}
	update := bson.M{"$set": bson.M{"name": name, "update_time": time.Now()}}
	return mongoutil.UpdateOne(ctx, f.coll, filter, update, true)
}

func (f *FriendTagMgo) Delete(ctx context.Context, ownerUserID, tagID string) error {
	filter := bson.M{"owner_user_id": ownerUserID, "tag_id": tagID}
	return mongoutil.DeleteOne(ctx, f.coll, filter)
}

func (f *FriendTagMgo) Take(ctx context.Context, ownerUserID, tagID string) (*model.FriendTag, error) {
	return mongoutil.FindOne[*model.FriendTag](ctx, f.coll, bson.M{"owner_user_id": ownerUserID, "tag_id": tagID})
}

func (f *FriendTagMgo) TakeByName(ctx context.Context, ownerUserID, name string) (*model.FriendTag, error) {
	return mongoutil.FindOne[*model.FriendTag](ctx, f.coll, bson.M{"owner_user_id": ownerUserID, "name": name})
}

func (f *FriendTagMgo) FindByIDs(ctx context.Context, ownerUserID string, tagIDs []string) ([]*model.FriendTag, error) {
	filter := bson.M{"owner_user_id": ownerUserID, "tag_id": bson.M{"$in": tagIDs}}
	return mongoutil.Find[*model.FriendTag](ctx, f.coll, filter)
}

func (f *FriendTagMgo) FindOwnerTags(ctx context.Context, ownerUserID string, pagination pagination.Pagination) (int64, []*model.FriendTag, error) {
	filter := bson.M{"owner_user_id": ownerUserID}
	opt := options.Find().SetSort(bson.D{{Key: "create_time", Value: -1}})
	return mongoutil.FindPage[*model.FriendTag](ctx, f.coll, filter, pagination, opt)
}
