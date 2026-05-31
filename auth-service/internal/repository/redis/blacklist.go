package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklist stores revoked JWT JTIs in Redis until they naturally expire.
// Key: blacklist:{jti}  Value: "1"  TTL: remaining JWT lifetime
type TokenBlacklist struct {
	rdb *redis.Client
}

func NewTokenBlacklist(rdb *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{rdb: rdb}
}

func (b *TokenBlacklist) Add(ctx context.Context, jti string, ttl time.Duration) error {
	return b.rdb.Set(ctx, "blacklist:"+jti, "1", ttl).Err()
}

func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	val, err := b.rdb.Get(ctx, "blacklist:"+jti).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}