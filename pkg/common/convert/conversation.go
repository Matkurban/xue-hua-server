package convert

import (
	"server/v3/pkg/util/datautil"
	"github.com/openimsdk/protocol/conversation"
	"server/v3/pkg/common/storage/model"
)

func ConversationDB2Pb(conversationDB *model.Conversation) *conversation.Conversation {
	conversationPB := &conversation.Conversation{}
	conversationPB.LatestMsgDestructTime = conversationDB.LatestMsgDestructTime.UnixMilli()
	if err := datautil.CopyStructFields(conversationPB, conversationDB); err != nil {
		return nil
	}
	return conversationPB
}

func ConversationsDB2Pb(conversationsDB []*model.Conversation) (conversationsPB []*conversation.Conversation) {
	for _, conversationDB := range conversationsDB {
		conversationPB := &conversation.Conversation{}
		if err := datautil.CopyStructFields(conversationPB, conversationDB); err != nil {
			continue
		}
		conversationPB.LatestMsgDestructTime = conversationDB.LatestMsgDestructTime.UnixMilli()
		conversationsPB = append(conversationsPB, conversationPB)
	}
	return conversationsPB
}
