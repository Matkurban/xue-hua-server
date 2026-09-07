package msggateway

import (
	"gitee.com/matkurban/openimsdk-tools/apiresp"
	"gitee.com/matkurban/openimsdk-tools/log"
)

func httpError(ctx *UserConnContext, err error) {
	log.ZWarn(ctx, "ws connection error", err)
	apiresp.HttpError(ctx.RespWriter, err)
}
