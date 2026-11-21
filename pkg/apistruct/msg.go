package apistruct

import (
	"server/v3/pkg/common/errs"
	"github.com/openimsdk/protocol/sdkws"
)

type PictureBaseInfo struct {
	UUID   string `mapstructure:"uuid"`
	Type   string `mapstructure:"type"   validate:"required"`
	Size   int64  `mapstructure:"size"`
	Width  int32  `mapstructure:"width"  validate:"required"`
	Height int32  `mapstructure:"height" validate:"required"`
	Url    string `mapstructure:"url"    validate:"required"`
}

type PictureElem struct {
	SourcePath      string          `mapstructure:"sourcePath"`
	SourcePicture   PictureBaseInfo `mapstructure:"sourcePicture"   validate:"required"`
	BigPicture      PictureBaseInfo `mapstructure:"bigPicture"      validate:"required"`
	SnapshotPicture PictureBaseInfo `mapstructure:"snapshotPicture" validate:"required"`
}

type SoundElem struct {
	UUID      string `mapstructure:"uuid"`
	SoundPath string `mapstructure:"soundPath"`
	SourceURL string `mapstructure:"sourceUrl" validate:"required"`
	DataSize  int64  `mapstructure:"dataSize"`
	Duration  int64  `mapstructure:"duration"  validate:"required,min=1"`
}

type VideoElem struct {
	VideoPath      string `mapstructure:"videoPath"`
	VideoUUID      string `mapstructure:"videoUUID"`
	VideoURL       string `mapstructure:"videoUrl"       validate:"required"`
	VideoType      string `mapstructure:"videoType"      validate:"required"`
	VideoSize      int64  `mapstructure:"videoSize"      validate:"required"`
	Duration       int64  `mapstructure:"duration"       validate:"required"`
	SnapshotPath   string `mapstructure:"snapshotPath"`
	SnapshotUUID   string `mapstructure:"snapshotUUID"`
	SnapshotSize   int64  `mapstructure:"snapshotSize"`
	SnapshotURL    string `mapstructure:"snapshotUrl"    validate:"required"`
	SnapshotWidth  int32  `mapstructure:"snapshotWidth"  validate:"required"`
	SnapshotHeight int32  `mapstructure:"snapshotHeight" validate:"required"`
}

type FileElem struct {
	FilePath  string `mapstructure:"filePath"`
	UUID      string `mapstructure:"uuid"`
	SourceURL string `mapstructure:"sourceUrl" validate:"required"`
	FileName  string `mapstructure:"fileName"  validate:"required"`
	FileSize  int64  `mapstructure:"fileSize"  validate:"required"`
}
type AtElem struct {
	Text         string     `mapstructure:"text"`
	AtUserList   []string   `mapstructure:"atUserList" validate:"required,max=1000"`
	AtUsersInfo  []*AtInfo  `json:"atUsersInfo"`
	QuoteMessage *MsgStruct `json:"quoteMessage"`
	IsAtSelf     bool       `mapstructure:"isAtSelf"`
}
type LocationElem struct {
	Description string  `mapstructure:"description"`
	Longitude   float64 `mapstructure:"longitude"   validate:"required"`
	Latitude    float64 `mapstructure:"latitude"    validate:"required"`
}

type CustomElem struct {
	Data        string `mapstructure:"data"        validate:"required"`
	Description string `mapstructure:"description"`
	Extension   string `mapstructure:"extension"`
}

type TextElem struct {
	Content string `json:"content" validate:"required"`
}

type MarkdownTextElem struct {
	Content string `mapstructure:"content" validate:"required"`
}

type StreamMsgElem struct {
	Type    string `mapstructure:"type" validate:"required"`
	Content string `mapstructure:"content" validate:"required"`
}

type RevokeElem struct {
	RevokeMsgClientID string `mapstructure:"revokeMsgClientID" validate:"required"`
}

