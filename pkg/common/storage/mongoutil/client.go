package mongoutil

import (
	"context"
	"errors"
	"time"

	"gitee.com/matkurban/openimsdk-tools/db/tx"
	"gitee.com/matkurban/openimsdk-tools/errs"
	"gitee.com/matkurban/openimsdk-tools/log"
	"gitee.com/matkurban/openimsdk-tools/mw/specialerror"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/v2/tag"
)

const (
	defaultMaxPoolSize = 100
	defaultMaxRetry    = 3
)

// Mongo deployment modes.
const (
	StandaloneMode = "standalone"
	ReplicaSetMode = "replicaSet"
)

// ReadPreference modes.
const (
	PrimaryMode            = "primary"
	PrimaryPreferredMode   = "primaryPreferred"
	SecondaryMode          = "secondary"
	SecondaryPreferredMode = "secondaryPreferred"
	NearestMode            = "nearest"
)

// WriteConcern presets.
const (
	MajorityWriteConcern = "majority"
)

// ReadConcern levels.
const (
	LocalReadConcern        = "local"
	AvailableReadConcern    = "available"
	MajorityReadConcern     = "majority"
	LinearizableReadConcern = "linearizable"
	SnapshotReadConcern     = "snapshot"
)

func init() {
	if err := specialerror.AddReplace(mongo.ErrNoDocuments, errs.ErrRecordNotFound); err != nil {
		panic(err)
	}
}

// Config represents MongoDB connection and topology configuration.
type Config struct {
	Uri         string
	Address     []string
	Database    string
	Username    string
	Password    string
	AuthSource  string
	MaxPoolSize int
	MaxRetry    int
	MongoMode   string // "standalone" (default) or "replicaSet"

	ReplicaSet     *ReplicaSetConfig
	ReadPreference *ReadPrefConfig
	WriteConcern   *WriteConcernConfig
}

type ReplicaSetConfig struct {
	Name         string        `json:"name" yaml:"name" config:"allowempty"`
	Hosts        []string      `json:"hosts" yaml:"hosts" config:"allowempty"`
	ReadConcern  string        `json:"readConcern" yaml:"readConcern" config:"allowempty"`
	MaxStaleness time.Duration `json:"maxStaleness" yaml:"maxStaleness" config:"allowempty"`
}

type ReadPrefConfig struct {
	Mode         string              `json:"mode" yaml:"mode" config:"allowempty"`
	TagSets      []map[string]string `json:"tagSets" yaml:"tagSets" config:"allowempty"`
	MaxStaleness time.Duration       `json:"maxStaleness" yaml:"maxStaleness" config:"allowempty"`
}

type WriteConcernConfig struct {
	W        any           `json:"w" yaml:"w" config:"allowempty"`
	J        bool          `json:"j" yaml:"j" config:"allowempty"`
	WTimeout time.Duration `json:"wtimeout" yaml:"wtimeout" config:"allowempty"`
}

// Client wraps a MongoDB v2 database handle and transaction runner.
type Client struct {
	tx tx.Tx
	db *mongo.Database
}

func (c *Client) GetDB() *mongo.Database {
	return c.db
}

func (c *Client) GetTx() tx.Tx {
	return c.tx
}

func (c *Client) GetClient() *mongo.Client {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Client()
}

func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Client().Disconnect(ctx)
}

// ValidateAndSetDefaults validates the configuration and applies safe defaults for both standalone and replicaSet modes.
func (c *Config) ValidateAndSetDefaults() error {
	if c == nil {
		return errs.New("mongo config is nil").Wrap()
	}
	if c.MongoMode == "" {
		c.MongoMode = StandaloneMode
	}
	if c.Database == "" {
		return errs.New("mongo database is required").Wrap()
	}
	if c.MaxPoolSize <= 0 {
		c.MaxPoolSize = defaultMaxPoolSize
	}
	if c.MaxRetry <= 0 {
		c.MaxRetry = defaultMaxRetry
	}
	switch c.MongoMode {
	case ReplicaSetMode:
		if c.ReplicaSet == nil {
			return errs.New("replicaSet configuration required for replicaSet mode").Wrap()
		}
		if c.ReplicaSet.Name == "" {
			return errs.New("replicaSet name required for replicaSet mode").Wrap()
		}
		if c.Uri == "" && len(c.ReplicaSet.Hosts) == 0 && len(c.Address) == 0 {
			return errs.New("replicaSet hosts, address, or uri required for replicaSet mode").Wrap()
		}
	case StandaloneMode:
		if c.Uri == "" && len(c.Address) == 0 {
			return errs.New("either uri or address must be provided for standalone mode").Wrap()
		}
	default:
		return errs.New("unsupported mongo mode", "mode", c.MongoMode).Wrap()
	}
	return nil
}

