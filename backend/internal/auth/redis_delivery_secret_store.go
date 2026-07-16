package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/hkdf"
)

const (
	deliverySecretRedisKeyPrefix = "easyinterview:auth:delivery-secret:v1:"
	deliverySecretCipherVersion  = "v1:"
	deliverySecretHKDFContext    = "easyinterview/auth/delivery-secret/v1"
)

type DeliverySecretRedisClient interface {
	Set(context.Context, string, string, time.Duration) error
	Get(context.Context, string) (string, bool, error)
	Del(context.Context, string) error
}

type goRedisDeliverySecretClient struct {
	client *redis.Client
}

func (c *goRedisDeliverySecretClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *goRedisDeliverySecretClient) Get(ctx context.Context, key string) (string, bool, error) {
	value, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (c *goRedisDeliverySecretClient) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

type RedisDeliverySecretStore struct {
	client DeliverySecretRedisClient
	aead   cipher.AEAD
	redis  *redis.Client
}

func NewRedisDeliverySecretStore(redisURL string, keyMaterial string) (*RedisDeliverySecretStore, error) {
	options, err := redis.ParseURL(strings.TrimSpace(redisURL))
	if err != nil {
		return nil, fmt.Errorf("delivery secret Redis configuration invalid")
	}
	client := redis.NewClient(options)
	store, err := NewRedisDeliverySecretStoreWithClient(&goRedisDeliverySecretClient{client: client}, keyMaterial)
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	store.redis = client
	return store, nil
}

func NewRedisDeliverySecretStoreWithClient(client DeliverySecretRedisClient, keyMaterial string) (*RedisDeliverySecretStore, error) {
	if client == nil {
		return nil, fmt.Errorf("delivery secret Redis client unavailable")
	}
	if strings.TrimSpace(keyMaterial) == "" {
		return nil, fmt.Errorf("delivery secret encryption key unavailable")
	}
	key := make([]byte, 32)
	reader := hkdf.New(sha256.New, []byte(keyMaterial), nil, []byte(deliverySecretHKDFContext))
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, fmt.Errorf("delivery secret encryption key derivation failed")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("delivery secret encryption initialization failed")
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("delivery secret encryption initialization failed")
	}
	return &RedisDeliverySecretStore{client: client, aead: aead}, nil
}

func (s *RedisDeliverySecretStore) PutDeliverySecret(ctx context.Context, ref string, token string, ttl time.Duration) error {
	if s == nil || s.client == nil || s.aead == nil {
		return fmt.Errorf("delivery secret store unavailable")
	}
	if strings.TrimSpace(ref) == "" || strings.TrimSpace(token) == "" || ttl <= 0 {
		return fmt.Errorf("delivery secret input invalid")
	}
	key := s.redisKey(ref)
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("delivery secret encryption failed")
	}
	sealed := s.aead.Seal(nil, nonce, []byte(token), []byte(key))
	payload := append(nonce, sealed...)
	value := deliverySecretCipherVersion + base64.RawURLEncoding.EncodeToString(payload)
	if err := s.client.Set(ctx, key, value, ttl); err != nil {
		return fmt.Errorf("delivery secret storage failed")
	}
	return nil
}

func (s *RedisDeliverySecretStore) GetDeliverySecret(ctx context.Context, ref string) (string, bool, error) {
	if s == nil || s.client == nil || s.aead == nil {
		return "", false, fmt.Errorf("delivery secret store unavailable")
	}
	if strings.TrimSpace(ref) == "" {
		return "", false, fmt.Errorf("delivery secret input invalid")
	}
	key := s.redisKey(ref)
	value, ok, err := s.client.Get(ctx, key)
	if err != nil {
		return "", false, fmt.Errorf("delivery secret lookup failed")
	}
	if !ok {
		return "", false, nil
	}
	if !strings.HasPrefix(value, deliverySecretCipherVersion) {
		return "", false, fmt.Errorf("delivery secret decryption failed")
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, deliverySecretCipherVersion))
	if err != nil || len(payload) < s.aead.NonceSize() {
		return "", false, fmt.Errorf("delivery secret decryption failed")
	}
	nonce, ciphertext := payload[:s.aead.NonceSize()], payload[s.aead.NonceSize():]
	plain, err := s.aead.Open(nil, nonce, ciphertext, []byte(key))
	if err != nil {
		return "", false, fmt.Errorf("delivery secret decryption failed")
	}
	return string(plain), true, nil
}

func (s *RedisDeliverySecretStore) DeleteDeliverySecret(ctx context.Context, ref string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("delivery secret store unavailable")
	}
	if strings.TrimSpace(ref) == "" {
		return fmt.Errorf("delivery secret input invalid")
	}
	if err := s.client.Del(ctx, s.redisKey(ref)); err != nil {
		return fmt.Errorf("delivery secret cleanup failed")
	}
	return nil
}

func (s *RedisDeliverySecretStore) Ping(ctx context.Context) error {
	if s == nil || s.redis == nil {
		return fmt.Errorf("delivery secret Redis client unavailable")
	}
	if err := s.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("delivery secret Redis unavailable")
	}
	return nil
}

func (s *RedisDeliverySecretStore) Close() error {
	if s == nil || s.redis == nil {
		return nil
	}
	return s.redis.Close()
}

func (s *RedisDeliverySecretStore) redisKey(ref string) string {
	digest := sha256.Sum256([]byte(ref))
	return deliverySecretRedisKeyPrefix + hex.EncodeToString(digest[:])
}

var _ DeliverySecretStore = (*RedisDeliverySecretStore)(nil)