// ModifyMsgContentReq updates persisted message Content JSON in-place (admin).
// MergeContent keys are shallow-merged into the existing content object.
type ModifyMsgContentReq struct {
	ConversationID string         `json:"conversationID" binding:"required"`
	ClientMsgID    string         `json:"clientMsgID" binding:"required"`
	MergeContent   map[string]any `json:"mergeContent" binding:"required"`
}

type ModifyMsgContentResp struct {
	Seq int64 `json:"seq"`
}

// ModifyMsgExReq merges fields into a stored message's Ex JSON.
// Only the message sender may update; used by burn-after-reading.
type ModifyMsgExReq struct {
	ConversationID string         `json:"conversationID" binding:"required"`
	ClientMsgID    string         `json:"clientMsgID" binding:"required"`
	MergeEx        map[string]any `json:"mergeEx" binding:"required"`
}

type ModifyMsgExResp struct {
	Seq int64  `json:"seq"`
	Ex  string `json:"ex"`
}

type QuoteElem struct {
	Text         string     `json:"text,omitempty"`
	QuoteMessage *MsgStruct `json:"quoteMessage,omitempty"`
}

type OANotificationElem struct {
	PictureElem         *PictureElem `mapstructure:"pictureElem"         json:"pictureElem"`
	SoundElem           *SoundElem   `mapstructure:"soundElem"           json:"soundElem"`
	VideoElem           *VideoElem   `mapstructure:"videoElem"           json:"videoElem"`
	FileElem            *FileElem    `mapstructure:"fileElem"            json:"fileElem"`
	NotificationName    string       `mapstructure:"notificationName"    json:"notificationName"    validate:"required"`
	NotificationFaceURL string       `mapstructure:"notificationFaceURL" json:"notificationFaceURL"`
	Text                string       `mapstructure:"text"                json:"text"                validate:"required"`
	Url                 string       `mapstructure:"url"                 json:"url"`
	Ex                  string       `mapstructure:"ex"                  json:"ex"`
	NotificationType    int32        `mapstructure:"notificationType"    json:"notificationType"    validate:"required"`
	MixType             int32        `mapstructure:"mixType"             json:"mixType"             validate:"gte=0,lte=5"`
}

type MessageRevoked struct {
	RevokerID       string `mapstructure:"revokerID"       json:"revokerID"       validate:"required"`
	RevokerRole     int32  `mapstructure:"revokerRole"     json:"revokerRole"     validate:"required"`
	ClientMsgID     string `mapstructure:"clientMsgID"     json:"clientMsgID"     validate:"required"`
	RevokerNickname string `mapstructure:"revokerNickname" json:"revokerNickname"`
	SessionType     int32  `mapstructure:"sessionType"     json:"sessionType"     validate:"required"`
	Seq             uint32 `mapstructure:"seq"             json:"seq"             validate:"required"`
}

// 一等公民业务消息 ContentType 常量。
//
// 选用 124-126：上游 protocol/constant 已占用 100-122（含 121/122
// ReactionMessage*），123 暂未占用但贴邻 reaction 段不利于演进；
// 124 起处于自定义消息段（110/119/120 之后）的空闲区。
const (
	CallSignal            = 124
	RedPacket             = 125
	RedPacketGrabNotify   = 126
	Conference            = 127
	MeetingReminder       = 128 // 会前提醒
	MeetingScheduleNotify = 129 // 预约变更/取消通知
)

// ValidateMessageContentType rejects retired IDs, which must never be reused.
// Other unknown types remain available to custom message protocols.
func ValidateMessageContentType(contentType int32) error {
	switch contentType {
	case 130, 131, 132:
		return errs.ErrArgs.WrapMsg("retired content type", "contentType", contentType)
	default:
		return nil
	}
}

// CallSignalConversation 通话信令中的轻量会话信息
type CallSignalConversation struct {
	ConversationID string `mapstructure:"conversationID" json:"conversationID"`
	UserID         string `mapstructure:"userID"         json:"userID"`
	GroupID        string `mapstructure:"groupID"        json:"groupID"`
	ShowName       string `mapstructure:"showName"       json:"showName"`
	FaceURL        string `mapstructure:"faceURL"        json:"faceURL"`
}

