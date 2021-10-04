package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	livekit "github.com/livekit/protocol/proto"
)

const Display = ":99"

type Config struct {
	ApiKey      string                    `yaml:"api_key"`
	ApiSecret   string                    `yaml:"api_secret"`
	WsUrl       string                    `yaml:"ws_url"`
	HealthPort  int                       `yaml:"health_port"`
	LogLevel    string                    `yaml:"log_level"`
	GstLogLevel string                    `yaml:"gst_log_level"`
	Defaults    *livekit.RecordingOptions `yaml:"defaults"`
	Redis       RedisConfig               `yaml:"redis"`
	S3          S3Config                  `yaml:"s3"`
	Test        bool                      `yaml:"-"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type S3Config struct {
	AccessKey string `yaml:"access_key"`
	Secret    string `yaml:"secret"`
	Region    string `yaml:"region"`
}

func NewConfig(confString string) (*Config, error) {
	// start with defaults
	conf := &Config{
		LogLevel:    "debug",
		GstLogLevel: "3",
		Defaults: &livekit.RecordingOptions{
			InputWidth:     1920,
			InputHeight:    1080,
			Depth:          24,
			Framerate:      30,
			AudioBitrate:   128,
			AudioFrequency: 44100,
			VideoBitrate:   4500,
		},
	}

	if confString != "" {
		if err := yaml.Unmarshal([]byte(confString), conf); err != nil {
			return nil, fmt.Errorf("could not parse config: %v", err)
		}
	}

	// apply preset options
	if conf.Defaults.Preset != livekit.RecordingPreset_NONE {
		conf.Defaults = fromPreset(conf.Defaults.Preset)
	}

	if err := os.Setenv("DISPLAY", Display); err != nil {
		return nil, err
	}
	if err := os.Setenv("GST_DEBUG", conf.GstLogLevel); err != nil {
		return nil, err
	}

	return conf, nil
}

func TestConfig() *Config {
	return &Config{
		Redis: RedisConfig{
			Address: "localhost:6379",
		},
		Defaults: &livekit.RecordingOptions{
			InputWidth:     1920,
			InputHeight:    1080,
			Depth:          24,
			Framerate:      30,
			AudioBitrate:   128,
			AudioFrequency: 44100,
			VideoBitrate:   4500,
		},
		Test: true,
	}
}

func UpdateRequestParams(conf *Config, req *livekit.StartRecordingRequest) {
	if req.Options.Preset != livekit.RecordingPreset_NONE {
		req.Options = fromPreset(req.Options.Preset)
		return
	}

	if req.Options.InputWidth == 0 || req.Options.InputHeight == 0 {
		req.Options.InputWidth = conf.Defaults.InputHeight
		req.Options.InputHeight = conf.Defaults.InputWidth
	}
	if req.Options.Depth == 0 {
		req.Options.Depth = conf.Defaults.Depth
	}
	if req.Options.Framerate == 0 {
		req.Options.Framerate = conf.Defaults.Framerate
	}
	if req.Options.AudioBitrate == 0 {
		req.Options.AudioBitrate = conf.Defaults.AudioBitrate
	}
	if req.Options.AudioFrequency == 0 {
		req.Options.AudioFrequency = conf.Defaults.AudioFrequency
	}
	if req.Options.VideoBitrate == 0 {
		req.Options.VideoBitrate = conf.Defaults.VideoBitrate
	}

	return
}

func fromPreset(preset livekit.RecordingPreset) *livekit.RecordingOptions {
	switch preset {
	case livekit.RecordingPreset_HD_30:
		return &livekit.RecordingOptions{
			InputWidth:     1280,
			InputHeight:    720,
			Depth:          24,
			Framerate:      30,
			AudioBitrate:   128,
			AudioFrequency: 44100,
			VideoBitrate:   3000,
		}
	case livekit.RecordingPreset_HD_60:
		return &livekit.RecordingOptions{
			InputWidth:     1280,
			InputHeight:    720,
			Depth:          24,
			Framerate:      60,
			AudioBitrate:   128,
			AudioFrequency: 44100,
			VideoBitrate:   4500,
		}
	case livekit.RecordingPreset_FULL_HD_30:
		return &livekit.RecordingOptions{
			InputWidth:     1920,
			InputHeight:    1080,
			Depth:          24,
			Framerate:      30,
			AudioBitrate:   128,
			AudioFrequency: 44100,
			VideoBitrate:   4500,
		}
	case livekit.RecordingPreset_FULL_HD_60:
		return &livekit.RecordingOptions{
			InputWidth:     1920,
			InputHeight:    1080,
			Depth:          24,
			Framerate:      60,
			AudioBitrate:   128,
			AudioFrequency: 44100,
			VideoBitrate:   6000,
		}
	default:
		return nil
	}
}
