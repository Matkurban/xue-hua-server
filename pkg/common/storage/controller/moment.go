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
	"regexp"
	"time"

	"gitee.com/matkurban/openimsdk-tools/db/pagination"
	"gitee.com/matkurban/openimsdk-tools/errs"
	"gitee.com/matkurban/openimsdk-tools/utils/datautil"
	"gitee.com/matkurban/openimsdk-tools/utils/idutil"
	"go.mongodb.org/mongo-driver/v2/bson"
	"server/v3/pkg/common/storage/database"
	"server/v3/pkg/common/storage/database/mgo"
	"server/v3/pkg/common/storage/model"
)

type MomentDatabase interface {
	CreateMoment(ctx context.Context, userID, content string, media []model.MomentMediaItem, visibleType int32, visibleGroupIDs []string, extra string) (*model.Moment, error)
	DeleteMoment(ctx context.Context, userID, momentID string) error
	ListVisibleMoments(ctx context.Context, viewerUserID, ownerUserID string, pagination pagination.Pagination) (int64, []*model.Moment, error)
	GetMoment(ctx context.Context, momentID string) (*model.Moment, error)
	LikeMoment(ctx context.Context, userID, momentID string) error
	UnlikeMoment(ctx context.Context, userID, momentID string) error
	CommentMoment(ctx context.Context, userID, momentID, replyToUserID, content string) (*model.MomentComment, error)
	DeleteComment(ctx context.Context, userID, commentID string) (*model.MomentComment, error)
	ListComments(ctx context.Context, momentID string, pagination pagination.Pagination) (int64, []*model.MomentComment, error)
	ListLikes(ctx context.Context, momentID string, pagination pagination.Pagination) (int64, []*model.MomentLike, error)
	ResolveVisibleUserIDs(ctx context.Context, moment *model.Moment) ([]string, error)
	FindCommentsByMoments(ctx context.Context, momentIDs []string) ([]*model.MomentComment, error)
	FindLikesByMoments(ctx context.Context, momentIDs []string) ([]*model.MomentLike, error)

	// admin operations
	AdminListMoments(ctx context.Context, filter AdminMomentFilter, pagination pagination.Pagination) (int64, []*model.Moment, error)
	AdminDeleteMoments(ctx context.Context, momentIDs []string) (int64, error)
	AdminSetMomentStatus(ctx context.Context, momentIDs []string, status int32) (int64, error)
	AdminDeleteComment(ctx context.Context, commentID string) (*model.MomentComment, error)
	AdminStats(ctx context.Context, dayStart time.Time) (*MomentStats, error)
}

type momentDatabase struct {
	moment    database.Moment
	comment   database.MomentComment
	like      database.MomentLike
	friend    database.Friend
	friendTag FriendTagDatabase
}

func NewMomentDatabase(moment database.Moment, comment database.MomentComment, like database.MomentLike, friend database.Friend, friendTag FriendTagDatabase) MomentDatabase {
	return &momentDatabase{
		moment:    moment,
		comment:   comment,
		like:      like,
		friend:    friend,
		friendTag: friendTag,
	}
}

func (m *momentDatabase) CreateMoment(ctx context.Context, userID, content string, media []model.MomentMediaItem, visibleType int32, visibleGroupIDs []string, extra string) (*model.Moment, error) {
	if userID == "" {
		return nil, errs.ErrArgs.WrapMsg("userID is empty")
	}
	if visibleType == int32(model.MomentVisibleGroups) && len(visibleGroupIDs) == 0 {
		return nil, errs.ErrArgs.WrapMsg("visibleGroupIDs is empty")
	}
	if visibleType != int32(model.MomentVisibleFriends) && visibleType != int32(model.MomentVisibleGroups) && visibleType != int32(model.MomentVisibleSelf) {
		return nil, errs.ErrArgs.WrapMsg("invalid visibleType")
	}
	moment := &model.Moment{
		MomentID:        idutil.OperationIDGenerator(),
		UserID:          userID,
		Content:         content,
		Media:           media,
		VisibleType:     visibleType,
		VisibleGroupIDs: visibleGroupIDs,
		Status:          0,
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
		LikeCount:       0,
		CommentCount:    0,
		Extra:           extra,
	}
	if err := m.moment.Create(ctx, moment); err != nil {
		return nil, err
	}
	return moment, nil
}

func (m *momentDatabase) DeleteMoment(ctx context.Context, userID, momentID string) error {
	if userID == "" || momentID == "" {
		return errs.ErrArgs.WrapMsg("userID/momentID is empty")
	}
	return m.moment.Delete(ctx, momentID, userID)
}