// CallSignalUser 通话信令中的轻量用户信息
type CallSignalUser struct {
	UserID   string `mapstructure:"userID"   json:"userID" validate:"required"`
	Nickname string `mapstructure:"nickname" json:"nickname"`
	FaceURL  string `mapstructure:"faceURL"  json:"faceURL"`
}

// CallSignalElem 通话信令消息体（ContentType=CallSignal）
type CallSignalElem struct {
	Conversation   CallSignalConversation `mapstructure:"conversation"   json:"conversation"`
	RoomID         string                 `mapstructure:"roomId"         json:"roomId"          validate:"required"`
	CallType       int32                  `mapstructure:"callType"       json:"callType"        validate:"required"`
	CallRoomType   int32                  `mapstructure:"callRoomType"   json:"callRoomType"    validate:"required"`
	CallSignalType int32                  `mapstructure:"callSignalType" json:"callSignalType"  validate:"required"`
	SenderUser     CallSignalUser         `mapstructure:"senderUser"     json:"senderUser"`
	RecvUser       []CallSignalUser       `mapstructure:"recvUser"       json:"recvUser"`
	Duration       *int64                 `mapstructure:"duration"       json:"duration,omitempty"`
}

// ConferenceElem 会议公告消息体（ContentType=Conference）
type ConferenceElem struct {
	RoomID    string         `mapstructure:"roomId"    json:"roomId"    validate:"required"`
	CallType  int32          `mapstructure:"callType"  json:"callType"  validate:"required"`
	Status    int32          `mapstructure:"status"    json:"status"    validate:"required"`
	Title     string         `mapstructure:"title"     json:"title,omitempty"`
	AdminUser CallSignalUser `mapstructure:"adminUser" json:"adminUser"`
	GroupID   string         `mapstructure:"groupID"   json:"groupID,omitempty"`
	StartedAt int64          `mapstructure:"startedAt" json:"startedAt"`
	EndedAt   *int64         `mapstructure:"endedAt"   json:"endedAt,omitempty"`
}

// RedPacketElem 红包消息体（ContentType=RedPacket）
type RedPacketElem struct {
	PacketID     string  `mapstructure:"packetID"     json:"packetID"     validate:"required"`
	Greeting     string  `mapstructure:"greeting"     json:"greeting"`
	TargetUserID string  `mapstructure:"targetUserID" json:"targetUserID,omitempty"`
	TotalAmount  float64 `mapstructure:"totalAmount"  json:"totalAmount"  validate:"required,gt=0"`
	PacketType   int32   `mapstructure:"packetType"   json:"packetType"`
	TotalCount   int32   `mapstructure:"totalCount"   json:"totalCount"   validate:"required,min=1"`
}

// RedPacketGrabNotifyElem 红包领取提示消息体（ContentType=RedPacketGrabNotify）
//
// 由 openim-chat 在抢红包成功后通过 admin /msg/send_msg 注入会话，
// 客户端只消费展示，不主动发送。
type RedPacketGrabNotifyElem struct {
	PacketID       string  `mapstructure:"packetID"       json:"packetID"       validate:"required"`
	GrabberID      string  `mapstructure:"grabberID"      json:"grabberID"      validate:"required"`
	GrabberName    string  `mapstructure:"grabberName"    json:"grabberName"`
	GrabberFaceURL string  `mapstructure:"grabberFaceURL" json:"grabberFaceURL"`
	SenderID       string  `mapstructure:"senderID"       json:"senderID"       validate:"required"`
	SenderName     string  `mapstructure:"senderName"     json:"senderName"`
	Amount         float64 `mapstructure:"amount"         json:"amount"         validate:"required"`
}

