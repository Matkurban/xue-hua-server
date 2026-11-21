package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/v3/pkg/common/errs"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/openimsdk/protocol/constant"
	"github.com/stretchr/testify/require"
	"server/v3/pkg/apistruct"
	"server/v3/pkg/authverify"
)

func TestGetSendMsgReqRejectsRetiredContentTypes(t *testing.T) {
	api := &MessageApi{validate: validator.New()}
	for _, sessionType := range []int32{constant.SingleChatType, constant.ReadGroupChatType, constant.NotificationChatType} {
		for _, contentType := range []int32{130, 131, 132} {
			t.Run(fmt.Sprintf("%d/%d", sessionType, contentType), func(t *testing.T) {
				req, err := api.getSendMsgReq(nil, apistruct.SendMsg{
					SessionType: sessionType,
					ContentType: contentType,
				})
				require.Nil(t, req)
				require.True(t, errs.ErrArgs.Is(err))
				require.ErrorContains(t, err, "retired content type")
			})
		}
	}
}

func TestGetSendMsgReqPreservesSupportedBusinessMessages(t *testing.T) {
	api := &MessageApi{validate: validator.New()}
	tests := []struct {
		contentType int32
		content     map[string]any
	}{
		{apistruct.CallSignal, map[string]any{
			"roomId": "room", "callType": 1, "callRoomType": 1, "callSignalType": 1,
			"senderUser": map[string]any{"userID": "sender"},
		}},
		{apistruct.RedPacket, map[string]any{"packetID": "packet", "totalAmount": 10, "totalCount": 2}},
		{apistruct.RedPacketGrabNotify, map[string]any{"packetID": "packet", "grabberID": "receiver", "senderID": "sender", "amount": 5}},
		{apistruct.Conference, map[string]any{
			"roomId": "room", "callType": 1, "status": 1, "adminUser": map[string]any{"userID": "sender"},
		}},
		{apistruct.MeetingReminder, map[string]any{"meetingID": "meeting", "occurrenceID": "occurrence", "startTime": 100, "remindMinutes": 5}},
		{apistruct.MeetingScheduleNotify, map[string]any{"meetingID": "meeting", "action": "cancel"}},
		{constant.Custom, map[string]any{"data": "custom payload"}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.contentType), func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req, err := api.getSendMsgReq(c, apistruct.SendMsg{
				SendID: "sender", SessionType: constant.SingleChatType,
				ContentType: tt.contentType, Content: tt.content,
			})
			require.NoError(t, err)
			require.Equal(t, tt.contentType, req.MsgData.ContentType)
			content, err := json.Marshal(tt.content)
			require.NoError(t, err)
			require.JSONEq(t, string(content), string(req.MsgData.Content))
		})
	}
}

func TestBatchSendMsgRejectsRetiredContentTypes(t *testing.T) {
	for _, contentType := range []int32{130, 131, 132} {
		t.Run(fmt.Sprint(contentType), func(t *testing.T) {
			body, err := json.Marshal(apistruct.BatchSendMsgReq{
				SendMsg: apistruct.SendMsg{
					SendID: "sender", SessionType: constant.SingleChatType,
					ContentType: contentType, Content: map[string]any{"data": "retired"},
				},
				RecvIDs: []string{"receiver"},
			})
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)
			engine.ContextWithFallback = true
			c.Request = httptest.NewRequest(http.MethodPost, "/batch_send_msg", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request = c.Request.WithContext(authverify.WithTempAdmin(c.Request.Context()))
			// Nil RPC and database dependencies ensure rejection precedes sending.
			api := &MessageApi{validate: validator.New()}
			api.BatchSendMsg(c)
			var response struct {
				ErrCode int    `json:"errCode"`
				ErrMsg  string `json:"errMsg"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			require.NotZero(t, response.ErrCode)
			require.Contains(t, w.Body.String(), "retired content type")
		})
	}
}
