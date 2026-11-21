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

package database

import (
	"context"

	"server/v3/pkg/common/storage/pagination"
	"server/v3/pkg/common/storage/model"
)

type Moment interface {
	Create(ctx context.Context, moment *model.Moment) error
	UpdateByMap(ctx context.Context, momentID, userID string, args map[string]any) error
	Delete(ctx context.Context, momentID, userID string) error
	Take(ctx context.Context, momentID string) (*model.Moment, error)
	FindPage(ctx context.Context, filter any, pagination pagination.Pagination) (int64, []*model.Moment, error)
	IncCounts(ctx context.Context, momentID string, likeDelta, commentDelta int64) error
	// admin operations
	Count(ctx context.Context, filter any) (int64, error)
	SetStatus(ctx context.Context, momentIDs []string, status int32) (int64, error)
	DeleteByIDs(ctx context.Context, momentIDs []string) (int64, error)
	FindByIDs(ctx context.Context, momentIDs []string) ([]*model.Moment, error)
}

type MomentComment interface {
	Create(ctx context.Context, comment *model.MomentComment) error
	Delete(ctx context.Context, commentID, userID string) error
	Take(ctx context.Context, commentID string) (*model.MomentComment, error)
	FindByMoment(ctx context.Context, momentID string, pagination pagination.Pagination) (int64, []*model.MomentComment, error)
	FindByMoments(ctx context.Context, momentIDs []string) ([]*model.MomentComment, error)
	// admin operations
	AdminDelete(ctx context.Context, commentID string) (bool, error)
	DeleteByMoments(ctx context.Context, momentIDs []string) error
	Count(ctx context.Context, filter any) (int64, error)
}

type MomentLike interface {
	Create(ctx context.Context, like *model.MomentLike) error
	Delete(ctx context.Context, momentID, userID string) error
	Take(ctx context.Context, momentID, userID string) (*model.MomentLike, error)
	FindByMoment(ctx context.Context, momentID string, pagination pagination.Pagination) (int64, []*model.MomentLike, error)
	FindByMoments(ctx context.Context, momentIDs []string) ([]*model.MomentLike, error)
	// admin operations
	DeleteByMoments(ctx context.Context, momentIDs []string) error
	Count(ctx context.Context, filter any) (int64, error)
}
