package config

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"

	livekit "github.com/livekit/livekit-recording/worker/proto"
)

type Config struct {
	Redis    RedisConfig              `json:"redis"`
	Input    *livekit.RecordingInput  `json:"input"`
	Output   *livekit.RecordingOutput `json:"output"`
	LogLevel string                   `json:"log_level"`
}

type RedisConfig struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func NewConfig(confString string, c *cli.Context) (*Config, error) {
	// start with defaults
	conf := &Config{
		Redis: RedisConfig{},
		Input: &livekit.RecordingInput{
			Width:     1920,
			Height:    1080,
			Depth:     24,
			Framerate: 25,
		},
		Output: &livekit.RecordingOutput{
			AudioBitrate:   "128k",
			AudioFrequency: "44100",
			VideoBitrate:   "2976k",
			VideoBuffer:    "5952k",
		},
	}

	if confString != "" {
		if err := json.Unmarshal([]byte(confString), conf); err != nil {
			return nil, fmt.Errorf("could not parse config: %v", err)
		}
	}

	if c != nil {
		if err := conf.updateFromCLI(c); err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func (conf *Config) updateFromCLI(c *cli.Context) error {
	if c.IsSet("redis-host") {
		conf.Redis.Address = c.String("redis-host")
	}
	if c.IsSet("redis-password") {
		conf.Redis.Password = c.String("redis-password")
	}

	return nil
}
