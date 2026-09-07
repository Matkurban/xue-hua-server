package mgo

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"server/v3/pkg/common/storage/model"
	"testing"
	"time"
)

//func Result[V any](val V, err error) V {
//	if err != nil {
//		panic(err)
//	}
//	return val
//}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func TestName(t *testing.T) {
	cli := Result(mongo.Connect(options.Client().ApplyURI("mongodb://server:openIM123@172.16.8.48:37017/server_v3?maxPoolSize=100").SetConnectTimeout(5 * time.Second)))
	coll := cli.Database("server_v3").Collection("version_test")
	tmp, err := NewVersionLog(coll)
	if err != nil {
		panic(err)
	}
	vl := tmp.(*VersionLogMgo)
	res, err := vl.incrVersionResult(context.Background(), "100", []string{"1000", "1001", "1003"}, model.VersionStateInsert)
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%+v", res)
}
