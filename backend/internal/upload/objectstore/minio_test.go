package objectstore

import (
	"sync"
	"testing"
)

func TestMinIOStoreClientForConcurrentFirstUse(t *testing.T) {
	store := NewMinIOStore(MinIOConfig{
		Endpoint:  "http://localhost:9000",
		Bucket:    "easyinterview-test",
		AccessKey: "access-key",
		SecretKey: "secret-key",
	})

	const workers = 16
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	start := make(chan struct{})

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := store.clientFor()
			errs <- err
		}()
	}

	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("clientFor: %v", err)
		}
	}
	if store.client == nil {
		t.Fatal("expected cached minio client")
	}
}
