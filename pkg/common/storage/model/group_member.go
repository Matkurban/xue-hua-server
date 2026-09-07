package model

import (
	"time"
)

type GroupMember struct {
	JoinTime       time.Time `bson:"join_time"`
	MuteEndTime    time.Time `bson:"mute_end_time"`
	GroupID        string    `bson:"group_id"`
	UserID         string    `bson:"user_id"`
	Nickname       string    `bson:"nickname"`
	FaceURL        string    `bson:"face_url"`
	InviterUserID  string    `bson:"inviter_user_id"`
	OperatorUserID string    `bson:"operator_user_id"`
	Ex             string    `bson:"ex"`
	RoleLevel      int32     `bson:"role_level"`
	JoinSource     int32     `bson:"join_source"`
}
