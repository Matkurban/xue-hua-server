package model

import (
	"time"
)

type Conversation struct {
	CreateTime            time.Time `bson:"create_time"`
	LatestMsgDestructTime time.Time `bson:"latest_msg_destruct_time"`
	OwnerUserID           string    `bson:"owner_user_id"`
	ConversationID        string    `bson:"conversation_id"`
	UserID                string    `bson:"user_id"`
	GroupID               string    `bson:"group_id"`
	AttachedInfo          string    `bson:"attached_info"`
	Ex                    string    `bson:"ex"`
	MaxSeq                int64     `bson:"max_seq"`
	MinSeq                int64     `bson:"min_seq"`
	MsgDestructTime       int64     `bson:"msg_destruct_time"`
	ConversationType      int32     `bson:"conversation_type"`
	RecvMsgOpt            int32     `bson:"recv_msg_opt"`
	BurnDuration          int32     `bson:"burn_duration"`
	GroupAtType           int32     `bson:"group_at_type"`
	IsPinned              bool      `bson:"is_pinned"`
	IsPrivateChat         bool      `bson:"is_private_chat"`
	IsMsgDestruct         bool      `bson:"is_msg_destruct"`
}