// MeetingReminderElem 会前提醒消息体（ContentType=MeetingReminder）
type MeetingReminderElem struct {
	MeetingID     string `mapstructure:"meetingID"     json:"meetingID"     validate:"required"`
	OccurrenceID  string `mapstructure:"occurrenceID"  json:"occurrenceID"  validate:"required"`
	Title         string `mapstructure:"title"         json:"title"`
	StartTime     int64  `mapstructure:"startTime"     json:"startTime"     validate:"required"`
	RemindMinutes int32  `mapstructure:"remindMinutes" json:"remindMinutes" validate:"required"`
	CallType      string `mapstructure:"callType"      json:"callType"`
}

// MeetingScheduleNotifyElem 预约变更/取消通知（ContentType=MeetingScheduleNotify）
type MeetingScheduleNotifyElem struct {
	Action         string `mapstructure:"action"         json:"action"         validate:"required"`
	MeetingID      string `mapstructure:"meetingID"      json:"meetingID"      validate:"required"`
	OccurrenceID   string `mapstructure:"occurrenceID"   json:"occurrenceID"`
	SeriesID       string `mapstructure:"seriesID"       json:"seriesID"`
	Title          string `mapstructure:"title"          json:"title"`
	StartTime      int64  `mapstructure:"startTime"      json:"startTime"`
	OldStartTime   int64  `mapstructure:"oldStartTime"   json:"oldStartTime"`
	EndTime        int64  `mapstructure:"endTime"        json:"endTime"`
	Timezone       string `mapstructure:"timezone"       json:"timezone"`
	Scope          string `mapstructure:"scope"          json:"scope"`
	OperatorUserID string `mapstructure:"operatorUserID" json:"operatorUserID"`
	HasPassword    bool   `mapstructure:"hasPassword"    json:"hasPassword"`
}

type MsgStruct struct {
	OfflinePush          *sdkws.OfflinePushInfo `json:"offlinePush,omitempty"`
	TextElem             *TextElem              `json:"textElem,omitempty"`
	PictureElem          *PictureElem           `json:"pictureElem,omitempty"`
	SoundElem            *SoundElem             `json:"soundElem,omitempty"`
	VideoElem            *VideoElem             `json:"videoElem,omitempty"`
	FileElem             *FileElem              `json:"fileElem,omitempty"`
	AtTextElem           *AtElem                `json:"atTextElem,omitempty"`
	LocationElem         *LocationElem          `json:"locationElem,omitempty"`
	CustomElem           *CustomElem            `json:"customElem,omitempty"`
	QuoteElem            *QuoteElem             `json:"quoteElem,omitempty"`
	ClientMsgID          string                 `json:"clientMsgID,omitempty"`
	ServerMsgID          string                 `json:"serverMsgID,omitempty"`
	SendID               string                 `json:"sendID,omitempty"`
	RecvID               string                 `json:"recvID,omitempty"`
	SenderNickname       string                 `json:"senderNickname,omitempty"`
	SenderFaceURL        string                 `json:"senderFaceUrl,omitempty"`
	GroupID              string                 `json:"groupID,omitempty"`
	Content              string                 `json:"content,omitempty"`
	AttachedInfo         string                 `json:"attachedInfo,omitempty"`
	Ex                   string                 `json:"ex,omitempty"`
	LocalEx              string                 `json:"localEx,omitempty"`
	CreateTime           int64                  `json:"createTime"`
	SendTime             int64                  `json:"sendTime"`
	Seq                  int64                  `json:"seq"`
	SessionType          int32                  `json:"sessionType"`
	MsgFrom              int32                  `json:"msgFrom"`
	ContentType          int32                  `json:"contentType"`
	SenderPlatformID     int32                  `json:"senderPlatformID"`
	Status               int32                  `json:"status"`
	IsRead               bool                   `json:"isRead"`
	IsReact              bool                   `json:"isReact,omitempty"`
	IsExternalExtensions bool                   `json:"isExternalExtensions,omitempty"`
}

type AtInfo struct {
	AtUserID      string `json:"atUserID,omitempty"`
	GroupNickname string `json:"groupNickname,omitempty"`
}
