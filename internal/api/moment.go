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
	"strings"
	"time"

	"server/v3/pkg/common/apiresp"
	"server/v3/pkg/common/errs"
	"github.com/gin-gonic/gin"
	"server/v3/pkg/authverify"
	"server/v3/pkg/common/storage/controller"
	"server/v3/pkg/common/storage/model"
	"server/v3/pkg/notification"
	"server/v3/pkg/rpcli"
)

type PageReq struct {
	PageNumber int32 `json:"pageNumber"`
	ShowNumber int32 `json:"showNumber"`
}

func (p PageReq) GetPageNumber() int32 { return p.PageNumber }
func (p PageReq) GetShowNumber() int32 { return p.ShowNumber }

type MomentApi struct {
	momentDB   controller.MomentDatabase
	notifier   *notification.NotificationSender
	userClient *rpcli.UserClient
}

func NewMomentApi(momentDB controller.MomentDatabase, notifier *notification.NotificationSender, userClient *rpcli.UserClient) *MomentApi {
	return &MomentApi{momentDB: momentDB, notifier: notifier, userClient: userClient}
}

type CreateMomentReq struct {
	UserID          string                  `json:"userID" binding:"required"`
	Content         string                  `json:"content"`
	Media           []model.MomentMediaItem `json:"media"`
	VisibleType     int32                   `json:"visibleType" binding:"required"`
	VisibleGroupIDs []string                `json:"visibleGroupIDs"`
	Extra           string                  `json:"extra"`
}

type CreateMomentResp struct {
	Moment         *model.MomentWithInteractions `json:"moment"`
	SyncSeq        int64                         `json:"syncSeq,omitempty"`
	VisibleUserIDs []string                      `json:"visibleUserIDs,omitempty"`
}

func (m *MomentApi) CreateMoment(c *gin.Context) {
	var req CreateMomentReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		apiresp.GinError(c, errs.ErrArgs.WithDetail("content is empty").Wrap())
		return
	}
	if err := validateMomentMedia(req.Media); err != nil {
		apiresp.GinError(c, err)
		return
	}
	moment, err := m.momentDB.CreateMoment(c, req.UserID, req.Content, req.Media, req.VisibleType, req.VisibleGroupIDs, req.Extra)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	userInfo := m.fetchMomentUserInfo(c, req.UserID)
	visibleUserIDs, _ := m.momentDB.ResolveVisibleUserIDs(c, moment)
	apiresp.GinSuccess(c, &CreateMomentResp{
		Moment: &model.MomentWithInteractions{
			Moment:   moment,
			UserInfo: userInfo,
			Likes:    []*model.MomentLikeWithUser{},
			Comments: []*model.MomentCommentWithUser{},
		},
		SyncSeq:        moment.UpdateSeq,
		VisibleUserIDs: visibleUserIDs,
	})
}

func validateMomentMedia(media []model.MomentMediaItem) error {
	if len(media) == 0 {
		return nil
	}
	const (
		mediaTypeImage = "image"
		mediaTypeVideo = "video"
	)
	imageCount := 0
	videoCount := 0
	for _, item := range media {
		t := strings.ToLower(strings.TrimSpace(item.Type))
		switch t {
		case mediaTypeImage:
			imageCount++
		case mediaTypeVideo:
			videoCount++
		default:
			return errs.ErrArgs.WithDetail("invalid media type").Wrap()
		}
	}
	if videoCount > 0 && imageCount > 0 {
		return errs.ErrArgs.WithDetail("cannot mix image and video").Wrap()
	}
	if videoCount > 1 {
		return errs.ErrArgs.WithDetail("only one video is allowed").Wrap()
	}
	if imageCount > 9 {
		return errs.ErrArgs.WithDetail("image count exceeds 9").Wrap()
	}
	return nil
}

type DeleteMomentReq struct {
	UserID   string `json:"userID" binding:"required"`
	MomentID string `json:"momentID" binding:"required"`
}

type MomentMutationResp struct {
	SyncSeq        int64                     `json:"syncSeq,omitempty"`
	VisibleUserIDs []string                  `json:"visibleUserIDs,omitempty"`
	Like           *model.MomentLikeWithUser `json:"like,omitempty"`
}

