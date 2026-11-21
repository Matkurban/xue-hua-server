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

package api

import (
	"server/v3/pkg/common/apiresp"
	"server/v3/pkg/common/errs"
	"github.com/gin-gonic/gin"
	"server/v3/pkg/authverify"
	"server/v3/pkg/common/storage/controller"
	"server/v3/pkg/common/storage/model"
)

type FavoriteApi struct {
	favoriteDB controller.FavoriteDatabase
}

func NewFavoriteApi(favoriteDB controller.FavoriteDatabase) *FavoriteApi {
	return &FavoriteApi{favoriteDB: favoriteDB}
}

type AddFavoriteReq struct {
	UserID     string `json:"userID" binding:"required"`
	TargetType string `json:"targetType" binding:"required"`
	TargetID   string `json:"targetID" binding:"required"`
	Data       string `json:"data"`
}

type AddFavoriteResp struct {
	Favorite *model.Favorite `json:"favorite"`
}

func (f *FavoriteApi) AddFavorite(c *gin.Context) {
	var req AddFavoriteReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	fav, err := f.favoriteDB.Add(c, req.UserID, req.TargetType, req.TargetID, req.Data)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &AddFavoriteResp{Favorite: fav})
}

type RemoveFavoriteReq struct {
	UserID     string `json:"userID" binding:"required"`
	TargetType string `json:"targetType" binding:"required"`
	TargetID   string `json:"targetID" binding:"required"`
}

func (f *FavoriteApi) RemoveFavorite(c *gin.Context) {
	var req RemoveFavoriteReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	if err := f.favoriteDB.Remove(c, req.UserID, req.TargetType, req.TargetID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

type ListFavoritesReq struct {
	UserID string `json:"userID" binding:"required"`
	PageReq
}

type ListFavoritesResp struct {
	Total     int64             `json:"total"`
	Favorites []*model.Favorite `json:"favorites"`
}

func (f *FavoriteApi) ListFavorites(c *gin.Context) {
	var req ListFavoritesReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	total, favorites, err := f.favoriteDB.List(c, req.UserID, req.PageReq)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &ListFavoritesResp{Total: total, Favorites: favorites})
}

type UpdateFavoriteReq struct {
	UserID     string `json:"userID" binding:"required"`
	TargetType string `json:"targetType" binding:"required"`
	TargetID   string `json:"targetID" binding:"required"`
	Data       string `json:"data" binding:"required"`
}

type UpdateFavoriteResp struct {
	Favorite *model.Favorite `json:"favorite"`
}

func (f *FavoriteApi) UpdateFavorite(c *gin.Context) {
	var req UpdateFavoriteReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	fav, err := f.favoriteDB.Update(c, req.UserID, req.TargetType, req.TargetID, req.Data)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &UpdateFavoriteResp{Favorite: fav})
}
