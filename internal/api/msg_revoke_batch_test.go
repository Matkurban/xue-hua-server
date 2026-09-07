package api

import (
	"testing"

	"github.com/openimsdk/protocol/constant"
	"github.com/openimsdk/protocol/sdkws"
	"server/v3/pkg/common/servererrs"
	"server/v3/pkg/msgprocessor"
)

func TestNotificationConversationIDIsStable(t *testing.T) {
	got := msgprocessor.GetConversationIDBySessionType(int(constant.NotificationChatType), "b-user", "a-user")
	want := "sn_a-user_b-user"
	if got != want {
		t.Fatalf("conversationID = %q, want %q", got, want)
	}
}

func TestRevokeErrIsSuccess(t *testing.T) {
	if !revokeErrIsSuccess(nil) {
		t.Fatal("nil error should count as success")
	}
	if !revokeErrIsSuccess(servererrs.ErrMsgAlreadyRevoke.Wrap()) {
		t.Fatal("already-revoked should count as success")
	}
	if revokeErrIsSuccess(servererrs.ErrArgs.WrapMsg("bad")) {
		t.Fatal("other errors should not count as success")
	}
}

func TestOASendAndRevokeUseSameConversationID(t *testing.T) {
	msg := &sdkws.MsgData{
		SendID:      "imAdmin",
		RecvID:      "user-1",
		SessionType: constant.NotificationChatType,
		Options:     map[string]bool{},
	}
	sendID := msgprocessor.GetConversationIDByMsg(msg)
	revokeID := msgprocessor.GetConversationIDBySessionType(int(constant.NotificationChatType), msg.SendID, msg.RecvID)
	if sendID != revokeID {
		t.Fatalf("send conversation %q != revoke conversation %q", sendID, revokeID)
	}
	if sendID != "sn_imAdmin_user-1" && sendID != "sn_user-1_imAdmin" {
		t.Fatalf("unexpected conversationID %q", sendID)
	}
}

func TestOAAndRevokeContentTypes(t *testing.T) {
	if constant.OANotification != 1400 {
		t.Fatalf("OANotification=%d want 1400", constant.OANotification)
	}
	if constant.MsgRevokeNotification != 2101 {
		t.Fatalf("MsgRevokeNotification=%d want 2101 (client revoke placeholder)", constant.MsgRevokeNotification)
	}
	if constant.NotificationChatType != 4 {
		t.Fatalf("NotificationChatType=%d want 4", constant.NotificationChatType)
	}
}
