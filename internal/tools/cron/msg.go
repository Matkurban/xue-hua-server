package cron

import (
	"fmt"
	"os"
	"time"

	"server/v3/pkg/common/log"
	"server/v3/pkg/common/mcontext"
	"github.com/openimsdk/protocol/msg"
)

func (c *cronServer) deleteMsg() {
	now := time.Now()
	deltime := now.Add(-time.Hour * 24 * time.Duration(c.config.CronTask.RetainChatRecords))
	operationID := fmt.Sprintf("cron_msg_%d_%d", os.Getpid(), deltime.UnixMilli())
	ctx := mcontext.SetOperationID(c.ctx, operationID)
	log.ZDebug(ctx, "Destruct chat records", "deltime", deltime, "timestamp", deltime.UnixMilli())
	const (
		deleteCount = 10000
		deleteLimit = 50
	)
	var count int
	for i := 1; i <= deleteCount; i++ {
		ctx := mcontext.SetOperationID(c.ctx, fmt.Sprintf("%s_%d", operationID, i))
		resp, err := c.msgClient.DestructMsgs(ctx, &msg.DestructMsgsReq{Timestamp: deltime.UnixMilli(), Limit: deleteLimit})
		if err != nil {
			log.ZError(ctx, "cron destruct chat records failed", err)
			break
		}
		count += int(resp.Count)
		if resp.Count < deleteLimit {
			break
		}
	}
	log.ZDebug(ctx, "cron destruct chat records end", "deltime", deltime, "cont", time.Since(now), "count", count)
}
