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

type FriendTagApi struct {
	tagDB controller.FriendTagDatabase
}

func NewFriendTagApi(tagDB controller.FriendTagDatabase) *FriendTagApi {
	return &FriendTagApi{tagDB: tagDB}
}

type CreateFriendTagReq struct {
	OwnerUserID string `json:"ownerUserID" binding:"required"`
	Name        string `json:"name" binding:"required"`
}

type CreateFriendTagResp struct {
	Tag *model.FriendTag `json:"tag"`
}

func (f *FriendTagApi) CreateTag(c *gin.Context) {
	var req CreateFriendTagReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.OwnerUserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	tag, err := f.tagDB.CreateTag(c, req.OwnerUserID, req.Name)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &CreateFriendTagResp{Tag: tag})
}

type RenameFriendTagReq struct {
	OwnerUserID string `json:"ownerUserID" binding:"required"`
	TagID       string `json:"tagID" binding:"required"`
	Name        string `json:"name" binding:"required"`
}

func (f *FriendTagApi) RenameTag(c *gin.Context) {
	var req RenameFriendTagReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.OwnerUserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	if err := f.tagDB.RenameTag(c, req.OwnerUserID, req.TagID, req.Name); err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

type DeleteFriendTagReq struct {
	OwnerUserID string `json:"ownerUserID" binding:"required"`
	TagID       string `json:"tagID" binding:"required"`
}

func (f *FriendTagApi) DeleteTag(c *gin.Context) {
	var req DeleteFriendTagReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.OwnerUserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	if err := f.tagDB.DeleteTag(c, req.OwnerUserID, req.TagID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

type AddFriendTagMembersReq struct {
	OwnerUserID   string   `json:"ownerUserID" binding:"required"`
	TagID         string   `json:"tagID" binding:"required"`
	FriendUserIDs []string `json:"friendUserIDs" binding:"required"`
}

func (f *FriendTagApi) AddMembers(c *gin.Context) {
	var req AddFriendTagMembersReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.OwnerUserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	if err := f.tagDB.AddMembers(c, req.OwnerUserID, req.TagID, req.FriendUserIDs); err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

type RemoveFriendTagMembersReq struct {
	OwnerUserID   string   `json:"ownerUserID" binding:"required"`
	TagID         string   `json:"tagID" binding:"required"`
	FriendUserIDs []string `json:"friendUserIDs" binding:"required"`
}

func (f *FriendTagApi) RemoveMembers(c *gin.Context) {
	var req RemoveFriendTagMembersReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.OwnerUserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	if err := f.tagDB.RemoveMembers(c, req.OwnerUserID, req.TagID, req.FriendUserIDs); err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

type ListFriendTagsReq struct {
	OwnerUserID string `json:"ownerUserID" binding:"required"`
	PageReq
}

type ListFriendTagsResp struct {
	Total int64              `json:"total"`
	Tags  []*model.FriendTag `json:"tags"`
}

func (f *FriendTagApi) ListTags(c *gin.Context) {
	var req ListFriendTagsReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.OwnerUserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	total, tags, err := f.tagDB.ListTags(c, req.OwnerUserID, req.PageReq)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &ListFriendTagsResp{Total: total, Tags: tags})
}

type ListFriendTagMembersReq struct {
	OwnerUserID string `json:"ownerUserID" binding:"required"`
	TagID       string `json:"tagID" binding:"required"`
	PageReq
}

type ListFriendTagMembersResp struct {
	Total   int64                    `json:"total"`
	Members []*model.FriendTagMember `json:"members"`
}

func (f *FriendTagApi) ListMembers(c *gin.Context) {
	var req ListFriendTagMembersReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.OwnerUserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	total, members, err := f.tagDB.ListTagMembers(c, req.OwnerUserID, req.TagID, req.PageReq)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &ListFriendTagMembersResp{Total: total, Members: members})
}
