package mgo

import (
	"server/v3/pkg/common/errs"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func IsNotFound(err error) bool {
	return errs.Unwrap(err) == mongo.ErrNoDocuments
}
