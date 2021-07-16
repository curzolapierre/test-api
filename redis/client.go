package redis

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/curzolapierre/hook-manager/config"
	"github.com/go-redis/redis"
)

var (
	prefix      = fmt.Sprintf("hook_manager")
	initRedis   = &sync.Once{}
	errInit     error
	redisClient *redis.Client
)

func Prefix() string {
	return fmt.Sprintf("%s:%s:", prefix, os.Getenv("GO_ENV"))
}

func Client(config config.Config) (*redis.Client, error) {
	initRedis.Do(func() {
		var pass string
		if config.RedisURL == "" {
			errInit = errors.New("No redis credentials (ENV[REDIS_URL])")
			return
		}

		redisURL, err := url.Parse(config.RedisURL)
		if err != nil {
			errInit = err
			return
		}

		if redisURL.User != nil {
			pass, _ = redisURL.User.Password()
		}

		redisClient = redis.NewClient(&redis.Options{
			Password:    pass,
			Addr:        redisURL.Host,
			PoolSize:    config.RedisPoolSize,
			MaxRetries:  3,
			IdleTimeout: 80 * time.Second,
		})
	})

	if errInit != nil {
		return nil, errInit
	}
	return redisClient, nil
}
