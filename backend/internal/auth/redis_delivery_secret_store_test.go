package auth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type memoryDeliverySecretRedis struct {
	mu     sync.Mutex
	values map[string]string
	ttls   map[string]time.Duration
	setErr error
	getErr error
	delErr error
}

type memoryDeliverySecretRedisClient struct {
	backend *memoryDeliverySecretRedis
}

func newMemoryDeliverySecretRedisClient(backend *memoryDeliverySecretRedis) *memoryDeliverySecretRedisClient {
	if backend.values == nil {
		backend.values = map[string]string{}
	}
	if backend.ttls == nil {
		backend.ttls = map[string]time.Duration{}
	}
	return &memoryDeliverySecretRedisClient{backend: backend}
}

func (c *memoryDeliverySecretRedisClient) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	c.backend.mu.Lock()
	defer c.backend.mu.Unlock()
	if c.backend.setErr != nil {
		return c.backend.setErr
	}
	c.backend.values[key] = value
	c.backend.ttls[key] = ttl
	return nil
}

func (c *memoryDeliverySecretRedisClient) Get(_ context.Context, key string) (string, bool, error) {
	c.backend.mu.Lock()
	defer c.backend.mu.Unlock()
	if c.backend.getErr != nil {
		return "", false, c.backend.getErr
	}
	value, ok := c.backend.values[key]
	return value, ok, nil
}

func (c *memoryDeliverySecretRedisClient) Del(_ context.Context, key string) error {
	c.backend.mu.Lock()
	defer c.backend.mu.Unlock()
	if c.backend.delErr != nil {
		return c.backend.delErr
	}
	delete(c.backend.values, key)
	delete(c.backend.ttls, key)
	return nil
}

func TestRedisDeliverySecretStoreEncryptsAndSharesAcrossClients(t *testing.T) {
	ctx := context.Background()
	backend := &memoryDeliverySecretRedis{}
	producer, err := NewRedisDeliverySecretStoreWithClient(newMemoryDeliverySecretRedisClient(backend), "challenge-pepper")
	if err != nil {
		t.Fatalf("NewRedisDeliverySecretStoreWithClient producer: %v", err)
	}
	consumer, err := NewRedisDeliverySecretStoreWithClient(newMemoryDeliverySecretRedisClient(backend), "challenge-pepper")
	if err != nil {
		t.Fatalf("NewRedisDeliverySecretStoreWithClient consumer: %v", err)
	}

	const ref = "auth_challenge:challenge-cross-instance"
	const code = "123456"
	if err := producer.PutDeliverySecret(ctx, ref, code, ChallengeTTL); err != nil {
		t.Fatalf("PutDeliverySecret: %v", err)
	}

	backend.mu.Lock()
	if len(backend.values) != 1 {
		backend.mu.Unlock()
		t.Fatalf("stored values = %d, want 1", len(backend.values))
	}
	var storedKey, storedValue string
	for key, value := range backend.values {
		storedKey, storedValue = key, value
	}
	storedTTL := backend.ttls[storedKey]
	backend.mu.Unlock()

	if !strings.HasPrefix(storedKey, "easyinterview:auth:delivery-secret:v1:") {
		t.Fatalf("stored key missing namespace: %q", storedKey)
	}
	for _, forbidden := range []string{ref, "challenge-cross-instance", code} {
		if strings.Contains(storedKey, forbidden) || strings.Contains(storedValue, forbidden) {
			t.Fatalf("Redis material leaked %q", forbidden)
		}
	}
	if storedTTL != ChallengeTTL {
		t.Fatalf("stored TTL = %s, want %s", storedTTL, ChallengeTTL)
	}

	got, ok, err := consumer.GetDeliverySecret(ctx, ref)
	if err != nil {
		t.Fatalf("GetDeliverySecret: %v", err)
	}
	if !ok || got != code {
		t.Fatalf("GetDeliverySecret = %q, %t; want %q, true", got, ok, code)
	}
	if err := consumer.DeleteDeliverySecret(ctx, ref); err != nil {
		t.Fatalf("DeleteDeliverySecret: %v", err)
	}
	if _, ok, err := producer.GetDeliverySecret(ctx, ref); err != nil || ok {
		t.Fatalf("GetDeliverySecret after delete ok=%t err=%v", ok, err)
	}
}

func TestRedisDeliverySecretStoreFailsClosedWithSafeErrors(t *testing.T) {
	ctx := context.Background()
	ref := "auth_challenge:challenge-private"
	code := "654321"
	providerErr := errors.New("redis://user:password@private-host:6379 code=654321")

	for _, tc := range []struct {
		name    string
		backend *memoryDeliverySecretRedis
		act     func(*RedisDeliverySecretStore) error
	}{
		{
			name:    "set",
			backend: &memoryDeliverySecretRedis{setErr: providerErr},
			act: func(store *RedisDeliverySecretStore) error {
				return store.PutDeliverySecret(ctx, ref, code, ChallengeTTL)
			},
		},
		{
			name:    "get",
			backend: &memoryDeliverySecretRedis{getErr: providerErr},
			act: func(store *RedisDeliverySecretStore) error {
				_, _, err := store.GetDeliverySecret(ctx, ref)
				return err
			},
		},
		{
			name:    "delete",
			backend: &memoryDeliverySecretRedis{delErr: providerErr},
			act: func(store *RedisDeliverySecretStore) error {
				return store.DeleteDeliverySecret(ctx, ref)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, err := NewRedisDeliverySecretStoreWithClient(newMemoryDeliverySecretRedisClient(tc.backend), "challenge-pepper")
			if err != nil {
				t.Fatalf("NewRedisDeliverySecretStoreWithClient: %v", err)
			}
			err = tc.act(store)
			if err == nil {
				t.Fatal("expected safe error")
			}
			for _, forbidden := range []string{"password", "private-host", ref, code} {
				if strings.Contains(err.Error(), forbidden) {
					t.Fatalf("error leaked %q: %v", forbidden, err)
				}
			}
		})
	}
}

func TestRedisDeliverySecretStoreRejectsInvalidAndCorruptMaterial(t *testing.T) {
	ctx := context.Background()
	backend := &memoryDeliverySecretRedis{}
	client := newMemoryDeliverySecretRedisClient(backend)
	if _, err := NewRedisDeliverySecretStoreWithClient(client, ""); err == nil {
		t.Fatal("empty key material accepted")
	}
	store, err := NewRedisDeliverySecretStoreWithClient(client, "challenge-pepper")
	if err != nil {
		t.Fatalf("NewRedisDeliverySecretStoreWithClient: %v", err)
	}
	if err := store.PutDeliverySecret(ctx, "", "123456", ChallengeTTL); err == nil {
		t.Fatal("empty ref accepted")
	}
	if err := store.PutDeliverySecret(ctx, "auth_challenge:one", "123456", 0); err == nil {
		t.Fatal("non-positive TTL accepted")
	}

	key := store.redisKey("auth_challenge:corrupt")
	backend.mu.Lock()
	backend.values[key] = "v1:not-valid-ciphertext"
	backend.mu.Unlock()
	if _, _, err := store.GetDeliverySecret(ctx, "auth_challenge:corrupt"); err == nil {
		t.Fatal("corrupt ciphertext accepted")
	} else if strings.Contains(err.Error(), "not-valid-ciphertext") {
		t.Fatalf("corrupt ciphertext leaked: %v", err)
	}
}
