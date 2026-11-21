package mw

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openimsdk/protocol/constant"
	"server/v3/pkg/common/apiresp"
	"server/v3/pkg/common/errs"
	"server/v3/pkg/common/log"
)

// CorsHandler gin cross-domain configuration.
func CorsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header(
			"Access-Control-Expose-Headers",
			"Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar",
		)
		c.Header("Access-Control-Max-Age", "172800")
		c.Header("Access-Control-Allow-Credentials", "false")
		c.Header("content-type", "application/json")
		if c.Request.Method == http.MethodOptions {
			c.JSON(http.StatusOK, "Options Request!")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GinParseOperationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {
			operationID := c.Request.Header.Get(constant.OperationID)
			if operationID == "" {
				err := errs.New("header must have operationID")
				apiresp.GinError(c, errs.ErrArgs.WrapMsg(err.Error()))
				c.Abort()
				return
			}
			c.Set(constant.OperationID, operationID)
		}
		c.Next()
	}
}

func GinPanicErr(c *gin.Context, err any) {
	log.ZPanic(c, "GinPanicErr panic", errs.ErrPanic(err))
	c.AbortWithStatus(http.StatusInternalServerError)
}
