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
	"gitee.com/matkurban/openimsdk-tools/db/pagination"
	"gitee.com/matkurban/openimsdk-tools/errs"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"server/v3/pkg/common/storage/database"
	"server/v3/pkg/common/storage/model"
)

type MomentLikeMgo struct {
	coll *mongo.Collection
}

func NewMomentLikeMongo(db *mongo.Database) (database.MomentLike, error) {
	coll := db.Collection(database.MomentLikeName)
	_, err := coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "moment_id", Value: 1}, {Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "moment_id", Value: 1}, {Key: "create_time", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "create_time", Value: -1}}},
	})
	if err != nil {
		return nil, err
	}
	return &MomentLikeMgo{coll: coll}, nil
}

func (m *MomentLikeMgo) Create(ctx context.Context, like *model.MomentLike) error {
	if like.CreateTime.IsZero() {
		like.CreateTime = time.Now()
	}
	return mongoutil.InsertOne(ctx, m.coll, like)
}

func (m *MomentLikeMgo) Delete(ctx context.Context, momentID, userID string) error {
	filter := bson.M{"moment_id": momentID, "user_id": userID}
	return mongoutil.DeleteOne(ctx, m.coll, filter)
}

func (m *MomentLikeMgo) Take(ctx context.Context, momentID, userID string) (*model.MomentLike, error) {
	return mongoutil.FindOne[*model.MomentLike](ctx, m.coll, bson.M{"moment_id": momentID, "user_id": userID})
}

func (m *MomentLikeMgo) FindByMoment(ctx context.Context, momentID string, pagination pagination.Pagination) (int64, []*model.MomentLike, error) {
	filter := bson.M{"moment_id": momentID}
	opt := options.Find().SetSort(bson.D{{Key: "create_time", Value: -1}})
	return mongoutil.FindPage[*model.MomentLike](ctx, m.coll, filter, pagination, opt)
}

func (m *MomentLikeMgo) FindByMoments(ctx context.Context, momentIDs []string) ([]*model.MomentLike, error) {
	if len(momentIDs) == 0 {
		return nil, nil
	}
	filter := bson.M{"moment_id": bson.M{"$in": momentIDs}}
	opt := options.Find().SetSort(bson.D{{Key: "create_time", Value: -1}})
	return mongoutil.Find[*model.MomentLike](ctx, m.coll, filter, opt)
}

func (m *MomentLikeMgo) DeleteByMoments(ctx context.Context, momentIDs []string) error {
	if len(momentIDs) == 0 {
		return nil
	}
	_, err := m.coll.DeleteMany(ctx, bson.M{"moment_id": bson.M{"$in": momentIDs}})
	return errs.Wrap(err)
}

func (m *MomentLikeMgo) Count(ctx context.Context, filter any) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}
