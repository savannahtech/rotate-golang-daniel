package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

const (
	ConfigName = "config"
	ConfigPath = "."
	ConfigType = "yaml"

	DefaultHTTPPort = "9000"

	LogsDBName         = "logsDB"
	LogsCollectionName = "logs"
)

type Config struct {
	Directory      string `validate:"required"`
	CheckFrequency int    `validate:"required,min=1"`
	ReportingAPI   string `validate:"required,url"`
	HTTPPort       string `validate:"required"`
	SocketPath     string `validate:"required"`
	MongoURI       string `validate:"required"`
}

func LoadConfig(name, path string) (*Config, error) {
	viper.SetConfigName(name)
	viper.SetConfigType(ConfigType)
	viper.AddConfigPath(path)

	viper.SetDefault("http_port", DefaultHTTPPort)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config.yaml: %w", err)
	}

	sanitizedDir := strings.ReplaceAll(viper.GetString("directory"), "'", "''")
	validPath := regexp.MustCompile(`^[a-zA-Z0-9/_-]+$`)
	if !validPath.MatchString(sanitizedDir) {
		return nil, fmt.Errorf("invalid directory format")
	}

	cfg := &Config{
		Directory:      sanitizedDir,
		CheckFrequency: viper.GetInt("check_frequency"),
		ReportingAPI:   viper.GetString("reporting_api"),
		HTTPPort:       viper.GetString("http_port"),
		SocketPath:     viper.GetString("socket_path"),
		MongoURI:       viper.GetString("mongo_uri"),
	}

	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return cfg, nil
}
