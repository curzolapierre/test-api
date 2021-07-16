package models

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/curzolapierre/hook-manager/redis"
	goRedis "github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Codexcuse Struct
// source is a field to identify where the entry was saved
// In case of discord client, the source will be the GuildID
type Codexcuse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   *User  `json:"author"`
	Reporter *User  `json:"reporter"`
	Content  string `json:"content"`
}

// User Struct
type User struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
}

// Meta is the struct that store pagination Meta datas
type Meta struct {
	CurrentPage int  `json:"current_page"`
	PrevPage    *int `json:"prev_page"`
	NextPage    *int `json:"next_page"`
	TotalPages  int  `json:"total_pages"`
	TotalCount  int  `json:"total_count"`
}

type RedisStoreCodexcuses struct {
	*goRedis.Client
}

var (
	PageSize = 10
)

func (c *RedisStoreCodexcuses) GetRandom(ctx context.Context, source string) (*Codexcuse, error) {
	log := logger.Get(ctx)

	log.WithField("function", "GetRandom").WithField("key", c.key(source))
	log.Debugln("source:", source)

	if c == nil {
		return nil, errors.New("fail to get redis client")
	}

	// Get all keys to know how many items there are in the table
	rangeRes := c.ZRevRange(c.excuseIDKey(source), 0, -1)
	if rangeRes.Err() != nil {
		return nil, errors.Wrap(rangeRes.Err(), "fail to get range of all IDs")
	}

	rangeResLen := 1
	if len(rangeRes.Val()) > 0 {
		rangeResLen = len(rangeRes.Val())
	}
	excusesPos := rand.Intn(rangeResLen)
	rangeRes = c.ZRevRange(c.excuseIDKey(source), int64(excusesPos), int64(excusesPos))
	if rangeRes.Err() != nil {
		return nil, errors.Wrap(rangeRes.Err(), "fail to get range of IDs")
	}

	if len(rangeRes.Val()) == 0 {
		return nil, nil
	}

	res := c.HGet(c.key(source), rangeRes.Val()[0])
	if res.Err() != nil {
		return nil, errors.Wrap(res.Err(), "fail to get all excuses")
	}
	var excuse Codexcuse
	json.Unmarshal([]byte(res.Val()), &excuse)
	return &excuse, nil
}

func (c *RedisStoreCodexcuses) GetByUser(ctx context.Context, source string, userID string) (*[]Codexcuse, error) {
	log := logger.Get(ctx)

	log.WithField("function", "GetByUser").WithField("key", c.key(source))
	log.Debugln("source:", source)

	if c == nil {
		return nil, errors.New("fail to get redis client")
	}

	var excuses []Codexcuse

	res := c.HGetAll(c.key(source))
	if res.Err() != nil {
		return &excuses, errors.Wrap(res.Err(), "fail to get all excuses")
	}
	for _, c := range res.Val() {
		var excuse Codexcuse
		json.Unmarshal([]byte(c), &excuse)
		if excuse.Author.ID == userID {
			excuses = append(excuses, excuse)
		}
	}
	return &excuses, nil
}

func (c *RedisStoreCodexcuses) GetAll(ctx context.Context, source string, requestedPage int, excuses *[]Codexcuse) (Meta, error) {
	log := logger.Get(ctx)

	log.WithField("function", "GetAll").WithField("key", c.key(source))
	log.Debugln("source:", source)
	meta := Meta{}

	if c == nil {
		return meta, errors.New("fail to get redis client")
	}

	// Get all keys to know how many items there are in the table
	rangeRes := c.ZRevRange(c.excuseIDKey(source), 0, -1)
	if rangeRes.Err() != nil {
		return meta, errors.Wrap(rangeRes.Err(), "fail to get range of all IDs")
	}
	meta.CurrentPage = requestedPage
	meta.TotalCount = len(rangeRes.Val())
	meta.TotalPages = meta.TotalCount / PageSize
	// We truncate to the higher integer except in the case of a "round" division
	if meta.TotalCount%PageSize != 0 {
		meta.TotalPages++
	}
	// NextPage must be null when unavailable
	if meta.CurrentPage < meta.TotalPages {
		meta.NextPage = new(int)
		*meta.NextPage = meta.CurrentPage + 1
	}
	// PrevPage must be null when unavailable
	if meta.CurrentPage > 1 {
		meta.PrevPage = new(int)
		*meta.PrevPage = meta.CurrentPage - 1
	}

	skipOffset := (requestedPage - 1) * PageSize
	rangeRes = c.ZRevRange(c.excuseIDKey(source), int64(skipOffset), int64(skipOffset+PageSize-1))
	if rangeRes.Err() != nil {
		return meta, errors.Wrap(rangeRes.Err(), "fail to get range of IDs")
	}

	if len(rangeRes.Val()) == 0 {
		return meta, nil
	}

	res := c.HMGet(c.key(source), rangeRes.Val()...)
	if res.Err() != nil {
		return meta, errors.Wrap(res.Err(), "fail to get all excuses")
	}
	*excuses = make([]Codexcuse, len(res.Val()))
	count := 0
	for _, c := range res.Val() {
		var excuse Codexcuse
		json.Unmarshal([]byte(c.(string)), &excuse)
		(*excuses)[count] = excuse
		count++
	}

	return meta, nil
}

