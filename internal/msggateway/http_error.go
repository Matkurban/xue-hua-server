package msggateway

import (
	"server/v3/pkg/common/apiresp"
	"server/v3/pkg/common/log"
)

func httpError(ctx *UserConnContext, err error) {
	log.ZWarn(ctx, "ws connection error", err)
	apiresp.HttpError(ctx.RespWriter, err)
}
