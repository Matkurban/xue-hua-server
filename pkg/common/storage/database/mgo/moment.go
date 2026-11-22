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

	"server/v3/pkg/common/errs"
	"server/v3/pkg/common/storage/database"
	"server/v3/pkg/common/storage/model"
	"server/v3/pkg/common/storage/mongoutil"
	"server/v3/pkg/common/storage/pagination"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	momentSeqCollName = "moment_seq"
	momentGlobalSeqID = "moment_global_seq"
)

type momentSeqDoc struct {
	ID  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}

type MomentMgo struct {
	coll    *mongo.Collection
	seqColl *mongo.Collection
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
		{Keys: bson.D{{Key: "timeline_seq", Value: -1}}},
		{Keys: bson.D{{Key: "update_seq", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "timeline_seq", Value: -1}}},
	})
	if err != nil {
		return nil, err
	}
	m := &MomentMgo{
		coll:    coll,
		seqColl: db.Collection(momentSeqCollName),
	}
	if err := m.backfillLegacySeq(context.Background()); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MomentMgo) backfillLegacySeq(ctx context.Context) error {
	filter := bson.M{
		"$or": []bson.M{
			{"timeline_seq": bson.M{"$exists": false}},
			{"timeline_seq": int64(0)},
		},
	}
	opt := options.Find().SetSort(bson.D{{Key: "create_time", Value: 1}, {Key: "_id", Value: 1}})
	legacy, err := mongoutil.Find[*model.Moment](ctx, m.coll, filter, opt)
	if err != nil || len(legacy) == 0 {
		return err
	}
	for _, item := range legacy {
		seq, err := m.NextSeq(ctx)
		if err != nil {
			return err
		}
		updateSeq := item.UpdateSeq
		if updateSeq == 0 {
			updateSeq = seq
		}
		if err := mongoutil.UpdateOne(ctx, m.coll, bson.M{"moment_id": item.MomentID}, bson.M{"$set": bson.M{
			"timeline_seq": seq,
			"update_seq":   updateSeq,
		}}, false); err != nil {
			return err
		}
	}
	return nil
}

func (m *MomentMgo) NextSeq(ctx context.Context) (int64, error) {
	opt := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)
	var doc momentSeqDoc
	err := m.seqColl.FindOneAndUpdate(
		ctx,
		bson.M{"_id": momentGlobalSeqID},
		bson.M{"$inc": bson.M{"seq": int64(1)}},
		opt,
	).Decode(&doc)
	if err != nil {
		return 0, errs.Wrap(err)
	}
	return doc.Seq, nil
}

func (m *MomentMgo) CurrentSeq(ctx context.Context) (int64, error) {
	var doc momentSeqDoc
	err := m.seqColl.FindOne(ctx, bson.M{"_id": momentGlobalSeqID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, errs.Wrap(err)
	}
	return doc.Seq, nil
}

func (m *MomentMgo) Create(ctx context.Context, moment *model.Moment) error {
	if moment.CreateTime.IsZero() {
		moment.CreateTime = time.Now()
	}
	moment.UpdateTime = moment.CreateTime
	seq, err := m.NextSeq(ctx)
	if err != nil {
		return err
	}
	if moment.TimelineSeq == 0 {
		moment.TimelineSeq = seq
	}
	moment.UpdateSeq = seq
	return mongoutil.InsertOne(ctx, m.coll, moment)
}

func (m *MomentMgo) UpdateByMap(ctx context.Context, momentID, userID string, args map[string]any) error {
	if len(args) == 0 {
		return nil
	}
	seq, err := m.NextSeq(ctx)
	if err != nil {
		return err
	}
	args["update_time"] = time.Now()
	args["update_seq"] = seq
	filter := bson.M{"moment_id": momentID, "user_id": userID}
	return mongoutil.UpdateOne(ctx, m.coll, filter, bson.M{"$set": args}, true)
}

func (m *MomentMgo) Delete(ctx context.Context, momentID, userID string) (int64, error) {
	seq, err := m.NextSeq(ctx)
	if err != nil {
		return 0, err
	}
	filter := bson.M{"moment_id": momentID, "user_id": userID}
	if err := mongoutil.UpdateOne(ctx, m.coll, filter, bson.M{"$set": bson.M{
		"status":      int32(1),
		"update_time": time.Now(),
		"update_seq":  seq,
	}}, true); err != nil {
		return 0, err
	}
	return seq, nil
}

func (m *MomentMgo) Take(ctx context.Context, momentID string) (*model.Moment, error) {
	return mongoutil.FindOne[*model.Moment](ctx, m.coll, bson.M{"moment_id": momentID})
}

func (m *MomentMgo) FindPage(ctx context.Context, filter any, pagination pagination.Pagination) (int64, []*model.Moment, error) {
	opt := options.Find().SetSort(bson.D{{Key: "timeline_seq", Value: -1}, {Key: "create_time", Value: -1}})
	return mongoutil.FindPage[*model.Moment](ctx, m.coll, filter, pagination, opt)
}

func (m *MomentMgo) FindTimeline(ctx context.Context, filter any, beforeSeq int64, limit int64) ([]*model.Moment, error) {
	query := bson.M{}
	if fm, ok := filter.(bson.M); ok {
		for k, v := range fm {
			query[k] = v
		}
	}
	if beforeSeq > 0 {
		query["timeline_seq"] = bson.M{"$lt": beforeSeq}
	}
	opt := options.Find().
		SetSort(bson.D{{Key: "timeline_seq", Value: -1}, {Key: "create_time", Value: -1}}).
		SetLimit(limit)
	return mongoutil.Find[*model.Moment](ctx, m.coll, query, opt)
}

func (m *MomentMgo) FindUpdatedSince(ctx context.Context, filter any, sinceSeq int64, limit int64) ([]*model.Moment, error) {
	query := bson.M{}
	if fm, ok := filter.(bson.M); ok {
		for k, v := range fm {
			query[k] = v
		}
	}
	query["update_seq"] = bson.M{"$gt": sinceSeq}
	opt := options.Find().
		SetSort(bson.D{{Key: "update_seq", Value: 1}}).
		SetLimit(limit)
	return mongoutil.Find[*model.Moment](ctx, m.coll, query, opt)
}

func (m *MomentMgo) IncCounts(ctx context.Context, momentID string, likeDelta, commentDelta int64) (int64, error) {
	seq, err := m.NextSeq(ctx)
	if err != nil {
		return 0, err
	}
	set := bson.M{
		"update_time": time.Now(),
		"update_seq":  seq,
	}
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
	if err := mongoutil.UpdateOne(ctx, m.coll, bson.M{"moment_id": momentID}, update, true); err != nil {
		return 0, err
	}
	return seq, nil
}

func (m *MomentMgo) Count(ctx context.Context, filter any) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}

// SetStatus sets the status of the given moments, ignoring user-deleted ones (status 1).
func (m *MomentMgo) SetStatus(ctx context.Context, momentIDs []string, status int32) (int64, error) {
	if len(momentIDs) == 0 {
		return 0, nil
	}
	seq, err := m.NextSeq(ctx)
	if err != nil {
		return 0, err
	}
	filter := bson.M{"moment_id": bson.M{"$in": momentIDs}, "status": bson.M{"$ne": int32(1)}}
	res, err := m.coll.UpdateMany(ctx, filter, bson.M{"$set": bson.M{
		"status":      status,
		"update_time": time.Now(),
		"update_seq":  seq,
	}})
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
