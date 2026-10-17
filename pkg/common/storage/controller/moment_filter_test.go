package controller

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"server/v3/pkg/common/storage/model"
)

func i32(v int32) *int32 { return &v }

func TestBuildAdminMomentFilter(t *testing.T) {
	f, err := BuildAdminMomentFilter(AdminMomentFilter{})
	if err != nil || f["status"].(bson.M)["$ne"] != model.MomentStatusDeleted || len(f) != 1 {
		t.Fatalf("default: %v %v", f, err)
	}
	f, err = BuildAdminMomentFilter(AdminMomentFilter{UserID: "u1", Keyword: "a.b(", StartTime: 1000, EndTime: 2000, Status: i32(2)})
	if err != nil {
		t.Fatal(err)
	}
	if f["user_id"] != "u1" || f["status"] != model.MomentStatusBanned {
		t.Errorf("%v", f)
	}
	if f["content"].(bson.M)["$regex"] != `a\.b\(` {
		t.Errorf("keyword not escaped: %v", f["content"])
	}
	r := f["create_time"].(bson.M)
	if !r["$gte"].(time.Time).Equal(time.UnixMilli(1000)) || !r["$lte"].(time.Time).Equal(time.UnixMilli(2000)) {
		t.Errorf("%v", r)
	}
	// normal status must include legacy docs without the field
	f, _ = BuildAdminMomentFilter(AdminMomentFilter{Status: i32(0)})
	if _, ok := f["status"].(bson.M)["$nin"]; !ok {
		t.Errorf("normal should use $nin: %v", f)
	}
	for _, bad := range []AdminMomentFilter{{StartTime: 5, EndTime: 1}, {StartTime: -1}, {Status: i32(9)}} {
		if _, err := BuildAdminMomentFilter(bad); err == nil {
			t.Errorf("expected error for %+v", bad)
		}
	}
	f, _ = BuildAdminMomentFilter(AdminMomentFilter{StartTime: 1000})
	if _, ok := f["create_time"].(bson.M)["$lte"]; ok {
		t.Error("open end expected")
	}
}