func (m *momentDatabase) GetMoment(ctx context.Context, momentID string) (*model.Moment, error) {
	if momentID == "" {
		return nil, errs.ErrArgs.WrapMsg("momentID is empty")
	}
	return m.moment.Take(ctx, momentID)
}

func (m *momentDatabase) LikeMoment(ctx context.Context, userID, momentID string) error {
	if userID == "" || momentID == "" {
		return errs.ErrArgs.WrapMsg("userID/momentID is empty")
	}
	if _, err := m.like.Take(ctx, momentID, userID); err == nil {
		return errs.ErrArgs.WrapMsg("already liked")
	} else if !mgo.IsNotFound(err) {
		return err
	}
	if err := m.like.Create(ctx, &model.MomentLike{MomentID: momentID, UserID: userID, CreateTime: time.Now()}); err != nil {
		return err
	}
	return m.moment.IncCounts(ctx, momentID, 1, 0)
}

func (m *momentDatabase) UnlikeMoment(ctx context.Context, userID, momentID string) error {
	if userID == "" || momentID == "" {
		return errs.ErrArgs.WrapMsg("userID/momentID is empty")
	}
	if err := m.like.Delete(ctx, momentID, userID); err != nil {
		return err
	}
	return m.moment.IncCounts(ctx, momentID, -1, 0)
}

func (m *momentDatabase) CommentMoment(ctx context.Context, userID, momentID, replyToUserID, content string) (*model.MomentComment, error) {
	if userID == "" || momentID == "" || content == "" {
		return nil, errs.ErrArgs.WrapMsg("userID/momentID/content is empty")
	}
	comment := &model.MomentComment{
		CommentID:     idutil.OperationIDGenerator(),
		MomentID:      momentID,
		UserID:        userID,
		ReplyToUserID: replyToUserID,
		Content:       content,
		Status:        0,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}
	if err := m.comment.Create(ctx, comment); err != nil {
		return nil, err
	}
	if err := m.moment.IncCounts(ctx, momentID, 0, 1); err != nil {
		return nil, err
	}
	return comment, nil
}

