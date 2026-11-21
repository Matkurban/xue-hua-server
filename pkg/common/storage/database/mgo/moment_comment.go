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

type MomentCommentMgo struct {
	coll *mongo.Collection
}

func NewMomentCommentMongo(db *mongo.Database) (database.MomentComment, error) {
	coll := db.Collection(database.MomentCommentName)
	_, err := coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "comment_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "moment_id", Value: 1}, {Key: "create_time", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "create_time", Value: -1}}},
	})
	if err != nil {
		return nil, err
	}
	return &MomentCommentMgo{coll: coll}, nil
}

func (m *MomentCommentMgo) Create(ctx context.Context, comment *model.MomentComment) error {
	if comment.CreateTime.IsZero() {
		comment.CreateTime = time.Now()
	}
	comment.UpdateTime = comment.CreateTime
	return mongoutil.InsertOne(ctx, m.coll, comment)
}

func (m *MomentCommentMgo) Delete(ctx context.Context, commentID, userID string) error {
	filter := bson.M{"comment_id": commentID, "user_id": userID}
	return mongoutil.UpdateOne(ctx, m.coll, filter, bson.M{"$set": bson.M{"status": int32(1), "update_time": time.Now()}}, true)
}

func (m *MomentCommentMgo) Take(ctx context.Context, commentID string) (*model.MomentComment, error) {
	return mongoutil.FindOne[*model.MomentComment](ctx, m.coll, bson.M{"comment_id": commentID})
}

func (m *MomentCommentMgo) FindByMoment(ctx context.Context, momentID string, pagination pagination.Pagination) (int64, []*model.MomentComment, error) {
	filter := bson.M{"moment_id": momentID, "status": bson.M{"$ne": int32(1)}}
	opt := options.Find().SetSort(bson.D{{Key: "create_time", Value: 1}})
	return mongoutil.FindPage[*model.MomentComment](ctx, m.coll, filter, pagination, opt)
}

func (m *MomentCommentMgo) FindByMoments(ctx context.Context, momentIDs []string) ([]*model.MomentComment, error) {
	if len(momentIDs) == 0 {
		return nil, nil
	}
	filter := bson.M{"moment_id": bson.M{"$in": momentIDs}, "status": bson.M{"$ne": int32(1)}}
	opt := options.Find().SetSort(bson.D{{Key: "create_time", Value: 1}})
	return mongoutil.Find[*model.MomentComment](ctx, m.coll, filter, opt)
}

// AdminDelete soft-deletes a comment regardless of its author. It reports whether
// the comment transitioned from visible to deleted (so counters must be decremented).
func (m *MomentCommentMgo) AdminDelete(ctx context.Context, commentID string) (bool, error) {
	filter := bson.M{"comment_id": commentID, "status": bson.M{"$ne": int32(1)}}
	res, err := m.coll.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"status": int32(1), "update_time": time.Now()}})
	if err != nil {
		return false, errs.Wrap(err)
	}
	return res.MatchedCount > 0, nil
}

func (m *MomentCommentMgo) DeleteByMoments(ctx context.Context, momentIDs []string) error {
	if len(momentIDs) == 0 {
		return nil
	}
	_, err := m.coll.DeleteMany(ctx, bson.M{"moment_id": bson.M{"$in": momentIDs}})
	return errs.Wrap(err)
}

func (m *MomentCommentMgo) Count(ctx context.Context, filter any) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}
