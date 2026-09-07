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

	"gitee.com/matkurban/openimsdk-tools/db/pagination"
	"server/v3/pkg/common/storage/model"
)

type Favorite interface {
	Add(ctx context.Context, favorite *model.Favorite) error
	Remove(ctx context.Context, userID, targetType, targetID string) error
	Take(ctx context.Context, userID, targetType, targetID string) (*model.Favorite, error)
	FindByUser(ctx context.Context, userID string, pagination pagination.Pagination) (int64, []*model.Favorite, error)
	UpdateData(ctx context.Context, userID, targetType, targetID, data string) error
}
