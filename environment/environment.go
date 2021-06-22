package environment

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	ENV = map[string]string{
		"GO_ENV":              "development",
		"PUBLIC_HOSTNAME":     "",
		"PRIVATE_HOSTNAME":    "",
		"HTTP_HOST":           "localhost",
		"HTTP_PORT":           "8080",
		"BASIC_AUTH_API_USER": "",
		"BASIC_AUTH_API_PASS": "",
		"REDIS_DUMP_STATS":    "false",
		"REDIS_URL":           "redis://localhost:6379",
		"REDIS_POOL_SIZE":     "10",
		"REDIS_SCAN_SIZE":     "10",
		"CONTEXT_TIMEOUT":     "20",

		// Worker concurrency
		"REDIS_ENTRIES_PUBLISH_CONCURRENCY": "10",
		"REDIS_ENTRIES_CACHE_CONCURRENCY":   "10",
	}

	RedisPoolSize                  int
	RedisScanSize                  int64
	RedisEntriesPublishConcurrency int
	RedisEntriesCacheConcurrency   int
	ContextTimeout                 time.Duration
)

func Lookup() {
	// init rand
	rand.Seed(time.Now().UTC().UnixNano())

	godotenv.Load()

	for k := range ENV {
		if os.Getenv(k) != "" {
			ENV[k] = os.Getenv(k)
		}
	}
	if os.Getenv("REDIS_PORT_6379_TCP_ADDR") != "" {
		ENV["REDIS_URL"] = "redis://" + os.Getenv("REDIS_PORT_6379_TCP_ADDR") + ":" + os.Getenv("REDIS_PORT_6379_TCP_PORT")
	}

	if strings.Contains(strings.Join(os.Args, ""), ".test") {
		ENV["GO_ENV"] = "test"
	}

	if ENV["GO_ENV"] == "production" {
		fmt.Println("Run in production ! 👌🔥")
	}

	s, err := strconv.Atoi(ENV["REDIS_POOL_SIZE"])
	if err != nil {
		panic(err)
	}
	RedisPoolSize = s

	s, err = strconv.Atoi(ENV["REDIS_SCAN_SIZE"])
	if err != nil {
		panic(err)
	}
	RedisScanSize = int64(s)

	c, err := strconv.Atoi(ENV["REDIS_ENTRIES_PUBLISH_CONCURRENCY"])
	if err != nil {
		panic(err)
	}
	RedisEntriesPublishConcurrency = c

	c, err = strconv.Atoi(ENV["REDIS_ENTRIES_CACHE_CONCURRENCY"])
	if err != nil {
		panic(err)
	}
	RedisEntriesCacheConcurrency = c

	if os.Getenv("PORT") != "" {
		ENV["HTTP_PORT"] = os.Getenv("PORT")
	}

	c, err = strconv.Atoi(ENV["CONTEXT_TIMEOUT"])
	if err != nil {
		panic(err)
	}
	ContextTimeout = time.Duration(c) * time.Second
}
