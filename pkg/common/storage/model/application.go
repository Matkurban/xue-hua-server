package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Application struct {
	CreateTime time.Time     `bson:"create_time"`
	Platform   string        `bson:"platform"`
	Version    string        `bson:"version"`
	Url        string        `bson:"url"`
	Text       string        `bson:"text"`
	ID         bson.ObjectID `bson:"_id"`
	Hot        bool          `bson:"hot"`
	Force      bool          `bson:"force"`
	Latest     bool          `bson:"latest"`
}
