//go:build integration

package auth

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRedisDeliverySecretStoreCrossClientIntegration(t *testing.T) {
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if redisURL == "" {
		t.Skip("REDIS_URL is required for Redis integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	producer, err := NewRedisDeliverySecretStore(redisURL, "integration-pepper")
	if err != nil {
		t.Fatalf("producer: %v", err)
	}
	defer producer.Close()
	consumer, err := NewRedisDeliverySecretStore(redisURL, "integration-pepper")
	if err != nil {
		t.Fatalf("consumer: %v", err)
	}
	defer consumer.Close()
	if err := producer.Ping(ctx); err != nil {
		t.Fatalf("producer ping: %v", err)
	}
	if err := consumer.Ping(ctx); err != nil {
		t.Fatalf("consumer ping: %v", err)
	}

	ref := "auth_challenge:integration-cross-client-" + NewID()
	t.Cleanup(func() { _ = producer.DeleteDeliverySecret(context.Background(), ref) })
	if err := producer.PutDeliverySecret(ctx, ref, "654321", ChallengeTTL); err != nil {
		t.Fatalf("PutDeliverySecret: %v", err)
	}
	key := producer.redisKey(ref)
	raw, err := producer.redis.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("raw Redis get: %v", err)
	}
	if strings.Contains(key, ref) || strings.Contains(raw, ref) || strings.Contains(raw, "654321") {
		t.Fatalf("Redis key/value leaked ref or raw code")
	}
	ttl, err := producer.redis.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("Redis TTL: %v", err)
	}
	if ttl <= 4*time.Minute+50*time.Second || ttl > ChallengeTTL {
		t.Fatalf("Redis TTL = %v, want approximately %v", ttl, ChallengeTTL)
	}
	code, ok, err := consumer.GetDeliverySecret(ctx, ref)
	if err != nil || !ok || code != "654321" {
		t.Fatalf("cross-client Get = %q/%v/%v", code, ok, err)
	}
	if err := consumer.DeleteDeliverySecret(ctx, ref); err != nil {
		t.Fatalf("cross-client Delete: %v", err)
	}
	if _, ok, err := producer.GetDeliverySecret(ctx, ref); err != nil || ok {
		t.Fatalf("producer observed deleted value: ok=%v err=%v", ok, err)
	}

	expiringRef := ref + ":expiry"
	if err := producer.PutDeliverySecret(ctx, expiringRef, "111222", 100*time.Millisecond); err != nil {
		t.Fatalf("Put expiring secret: %v", err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		_, ok, err := consumer.GetDeliverySecret(ctx, expiringRef)
		if err != nil {
			t.Fatalf("Get expiring secret: %v", err)
		}
		if !ok {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Redis did not expire delivery secret")
		}
		time.Sleep(25 * time.Millisecond)
	}
}
