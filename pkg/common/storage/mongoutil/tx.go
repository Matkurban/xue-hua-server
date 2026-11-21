package mongoutil

import (
	"context"

	"server/v3/pkg/common/storage/tx"
	"server/v3/pkg/common/errs"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// NewMongoTx inspects whether the MongoDB deployment supports multi-document transactions
// (replica set or sharded cluster) using the modern `hello` command.
func NewMongoTx(ctx context.Context, client *mongo.Client) (tx.Tx, error) {
	mtx := &mongoTx{client: client}
	if err := mtx.init(ctx); err != nil {
		return nil, err
	}
	return mtx, nil
}

// NewMongo returns a non-transactional tx.Tx fallback bound to client.
func NewMongo(client *mongo.Client) tx.Tx {
	return &mongoTx{client: client}
}

type mongoTx struct {
	client *mongo.Client
	txFn   func(context.Context, func(ctx context.Context) error) error
}

func (m *mongoTx) init(ctx context.Context) error {
	var res struct {
		SetName string `bson:"setName"`
		Msg     string `bson:"msg"`
	}
	if err := m.client.Database("admin").RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&res); err != nil {
		return errs.WrapMsg(err, "check whether mongo supports transactions via hello command")
	}
	// Replica sets set `setName`; mongos (sharded clusters) set `msg == "isdbgrid"`.
	if res.SetName == "" && res.Msg != "isdbgrid" {
		return nil
	}
	m.txFn = func(fnCtx context.Context, fn func(ctx context.Context) error) error {
		sess, err := m.client.StartSession()
		if err != nil {
			return errs.WrapMsg(err, "mongodb start session failed")
		}
		defer sess.EndSession(fnCtx)
		_, err = sess.WithTransaction(fnCtx, func(sessCtx context.Context) (any, error) {
			return nil, fn(sessCtx)
		})
		if err != nil {
			return errs.WrapMsg(err, "mongodb transaction failed")
		}
		return nil
	}
	return nil
}

func (m *mongoTx) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.txFn == nil {
		return fn(ctx)
	}
	return m.txFn(ctx, fn)
}
