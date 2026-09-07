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

type Favorite struct {
	FavoriteID string    `bson:"favorite_id" json:"favoriteID"`
	UserID     string    `bson:"user_id" json:"userID"`
	TargetType string    `bson:"target_type" json:"targetType"`
	TargetID   string    `bson:"target_id" json:"targetID"`
	Data       string    `bson:"data" json:"data"`
	CreateTime time.Time `bson:"create_time" json:"createTime"`
}
