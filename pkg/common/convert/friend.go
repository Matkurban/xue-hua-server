package convert

import (
	"context"

	"github.com/openimsdk/protocol/relation"
	"server/v3/pkg/common/storage/model"
	"server/v3/pkg/notification/common_user"

	"server/v3/pkg/util/datautil"
	"github.com/openimsdk/protocol/sdkws"
)

func FriendsDB2Pb(ctx context.Context, friendsDB []*model.Friend, getUsers func(ctx context.Context, userIDs []string) (map[string]*sdkws.UserInfo, error)) (friendsPb []*sdkws.FriendInfo, err error) {
	if len(friendsDB) == 0 {
		return nil, nil
	}
	var userID []string
	for _, friendDB := range friendsDB {
		userID = append(userID, friendDB.FriendUserID)
	}

	users, err := getUsers(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, friend := range friendsDB {
		friendPb := &sdkws.FriendInfo{FriendUser: &sdkws.UserInfo{}}
		err := datautil.CopyStructFields(friendPb, friend)
		if err != nil {
			return nil, err
		}

		friendPb.FriendUser.UserID = users[friend.FriendUserID].UserID
		friendPb.FriendUser.Nickname = users[friend.FriendUserID].Nickname
		friendPb.FriendUser.FaceURL = users[friend.FriendUserID].FaceURL
		friendPb.FriendUser.Ex = users[friend.FriendUserID].Ex
		friendPb.CreateTime = friend.CreateTime.Unix()
		friendPb.IsPinned = friend.IsPinned
		friendsPb = append(friendsPb, friendPb)
	}
	return friendsPb, nil
}

func FriendOnlyDB2PbOnly(friendsDB []*model.Friend) []*relation.FriendInfoOnly {
	return datautil.Slice(friendsDB, func(f *model.Friend) *relation.FriendInfoOnly {
		return &relation.FriendInfoOnly{
			OwnerUserID:    f.OwnerUserID,
			FriendUserID:   f.FriendUserID,
			Remark:         f.Remark,
			CreateTime:     f.CreateTime.UnixMilli(),
			AddSource:      f.AddSource,
			OperatorUserID: f.OperatorUserID,
			Ex:             f.Ex,
			IsPinned:       f.IsPinned,
		}
	})
}

func FriendRequestDB2Pb(ctx context.Context, friendRequests []*model.FriendRequest, getUsers func(ctx context.Context, userIDs []string) (map[string]common_user.CommonUser, error)) ([]*sdkws.FriendRequest, error) {
	if len(friendRequests) == 0 {
		return nil, nil
	}
	userIDMap := make(map[string]struct{})
	for _, friendRequest := range friendRequests {
		userIDMap[friendRequest.ToUserID] = struct{}{}
		userIDMap[friendRequest.FromUserID] = struct{}{}
	}
	users, err := getUsers(ctx, datautil.Keys(userIDMap))
	if err != nil {
		return nil, err
	}
	res := make([]*sdkws.FriendRequest, 0, len(friendRequests))
	for _, friendRequest := range friendRequests {
		toUser := users[friendRequest.ToUserID]
		fromUser := users[friendRequest.FromUserID]
		res = append(res, &sdkws.FriendRequest{
			FromUserID:    friendRequest.FromUserID,
			FromNickname:  fromUser.GetNickname(),
			FromFaceURL:   fromUser.GetFaceURL(),
			ToUserID:      friendRequest.ToUserID,
			ToNickname:    toUser.GetNickname(),
			ToFaceURL:     toUser.GetFaceURL(),
			HandleResult:  friendRequest.HandleResult,
			ReqMsg:        friendRequest.ReqMsg,
			CreateTime:    friendRequest.CreateTime.UnixMilli(),
			HandlerUserID: friendRequest.HandlerUserID,
			HandleMsg:     friendRequest.HandleMsg,
			HandleTime:    friendRequest.HandleTime.UnixMilli(),
			Ex:            friendRequest.Ex,
		})
	}
	return res, nil
}
