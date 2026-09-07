package servererrs

import "gitee.com/matkurban/openimsdk-tools/errs"

var (
	ErrNetwork = errs.NewCodeError(NetworkError, "NetworkError")

	ErrCallbackContinue = errs.NewCodeError(CallbackError, "ErrCallbackContinue")

	ErrArgs         = errs.NewCodeError(ArgsError, "ArgsError")
	ErrNoPermission = errs.NewCodeError(NoPermissionError, "NoPermissionError")

	ErrUserIDNotFound  = errs.NewCodeError(UserIDNotFoundError, "UserIDNotFoundError")
	ErrGroupIDNotFound = errs.NewCodeError(GroupIDNotFoundError, "GroupIDNotFoundError")
	ErrGroupIDExisted  = errs.NewCodeError(GroupIDExisted, "GroupIDExisted")

	ErrNotInGroupYet     = errs.NewCodeError(NotInGroupYetError, "NotInGroupYetError")
	ErrDismissedAlready  = errs.NewCodeError(DismissedAlreadyError, "DismissedAlreadyError")
	ErrRegisteredAlready = errs.NewCodeError(RegisteredAlreadyError, "RegisteredAlreadyError")

	ErrGroupRequestHandled = errs.NewCodeError(GroupRequestHandled, "GroupRequestHandled")

	ErrData = errs.NewCodeError(DataError, "DataError")

	ErrTokenInvalid = errs.NewCodeError(TokenInvalidError, "TokenInvalidError") //

	ErrTokenKicked   = errs.NewCodeError(TokenKickedError, "TokenKickedError")
	ErrTokenNotExist = errs.NewCodeError(TokenNotExistError, "TokenNotExistError") //

	ErrCanNotAddYourself    = errs.NewCodeError(CanNotAddYourselfError, "CanNotAddYourselfError")
	ErrBlockedByPeer        = errs.NewCodeError(BlockedByPeer, "BlockedByPeer")
	ErrNotPeersFriend       = errs.NewCodeError(NotPeersFriend, "NotPeersFriend")
	ErrRelationshipAlready  = errs.NewCodeError(RelationshipAlreadyError, "RelationshipAlreadyError")
	ErrFriendRequestHandled = errs.NewCodeError(FriendRequestHandled, "FriendRequestHandled")

	ErrMutedInGroup     = errs.NewCodeError(MutedInGroup, "MutedInGroup")
	ErrMutedGroup       = errs.NewCodeError(MutedGroup, "MutedGroup")
	ErrMsgAlreadyRevoke = errs.NewCodeError(MsgAlreadyRevoke, "MsgAlreadyRevoke")

	ErrConnOverMaxNumLimit = errs.NewCodeError(ConnOverMaxNumLimit, "ConnOverMaxNumLimit")

	ErrConnArgsErr          = errs.NewCodeError(ConnArgsErr, "args err, need token, sendID, platformID")
	ErrPushMsgErr           = errs.NewCodeError(PushMsgErr, "push msg err")
	ErrIOSBackgroundPushErr = errs.NewCodeError(IOSBackgroundPushErr, "ios background push err")
)
