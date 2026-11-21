package msg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/v3/pkg/common/errs"
	"github.com/openimsdk/protocol/constant"
	pbgroup "github.com/openimsdk/protocol/group"
	pbmsg "github.com/openimsdk/protocol/msg"
	"github.com/openimsdk/protocol/sdkws"
	pbuser "github.com/openimsdk/protocol/user"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"server/v3/pkg/apistruct"
	"server/v3/pkg/authverify"
	"server/v3/pkg/common/config"
	"server/v3/pkg/common/storage/controller"
	"server/v3/pkg/common/webhook"
	"server/v3/pkg/rpccache"
	"server/v3/pkg/rpcli"
)

type contentTypeMsgDatabase struct {
	controller.CommonMsgDatabase
	enqueued []*sdkws.MsgData
}

func (d *contentTypeMsgDatabase) MsgToMQ(_ context.Context, _ string, data *sdkws.MsgData) error {
	d.enqueued = append(d.enqueued, data)
	return nil
}

type contentTypeGroupClient struct{ pbgroup.GroupClient }

func (contentTypeGroupClient) GetGroupInfoCache(_ context.Context, req *pbgroup.GetGroupInfoCacheReq, _ ...grpc.CallOption) (*pbgroup.GetGroupInfoCacheResp, error) {
	return &pbgroup.GetGroupInfoCacheResp{
		GroupInfo: &sdkws.GroupInfo{GroupID: req.GroupID, GroupType: constant.SuperGroup},
	}, nil
}

type contentTypeUserClient struct{ pbuser.UserClient }

func (contentTypeUserClient) GetGlobalRecvMessageOpt(context.Context, *pbuser.GetGlobalRecvMessageOptReq, ...grpc.CallOption) (*pbuser.GetGlobalRecvMessageOptResp, error) {
	return &pbuser.GetGlobalRecvMessageOptResp{GlobalRecvMsgOpt: constant.ReceiveNotNotifyMessage}, nil
}

func newContentTypeServer() (*msgServer, *contentTypeMsgDatabase) {
	db := &contentTypeMsgDatabase{}
	return &msgServer{
		MsgDatabase: db, config: &Config{}, adminUserIDs: []string{"sender"},
		webhookClient: &webhook.Client{},
		GroupLocalCache: rpccache.NewGroupLocalCache(
			&rpcli.GroupClient{GroupClient: contentTypeGroupClient{}}, &config.LocalCache{}, nil),
		UserLocalCache: rpccache.NewUserLocalCache(
			&rpcli.UserClient{UserClient: contentTypeUserClient{}}, &config.LocalCache{}, nil),
	}, db
}

func contentTypeSendReq(sessionType, contentType int32) *pbmsg.SendMsgReq {
	return &pbmsg.SendMsgReq{MsgData: &sdkws.MsgData{
		SendID: "sender", RecvID: "receiver", GroupID: "group", ClientMsgID: "client-msg",
		SessionType: sessionType, ContentType: contentType, Content: []byte(`{"content":"hello"}`),
		Options: map[string]bool{},
	}}
}

func TestSendMsgRejectsRetiredContentTypes(t *testing.T) {
	ctx := authverify.WithTempAdmin(context.Background())
	for _, sessionType := range []int32{constant.SingleChatType, constant.ReadGroupChatType, constant.NotificationChatType} {
		for _, contentType := range []int32{130, 131, 132} {
			t.Run(fmt.Sprintf("%d/%d", sessionType, contentType), func(t *testing.T) {
				// No dependencies: every session must fail before verification or enqueue.
				server := &msgServer{}
				req := contentTypeSendReq(sessionType, contentType)
				resp, err := server.SendMsg(ctx, req)
				require.Nil(t, resp)
				require.True(t, errs.ErrArgs.Is(err))
				require.ErrorContains(t, err, "retired content type")
				require.Empty(t, req.MsgData.ServerMsgID)

				simpleResp, err := server.SendSimpleMsg(ctx, &pbmsg.SendSimpleMsgReq{MsgData: req.MsgData})
				require.Nil(t, simpleResp)
				require.True(t, errs.ErrArgs.Is(err))
			})
		}
	}
}

