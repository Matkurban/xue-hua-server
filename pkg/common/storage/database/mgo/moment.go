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
	"server/v3/pkg/common/errs"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"server/v3/pkg/common/storage/database"
	"server/v3/pkg/common/storage/model"
)

type MomentMgo struct {
	coll *mongo.Collection
}

func NewMomentMongo(db *mongo.Database) (database.Moment, error) {
	coll := db.Collection(database.MomentName)
	_, err := coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "moment_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "create_time", Value: -1}}},
		{Keys: bson.D{{Key: "visible_type", Value: 1}, {Key: "create_time", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "create_time", Value: -1}}},
		{Keys: bson.D{{Key: "create_time", Value: -1}}},
		{Keys: bson.D{{Key: "visible_group_ids", Value: 1}}},
	})
	if err != nil {
		return nil, err
	}
	return &MomentMgo{coll: coll}, nil
}

func (m *MomentMgo) Create(ctx context.Context, moment *model.Moment) error {
	if moment.CreateTime.IsZero() {
		moment.CreateTime = time.Now()
	}
	moment.UpdateTime = moment.CreateTime
	return mongoutil.InsertOne(ctx, m.coll, moment)
}

func (m *MomentMgo) UpdateByMap(ctx context.Context, momentID, userID string, args map[string]any) error {
	if len(args) == 0 {
		return nil
	}
	args["update_time"] = time.Now()
	filter := bson.M{"moment_id": momentID, "user_id": userID}
	return mongoutil.UpdateOne(ctx, m.coll, filter, bson.M{"$set": args}, true)
}

func (m *MomentMgo) Delete(ctx context.Context, momentID, userID string) error {
	filter := bson.M{"moment_id": momentID, "user_id": userID}
	return mongoutil.UpdateOne(ctx, m.coll, filter, bson.M{"$set": bson.M{"status": int32(1), "update_time": time.Now()}}, true)
}

func (m *MomentMgo) Take(ctx context.Context, momentID string) (*model.Moment, error) {
	return mongoutil.FindOne[*model.Moment](ctx, m.coll, bson.M{"moment_id": momentID})
}

func (m *MomentMgo) FindPage(ctx context.Context, filter any, pagination pagination.Pagination) (int64, []*model.Moment, error) {
	opt := options.Find().SetSort(bson.D{{Key: "create_time", Value: -1}})
	return mongoutil.FindPage[*model.Moment](ctx, m.coll, filter, pagination, opt)
}

func (m *MomentMgo) IncCounts(ctx context.Context, momentID string, likeDelta, commentDelta int64) error {
	set := bson.M{"update_time": time.Now()}
	inc := bson.M{}
	if likeDelta != 0 {
		inc["like_count"] = likeDelta
	}
	if commentDelta != 0 {
		inc["comment_count"] = commentDelta
	}
	update := bson.M{"$set": set}
	if len(inc) > 0 {
		update["$inc"] = inc
	}
	return mongoutil.UpdateOne(ctx, m.coll, bson.M{"moment_id": momentID}, update, true)
}

func (m *MomentMgo) Count(ctx context.Context, filter any) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}

// SetStatus sets the status of the given moments, ignoring user-deleted ones (status 1).
func (m *MomentMgo) SetStatus(ctx context.Context, momentIDs []string, status int32) (int64, error) {
	if len(momentIDs) == 0 {
		return 0, nil
	}
	filter := bson.M{"moment_id": bson.M{"$in": momentIDs}, "status": bson.M{"$ne": int32(1)}}
	res, err := m.coll.UpdateMany(ctx, filter, bson.M{"$set": bson.M{"status": status, "update_time": time.Now()}})
	if err != nil {
		return 0, errs.Wrap(err)
	}
	return res.MatchedCount, nil
}

// DeleteByIDs physically removes the moments.
func (m *MomentMgo) DeleteByIDs(ctx context.Context, momentIDs []string) (int64, error) {
	if len(momentIDs) == 0 {
		return 0, nil
	}
	res, err := m.coll.DeleteMany(ctx, bson.M{"moment_id": bson.M{"$in": momentIDs}})
	if err != nil {
		return 0, errs.Wrap(err)
	}
	return res.DeletedCount, nil
}

func (m *MomentMgo) FindByIDs(ctx context.Context, momentIDs []string) ([]*model.Moment, error) {
	if len(momentIDs) == 0 {
		return nil, nil
	}
	return mongoutil.Find[*model.Moment](ctx, m.coll, bson.M{"moment_id": bson.M{"$in": momentIDs}})
}
