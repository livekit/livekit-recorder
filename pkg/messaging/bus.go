package messaging

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/livekit/protocol/logger"
	"github.com/livekit/protocol/utils"
	"github.com/pkg/errors"
	"crypto/tls"

	"github.com/livekit/livekit-recorder/pkg/config"
)

func NewMessageBus(conf *config.Config) (utils.MessageBus, error) {
	logger.Infow("connecting to redis work queue", "addr", conf.Redis.Address)
	rcOptions :=  &redis.Options{
		Addr:     conf.Redis.Address,
		Username: conf.Redis.Username,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	}
	if conf.Redis.UseTLS {
		rcOptions = &redis.Options{
			Addr:     conf.Redis.Address,
			Username: conf.Redis.Username,
			Password: conf.Redis.Password,
			DB:       conf.Redis.DB,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
	}
	rc := redis.NewClient(rcOptions)
	err := rc.Ping(context.Background()).Err()
	if err != nil {
		err = errors.Wrap(err, "unable to connect to redis")
	}
	return utils.NewRedisMessageBus(rc), err
}
