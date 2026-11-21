package mongoutil

import (
	"context"

	"gitee.com/matkurban/openimsdk-tools/db/pagination"
	"gitee.com/matkurban/openimsdk-tools/errs"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func isBasicType[T any]() bool {
	var t T
	switch any(t).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, []byte:
		return true
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64, *string, *[]byte:
		return true
	default:
		return false
	}
}

func InsertOne[T any](ctx context.Context, coll *mongo.Collection, val T, opts ...options.Lister[options.InsertOneOptions]) error {
	if _, err := coll.InsertOne(ctx, val, opts...); err != nil {
		return errs.WrapMsg(err, "mongo insert one")
	}
	return nil
}

// InsertMany inserts a typed slice directly using mongo-driver/v2's native `any` documents parameter
// (avoiding the v1 []T -> []any copy allocation) and safely no-ops on empty slices.
func InsertMany[T any](ctx context.Context, coll *mongo.Collection, val []T, opts ...options.Lister[options.InsertManyOptions]) error {
	if len(val) == 0 {
		return nil
	}
	if _, err := coll.InsertMany(ctx, val, opts...); err != nil {
		return errs.WrapMsg(err, "mongo insert many")
	}
	return nil
}

func UpdateOne(ctx context.Context, coll *mongo.Collection, filter any, update any, notMatchedErr bool, opts ...options.Lister[options.UpdateOneOptions]) error {
	res, err := coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return errs.WrapMsg(err, "mongo update one")
	}
	if notMatchedErr && res.MatchedCount == 0 {
		return errs.WrapMsg(mongo.ErrNoDocuments, "mongo update not matched")
	}
	return nil
}

func UpdateOneResult(ctx context.Context, coll *mongo.Collection, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	res, err := coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, errs.WrapMsg(err, "mongo update one")
	}
	return res, nil
}

func UpdateMany(ctx context.Context, coll *mongo.Collection, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	res, err := coll.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, errs.WrapMsg(err, "mongo update many")
	}
	return res, nil
}

func Find[T any](ctx context.Context, coll *mongo.Collection, filter any, opts ...options.Lister[options.FindOptions]) ([]T, error) {
	cur, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return nil, errs.WrapMsg(err, "mongo find")
	}
	defer cur.Close(ctx)
	return Decodes[T](ctx, cur)
}

func FindOne[T any](ctx context.Context, coll *mongo.Collection, filter any, opts ...options.Lister[options.FindOneOptions]) (res T, err error) {
	sr := coll.FindOne(ctx, filter, opts...)
	if err := sr.Err(); err != nil {
		return res, errs.WrapMsg(err, "mongo find one")
	}
	return DecodeOne[T](sr.Decode)
}

func FindOneAndUpdate[T any](ctx context.Context, coll *mongo.Collection, filter any, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) (res T, err error) {
	sr := coll.FindOneAndUpdate(ctx, filter, update, opts...)
	if err := sr.Err(); err != nil {
		return res, errs.WrapMsg(err, "mongo find one and update")
	}
	return DecodeOne[T](sr.Decode)
}

func FindPage[T any](ctx context.Context, coll *mongo.Collection, filter any, page pagination.Pagination, opts ...options.Lister[options.FindOptions]) (int64, []T, error) {
	count, err := Count(ctx, coll, filter)
	if err != nil {
		return 0, nil, errs.WrapMsg(err, "mongo failed to count documents in collection")
	}
	if count == 0 || page == nil {
		return count, nil, nil
	}
	skip := int64(page.GetPageNumber()-1) * int64(page.GetShowNumber())
	if skip < 0 || skip >= count || page.GetShowNumber() <= 0 {
		return count, nil, nil
	}
	findOpts := make([]options.Lister[options.FindOptions], 0, len(opts)+1)
	findOpts = append(findOpts, opts...)
	findOpts = append(findOpts, options.Find().SetSkip(skip).SetLimit(int64(page.GetShowNumber())))
	res, err := Find[T](ctx, coll, filter, findOpts...)
	if err != nil {
		return 0, nil, err
	}
	return count, res, nil
}

