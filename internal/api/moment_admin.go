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
)

const (
	momentAdminMaxBatch    = 100
	momentAdminDefaultPage = 20
	momentAdminMaxPage     = 100
	momentAdminMaxKeyword  = 100
)

// normalizeAdminPage clamps pagination: pageNumber>=1, 1<=showNumber<=100 (default 20).
func normalizeAdminPage(p PageReq) PageReq {
	if p.PageNumber <= 0 {
		p.PageNumber = 1
	}
	if p.ShowNumber <= 0 {
		p.ShowNumber = momentAdminDefaultPage
	}
	if p.ShowNumber > momentAdminMaxPage {
		p.ShowNumber = momentAdminMaxPage
	}
	return p
}

// normalizeMomentIDs trims, de-duplicates and validates a batch of moment IDs.
func normalizeMomentIDs(ids []string) ([]string, error) {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil, errs.ErrArgs.WrapMsg("momentIDs is empty")
	}
	if len(out) > momentAdminMaxBatch {
		return nil, errs.ErrArgs.WrapMsg("too many momentIDs, max 100")
	}
	return out, nil
}

// shanghaiDayStart returns 00:00 (Asia/Shanghai, fixed UTC+8) of the day containing now.
func shanghaiDayStart(now time.Time) time.Time {
	loc := time.FixedZone("Asia/Shanghai", 8*3600)
	n := now.In(loc)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
}

func (m *MomentApi) checkAdmin(c *gin.Context) bool {
	if err := authverify.CheckAdmin(c); err != nil {
		apiresp.GinError(c, err)
		return false
	}
	return true
}

func (m *MomentApi) userInfoMap(c *gin.Context, userIDs []string) map[string]*model.MomentUserInfo {
	res := make(map[string]*model.MomentUserInfo)
	set := make(map[string]struct{}, len(userIDs))
	ids := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		if _, ok := set[id]; id == "" || ok {
			continue
		}
		set[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 || m.userClient == nil {
		return res
	}
	users, err := m.userClient.GetUsersInfo(c, ids)
	if err != nil {
		return res
	}
	for _, u := range users {
		res[u.UserID] = &model.MomentUserInfo{UserID: u.UserID, Nickname: u.Nickname, FaceURL: u.FaceURL}
	}
	return res
}

// AdminMoment is a moment with its author info (likeCount/commentCount are on the moment).
type AdminMoment struct {
	*model.Moment
	UserInfo *model.MomentUserInfo `json:"userInfo,omitempty"`
}

type AdminListMomentsReq struct {
	UserID    string `json:"userID"`
	Keyword   string `json:"keyword"`
	StartTime int64  `json:"startTime"` // ms, inclusive, 0 = unbounded
	EndTime   int64  `json:"endTime"`   // ms, inclusive, 0 = unbounded
	Status    *int32 `json:"status"`    // omitted = normal+banned; 0 normal; 1 user-deleted; 2 banned
	PageReq
}

type AdminListMomentsResp struct {
	Total   int64          `json:"total"`
	Moments []*AdminMoment `json:"moments"`
}

func (m *MomentApi) AdminListMoments(c *gin.Context) {
	var req AdminListMomentsReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if !m.checkAdmin(c) {
		return
	}
	if len([]rune(req.Keyword)) > momentAdminMaxKeyword {
		apiresp.GinError(c, errs.ErrArgs.WrapMsg("keyword too long"))
		return
	}
	total, moments, err := m.momentDB.AdminListMoments(c, controller.AdminMomentFilter{
		UserID:    strings.TrimSpace(req.UserID),
		Keyword:   strings.TrimSpace(req.Keyword),
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    req.Status,
	}, normalizeAdminPage(req.PageReq))
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	userIDs := make([]string, 0, len(moments))
	for _, mo := range moments {
		userIDs = append(userIDs, mo.UserID)
	}
	infos := m.userInfoMap(c, userIDs)
	items := make([]*AdminMoment, len(moments))
	for i, mo := range moments {
		items[i] = &AdminMoment{Moment: mo, UserInfo: infos[mo.UserID]}
	}
	apiresp.GinSuccess(c, &AdminListMomentsResp{Total: total, Moments: items})
}

type AdminMomentIDReq struct {
	MomentID string `json:"momentID" binding:"required"`
}

type AdminMomentDetailResp struct {
	Moment *AdminMoment `json:"moment"`
}

func (m *MomentApi) AdminGetMoment(c *gin.Context) {
	var req AdminMomentIDReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if !m.checkAdmin(c) {
		return
	}
	mo, err := m.momentDB.GetMoment(c, req.MomentID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &AdminMomentDetailResp{Moment: &AdminMoment{Moment: mo, UserInfo: m.userInfoMap(c, []string{mo.UserID})[mo.UserID]}})
}

type AdminDeleteMomentsReq struct {
	MomentIDs []string `json:"momentIDs"`
}

type AdminDeleteMomentsResp struct {
	DeletedCount int64 `json:"deletedCount"`
}

// AdminDeleteMoments physically deletes moments (cascade comments + likes). Used for single (1 id) and batch (<=100).
func (m *MomentApi) AdminDeleteMoments(c *gin.Context) {
	var req AdminDeleteMomentsReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if !m.checkAdmin(c) {
		return
	}
	ids, err := normalizeMomentIDs(req.MomentIDs)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	n, err := m.momentDB.AdminDeleteMoments(c, ids)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &AdminDeleteMomentsResp{DeletedCount: n})
}