func TestEnqueueMsgRejectsRetiredContentTypes(t *testing.T) {
	for _, contentType := range []int32{130, 131, 132} {
		server := &msgServer{}
		err := server.enqueueMsg(context.Background(), "conversation", &sdkws.MsgData{ContentType: contentType})
		require.True(t, errs.ErrArgs.Is(err))
		require.ErrorContains(t, err, "retired content type")
	}
}

func TestSendMsgPreservesSupportedContentTypes(t *testing.T) {
	contentTypes := []int32{
		apistruct.CallSignal, apistruct.RedPacket, apistruct.RedPacketGrabNotify,
		apistruct.Conference, apistruct.MeetingReminder, apistruct.MeetingScheduleNotify,
		constant.Custom, constant.BusinessNotification, 133, 999,
	}
	for _, sessionType := range []int32{constant.SingleChatType, constant.ReadGroupChatType, constant.NotificationChatType} {
		for _, contentType := range contentTypes {
			t.Run(fmt.Sprintf("%d/%d", sessionType, contentType), func(t *testing.T) {
				server, db := newContentTypeServer()
				req := contentTypeSendReq(sessionType, contentType)
				resp, err := server.SendMsg(authverify.WithTempAdmin(context.Background()), req)
				require.NoError(t, err)
				require.NotEmpty(t, resp.ServerMsgID)
				require.Len(t, db.enqueued, 1)
				require.Equal(t, contentType, db.enqueued[0].ContentType)
			})
		}
	}
}

func TestSendMsgCallbackCannotEnqueueRetiredContentTypes(t *testing.T) {
	for _, sessionType := range []int32{constant.SingleChatType, constant.ReadGroupChatType} {
		for _, contentType := range []int32{130, 131, 132, apistruct.RedPacket, apistruct.MeetingReminder, 999} {
			t.Run(fmt.Sprintf("%d/%d", sessionType, contentType), func(t *testing.T) {
				hook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					// Changing the session as well must not bypass the final guard.
					_ = json.NewEncoder(w).Encode(map[string]any{
						"contentType": contentType, "sessionType": constant.NotificationChatType,
					})
				}))
				defer hook.Close()
				server, db := newContentTypeServer()
				server.webhookClient = webhook.NewWebhookClient(hook.URL)
				server.config.WebhooksConfig.BeforeMsgModify = config.BeforeConfig{Enable: true, Timeout: 5}
				req := contentTypeSendReq(sessionType, constant.Text)
				resp, err := server.SendMsg(authverify.WithTempAdmin(context.Background()), req)
				require.Equal(t, contentType, req.MsgData.ContentType, "callback must actually rewrite the message")
				if contentType >= 130 && contentType <= 132 {
					require.Nil(t, resp)
					require.True(t, errs.ErrArgs.Is(err))
					require.ErrorContains(t, err, "retired content type")
					require.Empty(t, db.enqueued)
				} else {
					require.NoError(t, err)
					require.Len(t, db.enqueued, 1)
					require.Equal(t, contentType, db.enqueued[0].ContentType)
					require.NotNil(t, resp.Modify)
				}
			})
		}
	}
}

func TestEncapsulatePreservesBusinessMessageOptionsAndIDs(t *testing.T) {
	contentTypes := []int32{
		apistruct.CallSignal, apistruct.RedPacket, apistruct.RedPacketGrabNotify,
		apistruct.Conference, apistruct.MeetingReminder, apistruct.MeetingScheduleNotify,
	}
	require.Equal(t, []int32{124, 125, 126, 127, 128, 129}, contentTypes)
	for _, contentType := range contentTypes {
		req := contentTypeSendReq(constant.SingleChatType, contentType)
		req.MsgData.Options = map[string]bool{
			constant.IsHistory: true, constant.IsPersistent: true,
			constant.IsUnreadCount: true, constant.IsOfflinePush: true,
		}
		(&msgServer{}).encapsulateMsgData(req.MsgData)
		for option, enabled := range req.MsgData.Options {
			require.True(t, enabled, "content type %d option %s", contentType, option)
		}
	}
}
