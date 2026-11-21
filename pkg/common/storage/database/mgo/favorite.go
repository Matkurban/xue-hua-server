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

type FavoriteMgo struct {
	coll *mongo.Collection
}

func NewFavoriteMongo(db *mongo.Database) (database.Favorite, error) {
	coll := db.Collection(database.FavoriteName)
	_, err := coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "favorite_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "target_type", Value: 1}, {Key: "target_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "create_time", Value: -1}}},
	})
	if err != nil {
		return nil, err
	}
	return &FavoriteMgo{coll: coll}, nil
}

func (f *FavoriteMgo) Add(ctx context.Context, favorite *model.Favorite) error {
	if favorite.CreateTime.IsZero() {
		favorite.CreateTime = time.Now()
	}
	return mongoutil.InsertOne(ctx, f.coll, favorite)
}

func (f *FavoriteMgo) Remove(ctx context.Context, userID, targetType, targetID string) error {
	filter := bson.M{"user_id": userID, "target_type": targetType, "target_id": targetID}
	return mongoutil.DeleteOne(ctx, f.coll, filter)
}

func (f *FavoriteMgo) Take(ctx context.Context, userID, targetType, targetID string) (*model.Favorite, error) {
	filter := bson.M{"user_id": userID, "target_type": targetType, "target_id": targetID}
	return mongoutil.FindOne[*model.Favorite](ctx, f.coll, filter)
}

func (f *FavoriteMgo) FindByUser(ctx context.Context, userID string, pagination pagination.Pagination) (int64, []*model.Favorite, error) {
	filter := bson.M{"user_id": userID}
	opt := options.Find().SetSort(bson.D{{Key: "create_time", Value: -1}})
	return mongoutil.FindPage[*model.Favorite](ctx, f.coll, filter, pagination, opt)
}

func (f *FavoriteMgo) UpdateData(ctx context.Context, userID, targetType, targetID, data string) error {
	filter := bson.M{"user_id": userID, "target_type": targetType, "target_id": targetID}
	update := bson.M{"$set": bson.M{"data": data}}
	return mongoutil.UpdateOne(ctx, f.coll, filter, update, false)
}
