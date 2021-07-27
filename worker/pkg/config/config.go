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
	Test     bool
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

func TestConfig() *Config {
	return &Config{
		Redis: RedisConfig{
			Address: "localhost:6379",
		},
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
		Test: true,
	}
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

func Merge(defaults *Config, req *livekit.RecordingReservation) (string, error) {
	merged := &Config{
		Input: &livekit.RecordingInput{
			Url:       req.Input.Url,
			Template:  req.Input.Template,
			Width:     defaults.Input.Width,
			Height:    defaults.Input.Height,
			Depth:     defaults.Input.Depth,
			Framerate: defaults.Input.Framerate,
		},
		Output: &livekit.RecordingOutput{
			File:           req.Output.File,
			Rtmp:           req.Output.Rtmp,
			S3:             req.Output.S3,
			Width:          defaults.Output.Width,
			Height:         defaults.Output.Height,
			AudioBitrate:   defaults.Output.AudioBitrate,
			AudioFrequency: defaults.Output.AudioFrequency,
			VideoBitrate:   defaults.Output.VideoBitrate,
			VideoBuffer:    defaults.Output.VideoBuffer,
		},
	}

	// input overrides
	if req.Input.Width != 0 && req.Input.Height != 0 {
		merged.Input.Width = req.Input.Width
		merged.Input.Height = req.Input.Height
	}
	if req.Input.Depth != 0 {
		merged.Input.Depth = req.Input.Depth
	}
	if req.Input.Framerate != 0 {
		merged.Input.Framerate = req.Input.Framerate
	}

	// output overrides
	if req.Output.Width != 0 && req.Output.Height != 0 {
		merged.Output.Width = req.Output.Width
		merged.Output.Height = req.Output.Height
	}
	if req.Output.AudioBitrate != "" {
		merged.Output.AudioBitrate = req.Output.AudioBitrate
	}
	if req.Output.AudioFrequency != "" {
		merged.Output.AudioFrequency = req.Output.AudioFrequency
	}
	if req.Output.VideoBitrate != "" {
		merged.Output.VideoBitrate = req.Output.VideoBitrate
	}
	if req.Output.VideoBuffer != "" {
		merged.Output.VideoBuffer = req.Output.VideoBuffer
	}

	b, err := json.Marshal(merged)
	return string(b), err
}
