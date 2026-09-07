package user

import (
	"context"

	"gitee.com/matkurban/openimsdk-tools/utils/datautil"
	pbuser "github.com/openimsdk/protocol/user"
	"server/v3/pkg/authverify"
	"server/v3/pkg/common/storage/model"
)

func (s *userServer) GetUserClientConfig(ctx context.Context, req *pbuser.GetUserClientConfigReq) (*pbuser.GetUserClientConfigResp, error) {
	if req.UserID != "" {
		if err := authverify.CheckAccess(ctx, req.UserID); err != nil {
			return nil, err
		}
		if _, err := s.db.GetUserByID(ctx, req.UserID); err != nil {
			return nil, err
		}
	}
	res, err := s.clientConfig.GetUserConfig(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	return &pbuser.GetUserClientConfigResp{Configs: res}, nil
}

func (s *userServer) SetUserClientConfig(ctx context.Context, req *pbuser.SetUserClientConfigReq) (*pbuser.SetUserClientConfigResp, error) {
	if err := authverify.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	if req.UserID != "" {
		if _, err := s.db.GetUserByID(ctx, req.UserID); err != nil {
			return nil, err
		}
	}
	if err := s.clientConfig.SetUserConfig(ctx, req.UserID, req.Configs); err != nil {
		return nil, err
	}
	return &pbuser.SetUserClientConfigResp{}, nil
}

func (s *userServer) DelUserClientConfig(ctx context.Context, req *pbuser.DelUserClientConfigReq) (*pbuser.DelUserClientConfigResp, error) {
	if err := authverify.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	if err := s.clientConfig.DelUserConfig(ctx, req.UserID, req.Keys); err != nil {
		return nil, err
	}
	return &pbuser.DelUserClientConfigResp{}, nil
}

func (s *userServer) PageUserClientConfig(ctx context.Context, req *pbuser.PageUserClientConfigReq) (*pbuser.PageUserClientConfigResp, error) {
	if err := authverify.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	total, res, err := s.clientConfig.GetUserConfigPage(ctx, req.UserID, req.Key, req.Pagination)
	if err != nil {
		return nil, err
	}
	return &pbuser.PageUserClientConfigResp{
		Total: total,
		Configs: datautil.Slice(res, func(e *model.ClientConfig) *pbuser.ClientConfig {
			return &pbuser.ClientConfig{
				UserID: e.UserID,
				Key:    e.Key,
				Value:  e.Value,
			}
		}),
	}, nil
}
