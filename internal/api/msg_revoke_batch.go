package api

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"gitee.com/matkurban/openimsdk-tools/apiresp"
	"gitee.com/matkurban/openimsdk-tools/errs"
	"gitee.com/matkurban/openimsdk-tools/log"
	"gitee.com/matkurban/openimsdk-tools/mcontext"
	"github.com/openimsdk/protocol/constant"
	"github.com/openimsdk/protocol/msg"
	"server/v3/pkg/apistruct"
	"server/v3/pkg/authverify"
	"server/v3/pkg/common/servererrs"
	"server/v3/pkg/common/storage/model"
	"server/v3/pkg/msgprocessor"
)

const (
	batchRevokeMaxItems  = 2000
	batchRevokeWorkers   = 32
	batchRevokeFindTries = 8
	batchRevokeFindBase  = 300 * time.Millisecond
	batchRevokeFindCap   = 2 * time.Second
)

// BatchRevokeMsg revokes previously sent messages by conversation + clientMsgID (admin only).
// Seq is resolved from Mongo because SendMsg does not return it (assigned in msgtransfer).
func (m *MessageApi) BatchRevokeMsg(c *gin.Context) {
	if err := authverify.CheckAdmin(c); err != nil {
		apiresp.GinError(c, errs.ErrNoPermission.WrapMsg("only app manager can revoke message"))
		return
	}
	if m.msgDoc == nil {
		apiresp.GinError(c, errs.ErrInternalServer.WrapMsg("msg storage unavailable"))
		return
	}

	var req apistruct.BatchRevokeMsgReq
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, errs.ErrArgs.WithDetail(err.Error()).Wrap())
		return
	}
	if req.SendID == "" {
		apiresp.GinError(c, errs.ErrArgs.WrapMsg("sendID is empty"))
		return
	}
	if len(req.Items) == 0 {
		apiresp.GinError(c, errs.ErrArgs.WrapMsg("items is empty"))
		return
	}
	if len(req.Items) > batchRevokeMaxItems {
		apiresp.GinError(c, errs.ErrArgs.WrapMsg("too many items"))
		return
	}
	sessionType := req.SessionType
	if sessionType == 0 {
		sessionType = constant.NotificationChatType
	}

	ctx, cancel := cloneAdminRPCCtx(c)
	defer cancel()
	var (
		mu            sync.Mutex
		successCount  int
		failedUserIDs []string
		wg            sync.WaitGroup
		sem           = make(chan struct{}, batchRevokeWorkers)
	)

	for _, raw := range req.Items {
		item := raw
		if item.RecvID == "" || item.ClientMsgID == "" {
			mu.Lock()
			failedUserIDs = append(failedUserIDs, item.RecvID)
			mu.Unlock()
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := m.revokeOneByClientMsgID(ctx, req.SendID, sessionType, item); err != nil {
				log.ZWarn(ctx, "batch revoke item failed", err, "recvID", item.RecvID, "clientMsgID", item.ClientMsgID)
				mu.Lock()
				failedUserIDs = append(failedUserIDs, item.RecvID)
				mu.Unlock()
				return
			}
			mu.Lock()
			successCount++
			mu.Unlock()
		}()
	}
	wg.Wait()

	apiresp.GinSuccess(c, &apistruct.BatchRevokeMsgResp{
		SuccessCount:  successCount,
		FailedCount:   len(failedUserIDs),
		FailedUserIDs: failedUserIDs,
	})
}

func cloneAdminRPCCtx(c *gin.Context) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	ctx = mcontext.SetOpUserID(ctx, mcontext.GetOpUserID(c))
	ctx = mcontext.SetOperationID(ctx, mcontext.GetOperationID(c))
	ctx = mcontext.WithOpUserPlatformContext(ctx, mcontext.GetOpUserPlatform(c))
	ctx = authverify.WithIMAdminUserIDs(ctx, authverify.GetIMAdminUserIDs(c))
	return context.WithTimeout(ctx, 10*time.Minute)
}

func (m *MessageApi) revokeOneByClientMsgID(ctx context.Context, sendID string, sessionType int32, item apistruct.BatchRevokeMsgItem) error {
	conversationID := msgprocessor.GetConversationIDBySessionType(int(sessionType), sendID, item.RecvID)
	if conversationID == "" {
		return errs.ErrArgs.WrapMsg("unsupported sessionType")
	}
	info, err := m.findMsgByClientMsgIDRetry(ctx, conversationID, item.ClientMsgID)
	if err != nil {
		return err
	}
	if info.Revoke != nil {
		return nil
	}
	if info.Msg == nil || info.Msg.Seq < 1 {
		return errs.ErrRecordNotFound.WrapMsg("seq not assigned")
	}
	if info.Msg.ContentType == constant.MsgRevokeNotification {
		return nil
	}
	_, err = m.Client.RevokeMsg(ctx, &msg.RevokeMsgReq{
		ConversationID: conversationID,
		Seq:            info.Msg.Seq,
		UserID:         item.RecvID,
	})
	if revokeErrIsSuccess(err) {
		return nil
	}
	return err
}

func (m *MessageApi) findMsgByClientMsgIDRetry(ctx context.Context, conversationID, clientMsgID string) (*model.MsgInfoModel, error) {
	delay := batchRevokeFindBase
	var lastErr error
	for i := 0; i < batchRevokeFindTries; i++ {
		info, err := m.msgDoc.FindByClientMsgID(ctx, conversationID, clientMsgID)
		if err == nil && info != nil && info.Msg != nil && info.Msg.Seq > 0 {
			return info, nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = errs.ErrRecordNotFound.WrapMsg("msg not found by clientMsgID")
		}
		if i == batchRevokeFindTries-1 {
			break
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
		delay *= 2
		if delay > batchRevokeFindCap {
			delay = batchRevokeFindCap
		}
	}
	return nil, lastErr
}

func revokeErrIsSuccess(err error) bool {
	if err == nil {
		return true
	}
	return servererrs.ErrMsgAlreadyRevoke.Is(err)
}
