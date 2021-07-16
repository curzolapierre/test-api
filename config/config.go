package config

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type Config struct {
	GoEnv            string `envconfig:"GO_ENV" default:"development"`
	Public_hostname  string `envconfig:"PUBLIC_HOSTNAME"`
	Private_hostname string `envconfig:"PRIVATE_HOSTNAME"`
	HttpHost         string `envconfig:"HOST" default:"localhost"`
	HttpPort         string `envconfig:"PORT" default:"8080"`
	BasicAuthApiUser string `envconfig:"BASIC_AUTH_API_USER"`
	BasicAuthApiPass string `envconfig:"BASIC_AUTH_API_PASS"`
	Redis_dump_stats string `envconfig:"REDIS_DUMP_STATS" default:"false"`
	RedisURL         string `envconfig:"REDIS_URL" default:"redis://localhost:6379"`
	RedisPoolSize    int    `envconfig:"REDIS_POOL_SIZE" default:"10"`
	RedisScanSize    int64  `envconfig:"REDIS_SCAN_SIZE" default:"10"`
	ContextTimeout   int    `envconfig:"CONTEXT_TIMEOUT" default:"20"`

	// Worker concurrency
	RedisEntriesPublishConcurrency int `envconfig:"REDIS_ENTRIES_PUBLISH_CONCURRENCY" default:"10"`
	RedisEntriesCacheConcurrency   int `envconfig:"REDIS_ENTRIES_CACHE_CONCURRENCY" default:"10"`
}

func Lookup() (Config, error) {
	// init rand
	rand.Seed(time.Now().UTC().UnixNano())

	env := Config{}
	err := envconfig.Process("", &env)
	if err != nil {
		return env, errors.Wrap(err, "fail to parse the application environment")
	}

	if env.GoEnv == "production" {
		fmt.Println("Run in production ! 👌🔥")
	}

	return env, nil
}
