package controllers

import (
	"github.com/curzolapierre/hook-manager/models"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type RequestContext struct {
	Log        logrus.FieldLogger
	Codexcuse  models.Codexcuse
	RedisStore *models.RedisStoreCodexcuses
}

func (r *RequestContext) InitStore(redisClient *redis.Client) {
}
