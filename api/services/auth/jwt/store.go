package jwt

import (
	"agones-minecraft/config"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Global JWTStore
var TokenStore JWTStore

// Client interface for storing, checking, and deleting tokenId from a store
type JWTStore interface {
	Add(tokenId string, userId string, exp time.Time) error
	Exists(tokenId string) error
	Delete(tokenId string) error
}

// Redis JWTStore implentation.
type RedisStore struct {
	redis *redis.Client
}

// Creates and returns a new JWT store client
func New() JWTStore {
	addr, pass := config.GetRedisCreds()
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
	})
	TokenStore = &RedisStore{client}
	return TokenStore
}

// Returns the existing JWT store client
// returns nil if the client has not been initialized
func Get() JWTStore {
	return TokenStore
}

// Check if the tokenId exists in the store
func (r *RedisStore) Add(tokenId, userId string, exp time.Time) error {
	return r.redis.Set(context.Background(), tokenId, userId, time.Until(exp)).Err()
}

// Check if the tokenId exists in the store
func (r *RedisStore) Exists(tokenId string) error {
	return r.redis.Get(context.Background(), tokenId).Err()
}

// Delete the tokenId from the store
func (r *RedisStore) Delete(tokenId string) error {
	return r.redis.Del(context.Background(), tokenId).Err()
}
