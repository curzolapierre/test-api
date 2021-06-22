package redis

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/curzolapierre/hook-manager/environment"
	"github.com/go-redis/redis"
	"gopkg.in/errgo.v1"
)

var (
	prefix      = fmt.Sprintf("hook_manager")
	initRedis   = &sync.Once{}
	errInit     error
	redisClient *redis.Client
)

func Prefix() string {
	return fmt.Sprintf("%s:%s:", prefix, environment.ENV["GO_ENV"])
}

func Client() (*redis.Client, error) {
	initRedis.Do(func() {
		var pass string
		if environment.ENV["REDIS_URL"] == "" {
			errInit = errgo.New("No redis credentials (ENV[REDIS_URL])")
			return
		}

		redisURL, err := url.Parse(environment.ENV["REDIS_URL"])
		if err != nil {
			errInit = errgo.Mask(err)
			return
		}

		if redisURL.User != nil {
			pass, _ = redisURL.User.Password()
		}

		redisClient = redis.NewClient(&redis.Options{
			Password:    pass,
			Addr:        redisURL.Host,
			PoolSize:    environment.RedisPoolSize,
			MaxRetries:  3,
			IdleTimeout: 80 * time.Second,
		})
	})

	if errInit != nil {
		return nil, errInit
	}
	return redisClient, nil
}
