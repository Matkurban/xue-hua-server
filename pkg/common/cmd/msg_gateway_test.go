package cmd

import (
	"math"
	"testing"

	"server/v3/pkg/common/apiresp"
	"server/v3/pkg/util/jsonutil"
	"github.com/openimsdk/protocol/auth"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestName(t *testing.T) {
	resp := &apiresp.ApiResponse{
		ErrCode: 1234,
		ErrMsg:  "test",
		ErrDlt:  "4567",
		Data: &auth.GetUserTokenResp{
			Token:             "1234567",
			ExpireTimeSeconds: math.MaxInt64,
		},
	}
	data, err := resp.MarshalJSON()
	if err != nil {
		panic(err)
	}
	t.Log(string(data))

	var rReso apiresp.ApiResponse
	rReso.Data = &auth.GetUserTokenResp{}

	if err := jsonutil.JsonUnmarshal(data, &rReso); err != nil {
		panic(err)
	}

	t.Logf("%+v\n", rReso)

}

func TestName1(t *testing.T) {
	t.Log(bson.NewObjectID().String())
	t.Log(bson.NewObjectID().Hex())

}
