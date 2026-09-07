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

package model

import "time"

type MomentVisibleType int32

const (
	MomentVisibleFriends MomentVisibleType = 1
	MomentVisibleGroups  MomentVisibleType = 2
	MomentVisibleSelf    MomentVisibleType = 3
)

type MomentMediaItem struct {
	Type     string `bson:"type" json:"type"`
	URL      string `bson:"url" json:"url"`
	CoverURL string `bson:"cover_url" json:"coverURL"`
	Duration int64  `bson:"duration" json:"duration"`
	Extra    string `bson:"extra" json:"extra"`
}

type Moment struct {
	CreateTime      time.Time         `bson:"create_time" json:"createTime"`
	UpdateTime      time.Time         `bson:"update_time" json:"updateTime"`
	MomentID        string            `bson:"moment_id" json:"momentID"`
	UserID          string            `bson:"user_id" json:"userID"`
	Content         string            `bson:"content" json:"content"`
	Extra           string            `bson:"extra" json:"extra"`
	Media           []MomentMediaItem `bson:"media" json:"media"`
	VisibleGroupIDs []string          `bson:"visible_group_ids" json:"visibleGroupIDs"`
	LikeCount       int64             `bson:"like_count" json:"likeCount"`
	CommentCount    int64             `bson:"comment_count" json:"commentCount"`
	VisibleType     int32             `bson:"visible_type" json:"visibleType"`
	Status          int32             `bson:"status" json:"status"`
}

type MomentComment struct {
	CommentID     string    `bson:"comment_id" json:"commentID"`
	MomentID      string    `bson:"moment_id" json:"momentID"`
	UserID        string    `bson:"user_id" json:"userID"`
	ReplyToUserID string    `bson:"reply_to_user_id" json:"replyToUserID"`
	Content       string    `bson:"content" json:"content"`
	Status        int32     `bson:"status" json:"status"`
	CreateTime    time.Time `bson:"create_time" json:"createTime"`
	UpdateTime    time.Time `bson:"update_time" json:"updateTime"`
}

type MomentLike struct {
	MomentID   string    `bson:"moment_id" json:"momentID"`
	UserID     string    `bson:"user_id" json:"userID"`
	CreateTime time.Time `bson:"create_time" json:"createTime"`
}

// MomentUserInfo contains basic user info for display in moments
type MomentUserInfo struct {
	UserID   string `json:"userID"`
	Nickname string `json:"nickname"`
	FaceURL  string `json:"faceURL"`
}

// MomentCommentWithUser contains comment with user info
type MomentCommentWithUser struct {
	CommentID     string          `json:"commentID"`
	MomentID      string          `json:"momentID"`
	UserID        string          `json:"userID"`
	ReplyToUserID string          `json:"replyToUserID"`
	Content       string          `json:"content"`
	Status        int32           `json:"status"`
	CreateTime    time.Time       `json:"createTime"`
	UpdateTime    time.Time       `json:"updateTime"`
	UserInfo      *MomentUserInfo `json:"userInfo"`
	ReplyToUser   *MomentUserInfo `json:"replyToUser,omitempty"`
}

// MomentLikeWithUser contains like with user info
type MomentLikeWithUser struct {
	MomentID   string          `json:"momentID"`
	UserID     string          `json:"userID"`
	CreateTime time.Time       `json:"createTime"`
	UserInfo   *MomentUserInfo `json:"userInfo"`
}

// MomentWithInteractions contains moment with likes and comments
type MomentWithInteractions struct {
	*Moment
	UserInfo *MomentUserInfo          `json:"userInfo,omitempty"`
	Likes    []*MomentLikeWithUser    `json:"likes"`
	Comments []*MomentCommentWithUser `json:"comments"`
}
