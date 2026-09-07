package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Friend represents the data structure for a friend relationship in MongoDB.
type Friend struct {
	CreateTime     time.Time     `bson:"create_time"`
	OwnerUserID    string        `bson:"owner_user_id"`
	FriendUserID   string        `bson:"friend_user_id"`
	Remark         string        `bson:"remark"`
	OperatorUserID string        `bson:"operator_user_id"`
	Ex             string        `bson:"ex"`
	AddSource      int32         `bson:"add_source"`
	ID             bson.ObjectID `bson:"_id"`
	IsPinned       bool          `bson:"is_pinned"`
}
