package user

import (
	"context"
	"time"

	"server/v3/pkg/common/errs"
	pbuser "github.com/openimsdk/protocol/user"
)

func (s *userServer) UserRegisterCount(ctx context.Context, req *pbuser.UserRegisterCountReq) (*pbuser.UserRegisterCountResp, error) {
	if req.Start > req.End {
		return nil, errs.ErrArgs.WrapMsg("start > end")
	}
	total, err := s.db.CountTotal(ctx, nil)
	if err != nil {
		return nil, err
	}
	start := time.UnixMilli(req.Start)
	before, err := s.db.CountTotal(ctx, &start)
	if err != nil {
		return nil, err
	}
	count, err := s.db.CountRangeEverydayTotal(ctx, start, time.UnixMilli(req.End))
	if err != nil {
		return nil, err
	}
	return &pbuser.UserRegisterCountResp{Total: total, Before: before, Count: count}, nil
}