func (m *momentDatabase) DeleteComment(ctx context.Context, userID, commentID string) (*model.MomentComment, error) {
	if userID == "" || commentID == "" {
		return nil, errs.ErrArgs.WrapMsg("userID/commentID is empty")
	}
	comment, err := m.comment.Take(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if comment.UserID != userID {
		return nil, errs.ErrArgs.WrapMsg("no permission to delete comment")
	}
	if err := m.comment.Delete(ctx, commentID, userID); err != nil {
		return nil, err
	}
	if err := m.moment.IncCounts(ctx, comment.MomentID, 0, -1); err != nil {
		return nil, err
	}
	return comment, nil
}

func (m *momentDatabase) ListComments(ctx context.Context, momentID string, pagination pagination.Pagination) (int64, []*model.MomentComment, error) {
	if momentID == "" {
		return 0, nil, errs.ErrArgs.WrapMsg("momentID is empty")
	}
	return m.comment.FindByMoment(ctx, momentID, pagination)
}

func (m *momentDatabase) ListLikes(ctx context.Context, momentID string, pagination pagination.Pagination) (int64, []*model.MomentLike, error) {
	if momentID == "" {
		return 0, nil, errs.ErrArgs.WrapMsg("momentID is empty")
	}
	return m.like.FindByMoment(ctx, momentID, pagination)
}

func (m *momentDatabase) FindCommentsByMoments(ctx context.Context, momentIDs []string) ([]*model.MomentComment, error) {
	if len(momentIDs) == 0 {
		return nil, nil
	}
	return m.comment.FindByMoments(ctx, momentIDs)
}

func (m *momentDatabase) FindLikesByMoments(ctx context.Context, momentIDs []string) ([]*model.MomentLike, error) {
	if len(momentIDs) == 0 {
		return nil, nil
	}
	return m.like.FindByMoments(ctx, momentIDs)
}

func (m *momentDatabase) ResolveVisibleUserIDs(ctx context.Context, moment *model.Moment) ([]string, error) {
	if moment == nil {
		return nil, errs.ErrArgs.WrapMsg("moment is nil")
	}
	switch moment.VisibleType {
	case int32(model.MomentVisibleSelf):
		return []string{moment.UserID}, nil
	case int32(model.MomentVisibleFriends):
		friendIDs, err := m.friend.FindFriendUserIDs(ctx, moment.UserID)
		if err != nil {
			return nil, err
		}
		friendIDs = append(friendIDs, moment.UserID)
		return datautil.Distinct(friendIDs), nil
	case int32(model.MomentVisibleGroups):
		if len(moment.VisibleGroupIDs) == 0 {
			return []string{moment.UserID}, nil
		}
		members, err := m.friendTag.FindMembersByOwner(ctx, moment.UserID, moment.VisibleGroupIDs)
		if err != nil {
			return nil, err
		}
		userIDs := []string{moment.UserID}
		for _, member := range members {
			userIDs = append(userIDs, member.FriendUserID)
		}
		return datautil.Distinct(userIDs), nil
	default:
		return []string{moment.UserID}, nil
	}
}

func (m *momentDatabase) ListVisibleMoments(ctx context.Context, viewerUserID, ownerUserID string, pagination pagination.Pagination) (int64, []*model.Moment, error) {
	if viewerUserID == "" {
		return 0, nil, errs.ErrArgs.WrapMsg("viewerUserID is empty")
	}
	filter := bson.M{"status": bson.M{"$nin": []int32{model.MomentStatusDeleted, model.MomentStatusBanned}}}
	if ownerUserID != "" {
		filter["user_id"] = ownerUserID
		if ownerUserID == viewerUserID {
			return m.moment.FindPage(ctx, filter, pagination)
		}
		friendIDs, err := m.friend.FindFriendUserIDs(ctx, viewerUserID)
		if err != nil {
			return 0, nil, err
		}
		orFilters := make([]bson.M, 0, 2)
		if datautil.Contain(ownerUserID, friendIDs...) {
			orFilters = append(orFilters, bson.M{"visible_type": int32(model.MomentVisibleFriends)})
		}
		tagIDs, err := m.friendTag.FindTagIDsByFriend(ctx, ownerUserID, viewerUserID)
		if err != nil {
			return 0, nil, err
		}
		if len(tagIDs) > 0 {
			orFilters = append(orFilters, bson.M{"visible_type": int32(model.MomentVisibleGroups), "visible_group_ids": bson.M{"$in": tagIDs}})
		}
		if len(orFilters) == 0 {
			return 0, []*model.Moment{}, nil
		}
		filter["$or"] = orFilters
		return m.moment.FindPage(ctx, filter, pagination)
	}

	friendIDs, err := m.friend.FindFriendUserIDs(ctx, viewerUserID)
	if err != nil {
		return 0, nil, err
	}
	friendIDs = datautil.Distinct(append(friendIDs, viewerUserID))

	groupMembers, err := m.friendTag.FindMembersByFriend(ctx, viewerUserID)
	if err != nil {
		return 0, nil, err
	}
	ownerTagMap := make(map[string][]string)
	for _, member := range groupMembers {
		ownerTagMap[member.OwnerUserID] = append(ownerTagMap[member.OwnerUserID], member.TagID)
	}
	groupFilters := make([]bson.M, 0, len(ownerTagMap))
	for owner, tagIDs := range ownerTagMap {
		groupFilters = append(groupFilters, bson.M{"user_id": owner, "visible_group_ids": bson.M{"$in": tagIDs}})
	}

	orFilters := []bson.M{
		{"user_id": viewerUserID},
		{"visible_type": int32(model.MomentVisibleFriends), "user_id": bson.M{"$in": friendIDs}},
	}
	if len(groupFilters) > 0 {
		orFilters = append(orFilters, bson.M{"visible_type": int32(model.MomentVisibleGroups), "$or": groupFilters})
	}
	filter["$or"] = orFilters

	return m.moment.FindPage(ctx, filter, pagination)
}

// AdminMomentFilter is the filter of the admin moment list. Zero values mean "no filter".
type AdminMomentFilter struct {
	UserID    string
	Keyword   string
	StartTime int64 // ms, inclusive
	EndTime   int64 // ms, inclusive
	// Status: nil = every non user-deleted moment; otherwise exactly that status
	// (0 matches legacy documents lacking the field).
	Status *int32
}

// BuildAdminMomentFilter converts the admin filter to a mongo filter.
func BuildAdminMomentFilter(f AdminMomentFilter) (bson.M, error) {
	if f.StartTime < 0 || f.EndTime < 0 {
		return nil, errs.ErrArgs.WrapMsg("time must not be negative")
	}
	if f.StartTime > 0 && f.EndTime > 0 && f.StartTime > f.EndTime {
		return nil, errs.ErrArgs.WrapMsg("startTime is after endTime")
	}
	filter := bson.M{}
	if f.Status == nil {
		filter["status"] = bson.M{"$ne": model.MomentStatusDeleted}
	} else {
		switch *f.Status {
		case model.MomentStatusNormal:
			filter["status"] = bson.M{"$nin": []int32{model.MomentStatusDeleted, model.MomentStatusBanned}}
		case model.MomentStatusDeleted, model.MomentStatusBanned:
			filter["status"] = *f.Status
		default:
			return nil, errs.ErrArgs.WrapMsg("invalid status")
		}
	}
	if f.UserID != "" {
		filter["user_id"] = f.UserID
	}
	if kw := f.Keyword; kw != "" {
		filter["content"] = bson.M{"$regex": regexp.QuoteMeta(kw), "$options": "i"}
	}
	if f.StartTime > 0 || f.EndTime > 0 {
		r := bson.M{}
		if f.StartTime > 0 {
			r["$gte"] = time.UnixMilli(f.StartTime)
		}
		if f.EndTime > 0 {
			r["$lte"] = time.UnixMilli(f.EndTime)
		}
		filter["create_time"] = r
	}
	return filter, nil
}

type MomentStats struct {
	TotalMoments  int64 `json:"totalMoments"`
	TodayMoments  int64 `json:"todayMoments"`
	BannedMoments int64 `json:"bannedMoments"`
	TotalComments int64 `json:"totalComments"`
	TotalLikes    int64 `json:"totalLikes"`
}

func (m *momentDatabase) AdminListMoments(ctx context.Context, f AdminMomentFilter, pagination pagination.Pagination) (int64, []*model.Moment, error) {
	filter, err := BuildAdminMomentFilter(f)
	if err != nil {
		return 0, nil, err
	}
	return m.moment.FindPage(ctx, filter, pagination)
}

// AdminDeleteMoments physically deletes moments together with their comments and likes.
// Returns the number of moments deleted.
func (m *momentDatabase) AdminDeleteMoments(ctx context.Context, momentIDs []string) (int64, error) {
	momentIDs = datautil.Distinct(momentIDs)
	if len(momentIDs) == 0 {
		return 0, errs.ErrArgs.WrapMsg("momentIDs is empty")
	}
	if err := m.comment.DeleteByMoments(ctx, momentIDs); err != nil {
		return 0, err
	}
	if err := m.like.DeleteByMoments(ctx, momentIDs); err != nil {
		return 0, err
	}
	return m.moment.DeleteByIDs(ctx, momentIDs)
}

// AdminSetMomentStatus bans (MomentStatusBanned) or restores (MomentStatusNormal) moments.
// User-deleted moments are untouched. Returns the number of matched moments.
func (m *momentDatabase) AdminSetMomentStatus(ctx context.Context, momentIDs []string, status int32) (int64, error) {
	if status != model.MomentStatusNormal && status != model.MomentStatusBanned {
		return 0, errs.ErrArgs.WrapMsg("invalid status")
	}
	momentIDs = datautil.Distinct(momentIDs)
	if len(momentIDs) == 0 {
		return 0, errs.ErrArgs.WrapMsg("momentIDs is empty")
	}
	return m.moment.SetStatus(ctx, momentIDs, status)
}

func (m *momentDatabase) AdminDeleteComment(ctx context.Context, commentID string) (*model.MomentComment, error) {
	if commentID == "" {
		return nil, errs.ErrArgs.WrapMsg("commentID is empty")
	}
	comment, err := m.comment.Take(ctx, commentID)
	if err != nil {
		return nil, err
	}
	changed, err := m.comment.AdminDelete(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if changed {
		if err := m.moment.IncCounts(ctx, comment.MomentID, 0, -1); err != nil && !mgo.IsNotFound(err) {
			return nil, err
		}
	}
	return comment, nil
}

func (m *momentDatabase) AdminStats(ctx context.Context, dayStart time.Time) (*MomentStats, error) {
	notDeleted := bson.M{"status": bson.M{"$ne": model.MomentStatusDeleted}}
	var st MomentStats
	var err error
	if st.TotalMoments, err = m.moment.Count(ctx, notDeleted); err != nil {
		return nil, err
	}
	if st.TodayMoments, err = m.moment.Count(ctx, bson.M{"status": bson.M{"$ne": model.MomentStatusDeleted}, "create_time": bson.M{"$gte": dayStart}}); err != nil {
		return nil, err
	}
	if st.BannedMoments, err = m.moment.Count(ctx, bson.M{"status": model.MomentStatusBanned}); err != nil {
		return nil, err
	}
	if st.TotalComments, err = m.comment.Count(ctx, notDeleted); err != nil {
		return nil, err
	}
	if st.TotalLikes, err = m.like.Count(ctx, bson.M{}); err != nil {
		return nil, err
	}
	return &st, nil
}