type AdminListCommentsReq struct {
	MomentID string `json:"momentID" binding:"required"`
	PageReq
}

type AdminListCommentsResp struct {
	Total    int64                          `json:"total"`
	Comments []*model.MomentCommentWithUser `json:"comments"`
}

func (m *MomentApi) AdminListComments(c *gin.Context) {
	var req AdminListCommentsReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if !m.checkAdmin(c) {
		return
	}
	total, comments, err := m.momentDB.ListComments(c, req.MomentID, normalizeAdminPage(req.PageReq))
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	ids := make([]string, 0, len(comments)*2)
	for _, cm := range comments {
		ids = append(ids, cm.UserID, cm.ReplyToUserID)
	}
	infos := m.userInfoMap(c, ids)
	out := make([]*model.MomentCommentWithUser, len(comments))
	for i, cm := range comments {
		out[i] = &model.MomentCommentWithUser{
			CommentID: cm.CommentID, MomentID: cm.MomentID, UserID: cm.UserID, ReplyToUserID: cm.ReplyToUserID,
			Content: cm.Content, Status: cm.Status, CreateTime: cm.CreateTime, UpdateTime: cm.UpdateTime,
			UserInfo: infos[cm.UserID],
		}
		if cm.ReplyToUserID != "" {
			out[i].ReplyToUser = infos[cm.ReplyToUserID]
		}
	}
	apiresp.GinSuccess(c, &AdminListCommentsResp{Total: total, Comments: out})
}

type AdminDeleteCommentReq struct {
	CommentID string `json:"commentID" binding:"required"`
}

func (m *MomentApi) AdminDeleteComment(c *gin.Context) {
	var req AdminDeleteCommentReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if !m.checkAdmin(c) {
		return
	}
	if _, err := m.momentDB.AdminDeleteComment(c, req.CommentID); err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

type AdminListLikesReq struct {
	MomentID string `json:"momentID" binding:"required"`
	PageReq
}

type AdminListLikesResp struct {
	Total int64                       `json:"total"`
	Likes []*model.MomentLikeWithUser `json:"likes"`
}

func (m *MomentApi) AdminListLikes(c *gin.Context) {
	var req AdminListLikesReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if !m.checkAdmin(c) {
		return
	}
	total, likes, err := m.momentDB.ListLikes(c, req.MomentID, normalizeAdminPage(req.PageReq))
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	ids := make([]string, 0, len(likes))
	for _, l := range likes {
		ids = append(ids, l.UserID)
	}
	infos := m.userInfoMap(c, ids)
	out := make([]*model.MomentLikeWithUser, len(likes))
	for i, l := range likes {
		out[i] = &model.MomentLikeWithUser{MomentID: l.MomentID, UserID: l.UserID, CreateTime: l.CreateTime, UserInfo: infos[l.UserID]}
	}
	apiresp.GinSuccess(c, &AdminListLikesResp{Total: total, Likes: out})
}

type AdminSetStatusResp struct {
	Affected int64 `json:"affected"`
}

func (m *MomentApi) adminSetStatus(c *gin.Context, status int32) {
	var req AdminDeleteMomentsReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if !m.checkAdmin(c) {
		return
	}
	ids, err := normalizeMomentIDs(req.MomentIDs)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	n, err := m.momentDB.AdminSetMomentStatus(c, ids, status)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, &AdminSetStatusResp{Affected: n})
}

func (m *MomentApi) AdminBanMoments(c *gin.Context) { m.adminSetStatus(c, model.MomentStatusBanned) }
func (m *MomentApi) AdminRestoreMoments(c *gin.Context) {
	m.adminSetStatus(c, model.MomentStatusNormal)
}

func (m *MomentApi) AdminStats(c *gin.Context) {
	if !m.checkAdmin(c) {
		return
	}
	st, err := m.momentDB.AdminStats(c, shanghaiDayStart(time.Now()))
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, st)
}
