package configuration

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Paths       Paths       `yaml:"paths"`
	FakeTunes   FakeTunes   `yaml:"faketunes"`
	Transcoding Transcoding `yaml:"transcoding"`
}

type FakeTunes struct {
	CacheSize int64        `yaml:"cache_size"`
	LogLevel  logrus.Level `yaml:"log_level"`
}

type Paths struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
}

type Transcoding struct {
	Parallel int64 `yaml:"parallel"`
}

func New() (*Config, error) {
	fakeTunesCfgPath := "/etc/faketunes.yaml"
	if customPath, ok := os.LookupEnv("FAKETUNES_CONFIG"); ok {
		fakeTunesCfgPath = customPath
	}

	rawConfig, err := os.ReadFile(fakeTunesCfgPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w (%w)", ErrConfiguration, ErrCantReadConfigFile, err)
	}

	config := new(Config)
	err = yaml.Unmarshal(rawConfig, config)
	if err != nil {
		return nil, fmt.Errorf("%w: %w (%w)", ErrConfiguration, ErrCantParseConfigFile, err)
	}

	return config, nil
}
