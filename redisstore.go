package redisstore

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore represents the session store.
type RedisStore struct {
	client *redis.Client
	prefix string
}

// New returns a new RedisStore instance. The pool parameter should be a pointer
// to a redigo connection pool. See https://godoc.org/github.com/gomodule/redigo/redis#Pool.
func New(client *redis.Client) *RedisStore {
	return NewWithPrefix(client, "scs:session:")
}

// NewWithPrefix returns a new RedisStore instance. The pool parameter should be a pointer
// to a redigo connection pool. The prefix parameter controls the Redis key
// prefix, which can be used to avoid naming clashes if necessary.
func NewWithPrefix(client *redis.Client, prefix string) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

// Find returns the data for a given session token from the RedisStore instance.
// If the session token is not found or is expired, the returned exists flag
// will be set to false.
func (r *RedisStore) Find(token string) (b []byte, exists bool, err error) {
	ctx := context.Background()
	b, err = r.client.Get(ctx, r.prefix+token).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Commit adds a session token and data to the RedisStore instance with the
// given expiry time. If the session token already exists then the data and
// expiry time are updated.
func (r *RedisStore) Commit(token string, b []byte, expiry time.Time) error {
	ctx := context.Background()
	expiration := time.Until(expiry)
	return r.client.Set(ctx, r.prefix+token, b, expiration).Err()
}

// Delete removes a session token and corresponding data from the RedisStore
// instance.
func (r *RedisStore) Delete(token string) error {
	ctx := context.Background()
	return r.client.Del(ctx, r.prefix+token).Err()
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the RedisStore instance.
func (r *RedisStore) All(ctx context.Context) (map[string][]byte, error) {
	keys, err := r.client.Keys(ctx, r.prefix+"*").Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	sessions := make(map[string][]byte)
	for _, key := range keys {
		token := key[len(r.prefix):]
		data, exists, err := r.Find(token)
		if err != nil {
			return nil, err
		}
		if exists {
			sessions[token] = data
		}
	}
	return sessions, nil
}

func makeMillisecondTimestamp(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
