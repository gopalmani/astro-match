
package cache

import (
    "github.com/go-redis/redis/v8"
    "context"
)

var ctx = context.Background()
var rdb *redis.Client

func InitRedis() {
    rdb = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
}

func SetSession(key, value string) error {
    return rdb.Set(ctx, key, value, 0).Err()
}

func GetSession(key string) (string, error) {
    return rdb.Get(ctx, key).Result()
}
