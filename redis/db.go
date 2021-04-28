package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/thss-cercis/cercis-server/config"
	"time"
)

var dbNow *redis.Client

// GetRedis 获得 redis 的 client 单例
func GetRedis() (client *redis.Client, err error) {
	if dbNow == nil {
		cr := config.GetConfig().Redis
		dbNow = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%v:%v", cr.Host, cr.Port),
			Username: cr.Username,
			Password: cr.Password,
			DB:       cr.Database,
		})
		_, err := dbNow.Ping(context.Background()).Result()
		if err != nil {
			return nil, err
		}
	}
	return dbNow, nil
}

// PutKV 存放一个键值对，含有有效期
func PutKV(tag string, key string, value string, exp time.Duration) error {
	client, err := GetRedis()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.Set(ctx, fmt.Sprintf("%v_%v", tag, key), value, exp).Err()
}

// GetKV 获得一个键值对中的值.
//
// Throws: redis.Nil 表示找不到此 key.
func GetKV(tag string, key string) (string, error) {
	client, err := GetRedis()
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	return client.Get(ctx, fmt.Sprintf("%v_%v", tag, key)).Result()
}

// GetKVExp 获得一个键值对的剩余过期时间.
//
// Throws: redis.Nil 表示找不到此 key.
func GetKVExp(tag string, key string) (time.Duration, error) {
	client, err := GetRedis()
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	return client.TTL(ctx, fmt.Sprintf("%v_%v", tag, key)).Result()
}