func FindPageOnly[T any](ctx context.Context, coll *mongo.Collection, filter any, page pagination.Pagination, opts ...options.Lister[options.FindOptions]) ([]T, error) {
	if page == nil {
		return nil, nil
	}
	skip := int64(page.GetPageNumber()-1) * int64(page.GetShowNumber())
	if skip < 0 || page.GetShowNumber() <= 0 {
		return nil, nil
	}
	findOpts := make([]options.Lister[options.FindOptions], 0, len(opts)+1)
	findOpts = append(findOpts, opts...)
	findOpts = append(findOpts, options.Find().SetSkip(skip).SetLimit(int64(page.GetShowNumber())))
	return Find[T](ctx, coll, filter, findOpts...)
}

func Count(ctx context.Context, coll *mongo.Collection, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	count, err := coll.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errs.WrapMsg(err, "mongo count")
	}
	return count, nil
}

func Exist(ctx context.Context, coll *mongo.Collection, filter any, opts ...options.Lister[options.CountOptions]) (bool, error) {
	countOpts := make([]options.Lister[options.CountOptions], 0, len(opts)+1)
	countOpts = append(countOpts, opts...)
	countOpts = append(countOpts, options.Count().SetLimit(1))
	count, err := Count(ctx, coll, filter, countOpts...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func DeleteOne(ctx context.Context, coll *mongo.Collection, filter any, opts ...options.Lister[options.DeleteOneOptions]) error {
	if _, err := coll.DeleteOne(ctx, filter, opts...); err != nil {
		return errs.WrapMsg(err, "mongo delete one")
	}
	return nil
}

func DeleteOneResult(ctx context.Context, coll *mongo.Collection, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	res, err := coll.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return nil, errs.WrapMsg(err, "mongo delete one")
	}
	return res, nil
}

func DeleteMany(ctx context.Context, coll *mongo.Collection, filter any, opts ...options.Lister[options.DeleteManyOptions]) error {
	if _, err := coll.DeleteMany(ctx, filter, opts...); err != nil {
		return errs.WrapMsg(err, "mongo delete many")
	}
	return nil
}

func DeleteManyResult(ctx context.Context, coll *mongo.Collection, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	res, err := coll.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return nil, errs.WrapMsg(err, "mongo delete many")
	}
	return res, nil
}

func Aggregate[T any](ctx context.Context, coll *mongo.Collection, pipeline any, opts ...options.Lister[options.AggregateOptions]) ([]T, error) {
	cur, err := coll.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, errs.WrapMsg(err, "mongo aggregate")
	}
	defer cur.Close(ctx)
	return Decodes[T](ctx, cur)
}

func Decodes[T any](ctx context.Context, cur *mongo.Cursor) ([]T, error) {
	var res []T
	if isBasicType[T]() {
		var temp []map[string]T
		if err := cur.All(ctx, &temp); err != nil {
			return nil, errs.WrapMsg(err, "mongo decodes")
		}
		res = make([]T, 0, len(temp))
		for _, m := range temp {
			if len(m) != 1 {
				return nil, errs.ErrInternalServer.WrapMsg("mongo find result len(m) != 1")
			}
			for _, t := range m {
				res = append(res, t)
			}
		}
	} else {
		if err := cur.All(ctx, &res); err != nil {
			return nil, errs.WrapMsg(err, "mongo all")
		}
	}
	return res, nil
}

func DecodeOne[T any](decoder func(v any) error) (res T, err error) {
	if isBasicType[T]() {
		var temp map[string]T
		if err = decoder(&temp); err != nil {
			err = errs.WrapMsg(err, "mongo decodes one")
			return
		}
		if len(temp) != 1 {
			err = errs.ErrInternalServer.WrapMsg("mongo find result len(m) != 1")
			return
		}
		for _, v := range temp {
			res = v
		}
	} else {
		if err = decoder(&res); err != nil {
			err = errs.WrapMsg(err, "mongo decoder")
			return
		}
	}
	return
}

func Ignore[T any](_ T, err error) error {
	return err
}

func IgnoreWarp[T any](_ T, err error) error {
	if err != nil {
		return errs.Wrap(err)
	}
	return nil
}

func IncrVersion(dbs ...func() error) error {
	for _, fn := range dbs {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func BulkWrite(ctx context.Context, coll *mongo.Collection, models []mongo.WriteModel, opts ...options.Lister[options.BulkWriteOptions]) error {
	if len(models) == 0 {
		return nil
	}
	if _, err := coll.BulkWrite(ctx, models, opts...); err != nil {
		return errs.WrapMsg(err, "mongo bulk write")
	}
	return nil
}
