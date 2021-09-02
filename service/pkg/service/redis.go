package service

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/livekit/protocol/utils"
	"github.com/pkg/errors"

	"github.com/livekit/livekit-recorder/service/pkg/config"
	"github.com/livekit/livekit-recorder/service/pkg/logger"
)

func NewMessageBus(conf *config.Config) (utils.MessageBus, error) {
	logger.Infow("connecting to redis work queue", "addr", conf.Redis.Address)
	rc := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Username: conf.Redis.Username,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})
	err := rc.Ping(context.Background()).Err()
	if err != nil {
		err = errors.Wrap(err, "unable to connect to redis")
	}
	return utils.NewRedisMessageBus(rc), err
}