func (c *RedisStoreCodexcuses) Get(ctx context.Context, source, id string) (*Codexcuse, error) {
	log := logger.Get(ctx)

	log.WithField("function", "Get").WithField("key", c.key(source))
	log.Debugln("source:", source)
	if c == nil {
		return nil, errors.New("fail to get redis client")
	}
	res := c.HGet(c.key(source), id)
	if res.Val() == "" {
		return nil, nil
	}
	if res.Err() != nil {
		return nil, errors.Wrap(res.Err(), "fail to get excuses: "+id)
	}

	var excuse Codexcuse
	err := json.Unmarshal([]byte(res.Val()), &excuse)
	if err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal")
	}
	return &excuse, nil
}

func (c *RedisStoreCodexcuses) Add(ctx context.Context, source string, excuse Codexcuse) error {
	log := logger.Get(ctx)

	log.WithField("function", "Add").WithField("key", c.key(source))
	log.Debugln("source:", source)
	if c == nil {
		return errors.New("fail to get redis client")
	}

	excuse.ID = uuid.New().String()
	bytes, err := json.Marshal(excuse)
	if err != nil {
		return errors.Wrap(err, "fail to marshal excuse")
	}

	// Use a CodexcuseIDs key to store a sorted list of codexcuse's ID, sorted by
	// creation timestamp
	t := time.Now()
	timestamp := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)
	zAddRes := c.ZAdd(c.excuseIDKey(source), goRedis.Z{
		Score:  float64(timestamp),
		Member: excuse.ID,
	})
	if zAddRes.Err() != nil || zAddRes.Val() != 1 {
		return errors.Wrap(zAddRes.Err(), "fail to ZADD the ID of excuse")
	}

	// Use Codexcuse key to store excuse content store by ID
	res := c.HSet(c.key(source), excuse.ID, bytes)
	if res.Err() != nil {
		return errors.Wrap(res.Err(), "fail to set an excuses")
	}

	log.Debugln("addedd excuse:", excuse.ID)
	return nil
}

// Delete remove from Codexcuse, CodescuseIDs and CodexcuseAuthors the field
// corresponding with id parameter
func (c *RedisStoreCodexcuses) Delete(ctx context.Context, source, id string) error {
	log := logger.Get(ctx)

	log.WithField("function", "Delete").WithField("key", c.key(source))
	log.Debugln("source:", source)
	if c == nil {
		return errors.New("fail to get redis client")
	}

	res := c.HDel(c.key(source), id)
	if res.Err() != nil {
		return errors.Wrap(res.Err(), "fail to delete hash of excuse: "+id)
	}

	res = c.ZRem(c.excuseIDKey(source), id)
	if res.Err() != nil {
		return errors.Wrap(res.Err(), "fail to delete ID of excuse: "+id)
	}

	return nil
}

func (c *RedisStoreCodexcuses) key(source string) string {
	return fmt.Sprintf("%sCodexcuse:source:%s", redis.Prefix(), source)
}

func (c *RedisStoreCodexcuses) excuseIDKey(source string) string {
	return fmt.Sprintf("%sCodexcuseIDs:source:%s", redis.Prefix(), source)
}