func (m *MomentApi) DeleteMoment(c *gin.Context) {
	var req DeleteMomentReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	var visibleUserIDs []string
	if existing, err := m.momentDB.GetMoment(c, req.MomentID); err == nil && existing != nil {
		visibleUserIDs, _ = m.momentDB.ResolveVisibleUserIDs(c, existing)
	}
	syncSeq, err := m.momentDB.DeleteMoment(c, req.UserID, req.MomentID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &MomentMutationResp{
		SyncSeq:        syncSeq,
		VisibleUserIDs: visibleUserIDs,
	})
}

type ListMomentsReq struct {
	UserID      string `json:"userID" binding:"required"`
	OwnerUserID string `json:"ownerUserID"`
	PageReq
}

type ListMomentsResp struct {
	Total   int64                           `json:"total"`
	Moments []*model.MomentWithInteractions `json:"moments"`
}

func (m *MomentApi) ListMoments(c *gin.Context) {
	var req ListMomentsReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	total, moments, err := m.momentDB.ListVisibleMoments(c, req.UserID, req.OwnerUserID, req.PageReq)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	result, err := m.enrichMoments(c, moments)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &ListMomentsResp{Total: total, Moments: result})
}

type ListTimelineReq struct {
	UserID      string `json:"userID" binding:"required"`
	OwnerUserID string `json:"ownerUserID"`
	BeforeSeq   int64  `json:"beforeSeq"`
	Limit       int64  `json:"limit"`
}

type ListTimelineResp struct {
	Moments       []*model.MomentWithInteractions `json:"moments"`
	NextBeforeSeq int64                           `json:"nextBeforeSeq"`
	HasMore       bool                            `json:"hasMore"`
	LatestSyncSeq int64                           `json:"latestSyncSeq"`
}

func (m *MomentApi) ListTimeline(c *gin.Context) {
	var req ListTimelineReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	moments, nextBeforeSeq, hasMore, latestSyncSeq, err := m.momentDB.ListVisibleTimeline(c, req.UserID, req.OwnerUserID, req.BeforeSeq, req.Limit)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	result, err := m.enrichMoments(c, moments)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &ListTimelineResp{
		Moments:       result,
		NextBeforeSeq: nextBeforeSeq,
		HasMore:       hasMore,
		LatestSyncSeq: latestSyncSeq,
	})
}

type SyncMomentsReq struct {
	UserID      string `json:"userID" binding:"required"`
	LastSyncSeq int64  `json:"lastSyncSeq"`
	Limit       int64  `json:"limit"`
}

type SyncMomentsResp struct {
	LatestSyncSeq    int64                           `json:"latestSyncSeq"`
	NeedFullRefresh  bool                            `json:"needFullRefresh"`
	HasMore          bool                            `json:"hasMore"`
	Upserts          []*model.MomentWithInteractions `json:"upserts"`
	DeletedMomentIDs []string                        `json:"deletedMomentIDs"`
}

func (m *MomentApi) SyncMoments(c *gin.Context) {
	var req SyncMomentsReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	upserts, deletedIDs, latestSyncSeq, needFullRefresh, hasMore, err := m.momentDB.SyncVisibleMoments(c, req.UserID, req.LastSyncSeq, req.Limit)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	enriched, err := m.enrichMoments(c, upserts)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	if deletedIDs == nil {
		deletedIDs = []string{}
	}
	apiresp.GinSuccess(c, &SyncMomentsResp{
		LatestSyncSeq:    latestSyncSeq,
		NeedFullRefresh:  needFullRefresh,
		HasMore:          hasMore,
		Upserts:          enriched,
		DeletedMomentIDs: deletedIDs,
	})
}

