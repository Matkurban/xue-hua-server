package model

import (
	"time"
)

type GroupRequest struct {
	ReqTime       time.Time `bson:"req_time"`
	HandledTime   time.Time `bson:"handled_time"`
	UserID        string    `bson:"user_id"`
	GroupID       string    `bson:"group_id"`
	ReqMsg        string    `bson:"req_msg"`
	HandledMsg    string    `bson:"handled_msg"`
	HandleUserID  string    `bson:"handle_user_id"`
	InviterUserID string    `bson:"inviter_user_id"`
	Ex            string    `bson:"ex"`
	HandleResult  int32     `bson:"handle_result"`
	JoinSource    int32     `bson:"join_source"`
}
