package batcher

import (
	"context"
	"fmt"
	"gitee.com/matkurban/openimsdk-tools/utils/stringutil"
	"testing"
	"time"
)

func TestBatcher(t *testing.T) {
	config := Config{
		size:     1000,
		buffer:   10,
		worker:   10,
		interval: 5 * time.Millisecond,
	}

	b := New[string](
		WithSize(config.size),
		WithBuffer(config.buffer),
		WithWorker(config.worker),
		WithInterval(config.interval),
		WithSyncWait(true),
	)

	// Mock Do function to simply print values for demonstration
	b.Do = func(ctx context.Context, channelID int, vals *Msg[string]) {
		t.Logf("Channel %d Processed batch: %v", channelID, vals)
	}
	b.OnComplete = func(lastMessage *string, totalCount int) {
		t.Logf("Completed processing with last message: %v, total count: %d", *lastMessage, totalCount)
	}
	b.Sharding = func(key string) int {
		hashCode := stringutil.GetHashCode(key)
		return int(hashCode) % config.worker
	}
	b.Key = func(data *string) string {
		return *data
	}

	err := b.Start()
	if err != nil {
		t.Fatal(err)
	}

	// Test normal data processing
	for i := 0; i < 10000; i++ {
		data := "data" + fmt.Sprintf("%d", i)
		if err := b.Put(context.Background(), &data); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Duration(1) * time.Second)
	start := time.Now()
	// Wait for all processing to finish
	b.Close()

	elapsed := time.Since(start)
	t.Logf("Close took %s", elapsed)

	if len(b.data) != 0 {
		t.Error("Data channel should be empty after closing")
	}
}

func TestSchedulerExternalClose(t *testing.T) {
	b := New[string](
		WithSize(10),
		WithBuffer(2),
		WithWorker(1),
		WithInterval(50*time.Millisecond),
	)
	b.Do = func(ctx context.Context, channelID int, vals *Msg[string]) {}
	b.OnComplete = func(lastMessage *string, totalCount int) {}
	b.Sharding = func(key string) int { return 0 }
	b.Key = func(data *string) string { return *data }

	if err := b.Start(); err != nil {
		t.Fatal(err)
	}
	// Simulate an external close of b.data while the scheduler is running.
	// Without the externallyClosed guard this panics: close of closed channel.
	close(b.data)
	done := make(chan struct{})
	go func() {
		b.wait.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not exit after external close")
	}
}
