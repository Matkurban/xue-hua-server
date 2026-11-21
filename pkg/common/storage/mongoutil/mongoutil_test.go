package mongoutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestValidateAndSetDefaults(t *testing.T) {
	t.Run("standalone defaults", func(t *testing.T) {
		cfg := &Config{
			Address:  []string{"127.0.0.1:27017"},
			Database: "server_v3",
		}
		if err := cfg.ValidateAndSetDefaults(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.MongoMode != StandaloneMode || cfg.MaxPoolSize != defaultMaxPoolSize || cfg.MaxRetry != defaultMaxRetry {
			t.Fatalf("unexpected defaults: %+v", cfg)
		}
	})

	t.Run("replicaSet sets retry and pool defaults", func(t *testing.T) {
		cfg := &Config{
			MongoMode: ReplicaSetMode,
			Database:  "server_v3",
			ReplicaSet: &ReplicaSetConfig{
				Name:  "rs0",
				Hosts: []string{"127.0.0.1:27017", "127.0.0.1:27018"},
			},
		}
		if err := cfg.ValidateAndSetDefaults(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.MaxRetry != defaultMaxRetry || cfg.MaxPoolSize != defaultMaxPoolSize {
			t.Fatalf("replicaSet mode must apply default MaxRetry/MaxPoolSize: %+v", cfg)
		}
	})
}

func TestBuildClientOptionsSpecialPassword(t *testing.T) {
	cfg := &Config{
		Address:     []string{"127.0.0.1:27017"},
		Database:    "server_v3",
		Username:    "admin",
		Password:    "p@ss:w/rd?&#%",
		MaxPoolSize: 50,
		WriteConcern: &WriteConcernConfig{
			W:        MajorityWriteConcern,
			J:        true,
			WTimeout: 5 * time.Second,
		},
		ReadPreference: &ReadPrefConfig{
			Mode: SecondaryPreferredMode,
		},
	}
	if err := cfg.ValidateAndSetDefaults(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	opts, err := buildClientOptions(cfg)
	if err != nil {
		t.Fatalf("buildClientOptions failed: %v", err)
	}
	if opts.Auth == nil || opts.Auth.Password != "p@ss:w/rd?&#%" || opts.Auth.AuthSource != "server_v3" {
		t.Fatalf("unexpected auth credential: %+v", opts.Auth)
	}
	if opts.Timeout == nil || *opts.Timeout != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", opts.Timeout)
	}
}

func TestShouldRetry(t *testing.T) {
	authErr := fmt.Errorf("wrapped: %w", mongo.CommandError{Code: 18, Message: "auth failed"})
	if shouldRetry(context.Background(), authErr) {
		t.Fatal("wrapped auth CommandError(18) must not retry")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if shouldRetry(ctx, fmt.Errorf("transient")) {
		t.Fatal("cancelled context must not retry")
	}
	if !shouldRetry(context.Background(), fmt.Errorf("transient network blip")) {
		t.Fatal("transient error should retry")
	}
}

func TestEmptyBatchNoOp(t *testing.T) {
	if err := InsertMany[struct{}](context.Background(), nil, nil); err != nil {
		t.Fatalf("InsertMany empty slice should no-op, got %v", err)
	}
	if err := BulkWrite(context.Background(), nil, nil); err != nil {
		t.Fatalf("BulkWrite empty slice should no-op, got %v", err)
	}
}