func (m *MomentApi) enrichMoments(c *gin.Context, moments []*model.Moment) ([]*model.MomentWithInteractions, error) {
	if len(moments) == 0 {
		return []*model.MomentWithInteractions{}, nil
	}

	momentIDs := make([]string, len(moments))
	for i, moment := range moments {
		momentIDs[i] = moment.MomentID
	}

	comments, err := m.momentDB.FindCommentsByMoments(c, momentIDs)
	if err != nil {
		return nil, err
	}
	likes, err := m.momentDB.FindLikesByMoments(c, momentIDs)
	if err != nil {
		return nil, err
	}

	userIDSet := make(map[string]struct{})
	for _, moment := range moments {
		userIDSet[moment.UserID] = struct{}{}
	}
	for _, comment := range comments {
		userIDSet[comment.UserID] = struct{}{}
		if comment.ReplyToUserID != "" {
			userIDSet[comment.ReplyToUserID] = struct{}{}
		}
	}
	for _, like := range likes {
		userIDSet[like.UserID] = struct{}{}
	}

	userIDs := make([]string, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}

	userInfoMap := make(map[string]*model.MomentUserInfo)
	if len(userIDs) > 0 && m.userClient != nil {
		users, err := m.userClient.GetUsersInfo(c, userIDs)
		if err == nil {
			for _, u := range users {
				userInfoMap[u.UserID] = &model.MomentUserInfo{
					UserID:   u.UserID,
					Nickname: u.Nickname,
					FaceURL:  u.FaceURL,
				}
			}
		}
	}

	commentsByMoment := make(map[string][]*model.MomentCommentWithUser)
	for _, comment := range comments {
		commentWithUser := &model.MomentCommentWithUser{
			CommentID:     comment.CommentID,
			MomentID:      comment.MomentID,
			UserID:        comment.UserID,
			ReplyToUserID: comment.ReplyToUserID,
			Content:       comment.Content,
			Status:        comment.Status,
			CreateTime:    comment.CreateTime,
			UpdateTime:    comment.UpdateTime,
			UserInfo:      userInfoMap[comment.UserID],
		}
		if comment.ReplyToUserID != "" {
			commentWithUser.ReplyToUser = userInfoMap[comment.ReplyToUserID]
		}
		commentsByMoment[comment.MomentID] = append(commentsByMoment[comment.MomentID], commentWithUser)
	}

	likesByMoment := make(map[string][]*model.MomentLikeWithUser)
	for _, like := range likes {
		likeWithUser := &model.MomentLikeWithUser{
			MomentID:   like.MomentID,
			UserID:     like.UserID,
			CreateTime: like.CreateTime,
			UserInfo:   userInfoMap[like.UserID],
		}
		likesByMoment[like.MomentID] = append(likesByMoment[like.MomentID], likeWithUser)
	}

	result := make([]*model.MomentWithInteractions, len(moments))
	for i, moment := range moments {
		result[i] = &model.MomentWithInteractions{
			Moment:   moment,
			UserInfo: userInfoMap[moment.UserID],
			Likes:    likesByMoment[moment.MomentID],
			Comments: commentsByMoment[moment.MomentID],
		}
		if result[i].Likes == nil {
			result[i].Likes = []*model.MomentLikeWithUser{}
		}
		if result[i].Comments == nil {
			result[i].Comments = []*model.MomentCommentWithUser{}
		}
	}
	return result, nil
}

type CommentMomentReq struct {
	UserID        string `json:"userID" binding:"required"`
	MomentID      string `json:"momentID" binding:"required"`
	ReplyToUserID string `json:"replyToUserID"`
	Content       string `json:"content" binding:"required"`
}

type CommentMomentResp struct {
	Comment        *model.MomentCommentWithUser `json:"comment"`
	SyncSeq        int64                        `json:"syncSeq,omitempty"`
	VisibleUserIDs []string                     `json:"visibleUserIDs,omitempty"`
}

func (m *MomentApi) CommentMoment(c *gin.Context) {
	var req CommentMomentReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		apiresp.GinError(c, errs.ErrArgs.WithDetail("content is empty").Wrap())
		return
	}
	comment, syncSeq, err := m.momentDB.CommentMoment(c, req.UserID, req.MomentID, req.ReplyToUserID, req.Content)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	commentWithUser := &model.MomentCommentWithUser{
		CommentID:     comment.CommentID,
		MomentID:      comment.MomentID,
		UserID:        comment.UserID,
		ReplyToUserID: comment.ReplyToUserID,
		Content:       comment.Content,
		Status:        comment.Status,
		CreateTime:    comment.CreateTime,
		UpdateTime:    comment.UpdateTime,
		UserInfo:      m.fetchMomentUserInfo(c, comment.UserID),
	}
	if comment.ReplyToUserID != "" {
		commentWithUser.ReplyToUser = m.fetchMomentUserInfo(c, comment.ReplyToUserID)
	}
	var visibleUserIDs []string
	if existing, err := m.momentDB.GetMoment(c, req.MomentID); err == nil && existing != nil {
		visibleUserIDs, _ = m.momentDB.ResolveVisibleUserIDs(c, existing)
	}
	apiresp.GinSuccess(c, &CommentMomentResp{
		Comment:        commentWithUser,
		SyncSeq:        syncSeq,
		VisibleUserIDs: visibleUserIDs,
	})
}

