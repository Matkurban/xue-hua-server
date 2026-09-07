package msg

import (
	"gitee.com/matkurban/openimsdk-tools/errs"
	"github.com/openimsdk/protocol/msg"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func IsNotFound(err error) bool {
	switch errs.Unwrap(err) {
	case redis.Nil, mongo.ErrNoDocuments:
		return true
	default:
		return false
	}
}

type activeConversations []*msg.ActiveConversation

func (s activeConversations) Len() int {
	return len(s)
}

func (s activeConversations) Less(i, j int) bool {
	return s[i].LastTime > s[j].LastTime
}

func (s activeConversations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//type seqTime struct {
//	ConversationID string
//	Seq            int64
//	Time           int64
//	Unread         int64
//	Pinned         bool
//}
//
//func (s seqTime) String() string {
//	return fmt.Sprintf("<Time_%d,Unread_%d,Pinned_%t>", s.Time, s.Unread, s.Pinned)
//}
//
//type seqTimes []seqTime
//
//func (s seqTimes) Len() int {
//	return len(s)
//}
//
//// Less sticky priority, unread priority, time descending
//func (s seqTimes) Less(i, j int) bool {
//	iv, jv := s[i], s[j]
//	if iv.Pinned && (!jv.Pinned) {
//		return true
//	}
//	if jv.Pinned && (!iv.Pinned) {
//		return false
//	}
//	if iv.Unread > 0 && jv.Unread == 0 {
//		return true
//	}
//	if jv.Unread > 0 && iv.Unread == 0 {
//		return false
//	}
//	return iv.Time > jv.Time
//}
//
//func (s seqTimes) Swap(i, j int) {
//	s[i], s[j] = s[j], s[i]
//}
//
//type conversationStatus struct {
//	ConversationID string
//	Pinned         bool
//	Recv           bool
//}
