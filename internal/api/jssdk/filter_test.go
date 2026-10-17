package jssdk

import (
	"testing"

	"github.com/openimsdk/protocol/conversation"
)

func TestFilterConversationSeqs(t *testing.T) {
	x := &JSSdk{}
	conversations := []*conversation.Conversation{
		{ConversationID: "c1", MinSeq: 0},
		{ConversationID: "c2", MinSeq: 10},
		{ConversationID: "c3", MinSeq: 5},
		{ConversationID: "c4", MinSeq: 0},
	}
	maxSeqs := map[string]int64{
		"c1": 3,
		"c2": 8,  // below MinSeq -> filtered
		"c3": 5,  // equal MinSeq -> kept
		"c4": 0,  // zero -> filtered
	}
	got := x.filterConversationSeqs(conversations, maxSeqs)
	if len(got) != 2 {
		t.Fatalf("expected 2 seqs, got %d: %+v", len(got), got)
	}
	if got[0].ConversationID != "c1" || got[0].Seqs[0] != 3 {
		t.Fatalf("unexpected first: %+v", got[0])
	}
	if got[1].ConversationID != "c3" || got[1].Seqs[0] != 5 {
		t.Fatalf("unexpected second: %+v", got[1])
	}
}
