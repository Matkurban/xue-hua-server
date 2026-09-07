package servererrs

const (
	NoError = 0 // No error

	NetworkError = 90004 // Network error
	DataError    = 90007 // Data error

	CallbackError = 80000

	ArgsError         = 1001 // Input parameter error
	NoPermissionError = 1002 // Insufficient permission

	// Account error codes.
	UserIDNotFoundError    = 1101 // UserID does not exist or is not registered
	RegisteredAlreadyError = 1102 // user is already registered

	// Group error codes.
	GroupIDNotFoundError  = 1201 // GroupID does not exist
	GroupIDExisted        = 1202 // GroupID already exists
	NotInGroupYetError    = 1203 // Not in the group yet
	DismissedAlreadyError = 1204 // Group has already been dismissed

	GroupRequestHandled = 1206

	// Relationship error codes.
	CanNotAddYourselfError   = 1301 // Cannot add yourself as a friend
	BlockedByPeer            = 1302 // Blocked by the peer
	NotPeersFriend           = 1303 // Not the peer's friend
	RelationshipAlreadyError = 1304 // Already in a friend relationship
	FriendRequestHandled     = 1305 // Friend request has already been handled

	MutedInGroup     = 1402 // Member muted in the group
	MutedGroup       = 1403 // Group is muted
	MsgAlreadyRevoke = 1404 // Message already revoked

	TokenInvalidError = 1502

	TokenKickedError   = 1506
	TokenNotExistError = 1507

	// Long connection gateway error codes.
	ConnOverMaxNumLimit  = 1601
	ConnArgsErr          = 1602
	PushMsgErr           = 1603
	IOSBackgroundPushErr = 1604
)