type DeleteCommentReq struct {
	UserID    string `json:"userID" binding:"required"`
	CommentID string `json:"commentID" binding:"required"`
}

type DeleteCommentResp struct {
	MomentID       string   `json:"momentID,omitempty"`
	SyncSeq        int64    `json:"syncSeq,omitempty"`
	VisibleUserIDs []string `json:"visibleUserIDs,omitempty"`
}

func (m *MomentApi) DeleteComment(c *gin.Context) {
	var req DeleteCommentReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	comment, syncSeq, err := m.momentDB.DeleteComment(c, req.UserID, req.CommentID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	var visibleUserIDs []string
	var momentID string
	if comment != nil {
		momentID = comment.MomentID
		if existing, err := m.momentDB.GetMoment(c, comment.MomentID); err == nil && existing != nil {
			visibleUserIDs, _ = m.momentDB.ResolveVisibleUserIDs(c, existing)
		}
	}
	apiresp.GinSuccess(c, &DeleteCommentResp{
		MomentID:       momentID,
		SyncSeq:        syncSeq,
		VisibleUserIDs: visibleUserIDs,
	})
}

type LikeMomentReq struct {
	UserID   string `json:"userID" binding:"required"`
	MomentID string `json:"momentID" binding:"required"`
}

func (m *MomentApi) LikeMoment(c *gin.Context) {
	var req LikeMomentReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	syncSeq, err := m.momentDB.LikeMoment(c, req.UserID, req.MomentID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	var visibleUserIDs []string
	if existing, err := m.momentDB.GetMoment(c, req.MomentID); err == nil && existing != nil {
		visibleUserIDs, _ = m.momentDB.ResolveVisibleUserIDs(c, existing)
	}
	apiresp.GinSuccess(c, &MomentMutationResp{
		SyncSeq:        syncSeq,
		VisibleUserIDs: visibleUserIDs,
		Like: &model.MomentLikeWithUser{
			MomentID:   req.MomentID,
			UserID:     req.UserID,
			CreateTime: time.Now(),
			UserInfo:   m.fetchMomentUserInfo(c, req.UserID),
		},
	})
}

type UnlikeMomentReq struct {
	UserID   string `json:"userID" binding:"required"`
	MomentID string `json:"momentID" binding:"required"`
}

func (m *MomentApi) UnlikeMoment(c *gin.Context) {
	var req UnlikeMomentReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if err := authverify.CheckAccess(c, req.UserID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	syncSeq, err := m.momentDB.UnlikeMoment(c, req.UserID, req.MomentID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	var visibleUserIDs []string
	if existing, err := m.momentDB.GetMoment(c, req.MomentID); err == nil && existing != nil {
		visibleUserIDs, _ = m.momentDB.ResolveVisibleUserIDs(c, existing)
	}
	apiresp.GinSuccess(c, &MomentMutationResp{
		SyncSeq:        syncSeq,
		VisibleUserIDs: visibleUserIDs,
	})
}

func (m *MomentApi) fetchMomentUserInfo(c *gin.Context, userID string) *model.MomentUserInfo {
	if userID == "" || m.userClient == nil {
		return nil
	}
	users, err := m.userClient.GetUsersInfo(c, []string{userID})
	if err != nil || len(users) == 0 {
		return nil
	}
	u := users[0]
	return &model.MomentUserInfo{
		UserID:   u.UserID,
		Nickname: u.Nickname,
		FaceURL:  u.FaceURL,
	}
}