// NewMongoDB initializes a MongoDB v2 connection with retry and transaction support.
func NewMongoDB(ctx context.Context, config *Config) (*Client, error) {
	if err := config.ValidateAndSetDefaults(); err != nil {
		return nil, err
	}

	opts, err := buildClientOptions(config)
	if err != nil {
		return nil, err
	}

	var cli *mongo.Client
	for attempt := 0; attempt < config.MaxRetry; attempt++ {
		cli, err = connectMongo(ctx, opts)
		if err == nil {
			break
		}
		if !shouldRetry(ctx, err) || attempt == config.MaxRetry-1 {
			break
		}
		log.ZError(ctx, "failed to connect to MongoDB, retrying", err, "attempt", attempt+1)
		timer := time.NewTimer(500 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errs.WrapMsg(ctx.Err(), "mongodb connect cancelled")
		case <-timer.C:
		}
	}
	if err != nil {
		return nil, errs.WrapMsg(err, "failed to connect to MongoDB", "database", config.Database, "mode", config.MongoMode)
	}

	mtx, err := NewMongoTx(ctx, cli)
	if err != nil {
		_ = cli.Disconnect(context.Background())
		return nil, err
	}
	return &Client{
		tx: mtx,
		db: cli.Database(config.Database),
	}, nil
}

// Check tests MongoDB connectivity once without retries and cleanly disconnects.
func Check(ctx context.Context, config *Config) error {
	if config == nil {
		return errs.New("mongo config is nil").Wrap()
	}
	checkCfg := *config
	checkCfg.MaxRetry = 1
	client, err := NewMongoDB(ctx, &checkCfg)
	if err != nil {
		return errs.WrapMsg(err, "MongoDB health check failed", "database", checkCfg.Database, "mode", checkCfg.MongoMode)
	}
	defer func() {
		_ = client.Close(ctx)
	}()
	return nil
}

func buildClientOptions(config *Config) (*options.ClientOptions, error) {
	opts := options.Client()

	if config.Uri != "" {
		opts.ApplyURI(config.Uri)
	} else {
		hosts := config.Address
		if config.MongoMode == ReplicaSetMode && config.ReplicaSet != nil && len(config.ReplicaSet.Hosts) > 0 {
			hosts = config.ReplicaSet.Hosts
		}
		opts.SetHosts(hosts)
		if config.Username != "" && config.Password != "" {
			authSource := config.AuthSource
			if authSource == "" {
				authSource = config.Database
			}
			opts.SetAuth(options.Credential{
				Username:   config.Username,
				Password:   config.Password,
				AuthSource: authSource,
			})
		}
	}

	if config.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(uint64(config.MaxPoolSize))
	}

	if config.MongoMode == ReplicaSetMode && config.ReplicaSet != nil {
		if config.ReplicaSet.Name != "" {
			opts.SetReplicaSet(config.ReplicaSet.Name)
		}
		if config.ReplicaSet.ReadConcern != "" {
			opts.SetReadConcern(buildReadConcern(config.ReplicaSet.ReadConcern))
		}
	}

	if config.ReadPreference != nil && config.ReadPreference.Mode != "" {
		rp, err := buildReadPreference(config.ReadPreference)
		if err != nil {
			return nil, errs.WrapMsg(err, "invalid mongo readPreference")
		}
		opts.SetReadPreference(rp)
	}

	if config.WriteConcern != nil && (config.WriteConcern.W != nil || config.WriteConcern.J) {
		opts.SetWriteConcern(buildWriteConcern(config.WriteConcern))
	}
	if config.WriteConcern != nil && config.WriteConcern.WTimeout > 0 {
		opts.SetTimeout(config.WriteConcern.WTimeout)
	}

	return opts, nil
}

func buildReadPreference(config *ReadPrefConfig) (*readpref.ReadPref, error) {
	var mode readpref.Mode
	switch config.Mode {
	case PrimaryMode:
		mode = readpref.PrimaryMode
	case PrimaryPreferredMode:
		mode = readpref.PrimaryPreferredMode
	case SecondaryMode:
		mode = readpref.SecondaryMode
	case SecondaryPreferredMode:
		mode = readpref.SecondaryPreferredMode
	case NearestMode:
		mode = readpref.NearestMode
	default:
		mode = readpref.PrimaryMode
	}

	var opts []readpref.Option
	if len(config.TagSets) > 0 {
		opts = append(opts, readpref.WithTagSets(tag.NewTagSetsFromMaps(config.TagSets)...))
	}
	if config.MaxStaleness > 0 {
		opts = append(opts, readpref.WithMaxStaleness(config.MaxStaleness))
	}
	return readpref.New(mode, opts...)
}

func buildWriteConcern(config *WriteConcernConfig) *writeconcern.WriteConcern {
	wc := &writeconcern.WriteConcern{}
	switch w := config.W.(type) {
	case int:
		wc.W = w
	case string:
		if w != "" {
			wc.W = w
		}
	}
	if config.J {
		j := true
		wc.Journal = &j
	}
	return wc
}

func buildReadConcern(level string) *readconcern.ReadConcern {
	switch level {
	case LocalReadConcern:
		return readconcern.Local()
	case AvailableReadConcern:
		return readconcern.Available()
	case MajorityReadConcern:
		return readconcern.Majority()
	case LinearizableReadConcern:
		return readconcern.Linearizable()
	case SnapshotReadConcern:
		return readconcern.Snapshot()
	default:
		return readconcern.Local()
	}
}

func connectMongo(ctx context.Context, opts *options.ClientOptions) (*mongo.Client, error) {
	cli, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}
	if err := cli.Ping(ctx, nil); err != nil {
		_ = cli.Disconnect(context.Background())
		return nil, err
	}
	return cli, nil
}

func shouldRetry(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		// Do not retry authorization (13) or authentication (18) failures.
		return cmdErr.Code != 13 && cmdErr.Code != 18
	}
	return true
}
