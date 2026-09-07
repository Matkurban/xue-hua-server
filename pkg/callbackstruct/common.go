package callbackstruct

import (
	"gitee.com/matkurban/openimsdk-tools/errs"
	"server/v3/pkg/common/servererrs"
)

const (
	Next = 1
)

type CommonCallbackReq struct {
	SendID           string   `json:"sendID"`
	CallbackCommand  string   `json:"callbackCommand"`
	ServerMsgID      string   `json:"serverMsgID"`
	ClientMsgID      string   `json:"clientMsgID"`
	OperationID      string   `json:"operationID"`
	SenderNickname   string   `json:"senderNickname"`
	Content          string   `json:"content"`
	SenderFaceURL    string   `json:"faceURL"`
	Ex               string   `json:"ex"`
	AtUserIDList     []string `json:"atUserList"`
	SendTime         int64    `json:"sendTime"`
	CreateTime       int64    `json:"createTime"`
	SenderPlatformID int32    `json:"senderPlatformID"`
	SessionType      int32    `json:"sessionType"`
	MsgFrom          int32    `json:"msgFrom"`
	ContentType      int32    `json:"contentType"`
	Status           int32    `json:"status"`
	Seq              uint32   `json:"seq"`
}

func (c *CommonCallbackReq) GetCallbackCommand() string {
	return c.CallbackCommand
}

type CallbackReq interface {
	GetCallbackCommand() string
}

type CallbackResp interface {
	Parse() (err error)
}

type CommonCallbackResp struct {
	ActionCode int32  `json:"actionCode"`
	ErrCode    int32  `json:"errCode"`
	ErrMsg     string `json:"errMsg"`
	ErrDlt     string `json:"errDlt"`
	NextCode   int32  `json:"nextCode"`
}

func (c CommonCallbackResp) Parse() error {
	if c.ActionCode == servererrs.NoError && c.NextCode == Next {
		return errs.NewCodeError(int(c.ErrCode), c.ErrMsg).WithDetail(c.ErrDlt)
	}
	return nil
}

type UserStatusBaseCallback struct {
	CallbackCommand string `json:"callbackCommand"`
	OperationID     string `json:"operationID"`
	PlatformID      int    `json:"platformID"`
	Platform        string `json:"platform"`
}

func (c UserStatusBaseCallback) GetCallbackCommand() string {
	return c.CallbackCommand
}

type UserStatusCallbackReq struct {
	UserStatusBaseCallback
	UserID string `json:"userID"`
}

type UserStatusBatchCallbackReq struct {
	UserStatusBaseCallback
	UserIDList []string `json:"userIDList"`
}
