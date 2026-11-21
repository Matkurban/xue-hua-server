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

type FavoriteDatabase interface {
	Add(ctx context.Context, userID, targetType, targetID, data string) (*model.Favorite, error)
	Remove(ctx context.Context, userID, targetType, targetID string) error
	List(ctx context.Context, userID string, pagination pagination.Pagination) (int64, []*model.Favorite, error)
	Update(ctx context.Context, userID, targetType, targetID, data string) (*model.Favorite, error)
}

type favoriteDatabase struct {
	favorite database.Favorite
}

func NewFavoriteDatabase(favorite database.Favorite) FavoriteDatabase {
	return &favoriteDatabase{favorite: favorite}
}

func (f *favoriteDatabase) Add(ctx context.Context, userID, targetType, targetID, data string) (*model.Favorite, error) {
	if userID == "" || targetType == "" || targetID == "" {
		return nil, errs.ErrArgs.WrapMsg("userID/targetType/targetID is empty")
	}
	if _, err := f.favorite.Take(ctx, userID, targetType, targetID); err == nil {
		return nil, errs.ErrArgs.WrapMsg("favorite already exists")
	} else if !mgo.IsNotFound(err) {
		return nil, err
	}
	fav := &model.Favorite{
		FavoriteID: idutil.OperationIDGenerator(),
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		Data:       data,
		CreateTime: time.Now(),
	}
	if err := f.favorite.Add(ctx, fav); err != nil {
		return nil, err
	}
	return fav, nil
}

func (f *favoriteDatabase) Remove(ctx context.Context, userID, targetType, targetID string) error {
	if userID == "" || targetType == "" || targetID == "" {
		return errs.ErrArgs.WrapMsg("userID/targetType/targetID is empty")
	}
	return f.favorite.Remove(ctx, userID, targetType, targetID)
}

func (f *favoriteDatabase) List(ctx context.Context, userID string, pagination pagination.Pagination) (int64, []*model.Favorite, error) {
	if userID == "" {
		return 0, nil, errs.ErrArgs.WrapMsg("userID is empty")
	}
	return f.favorite.FindByUser(ctx, userID, pagination)
}

func (f *favoriteDatabase) Update(ctx context.Context, userID, targetType, targetID, data string) (*model.Favorite, error) {
	if userID == "" || targetType == "" || targetID == "" {
		return nil, errs.ErrArgs.WrapMsg("userID/targetType/targetID is empty")
	}
	if err := f.favorite.UpdateData(ctx, userID, targetType, targetID, data); err != nil {
		return nil, err
	}
	return f.favorite.Take(ctx, userID, targetType, targetID)
}
