package configs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	API         APIConfig
	Cache       *CachesConfig
	Parsers     *ParsersConfig
	Manager     *ParserManagerConfig
	HealthChech *HealthCheckConfig
}

type APIConfig struct {
	ConcSearchTimeout time.Duration
	ServerPort        string
}

// загружаем конфиг-данные из .env
func LoadConfig() (*Config, error) {
	err := godotenv.Load("c:\\Son_Alex\\GO_projects\\go_v_1_23\\Job_Parser\\parser\\.env")
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	concSearchTimeOut, err := strconv.Atoi(os.Getenv("CONC_SEARCH_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	cacheConfig, err := LoadYAMLConfig[CachesConfig](os.Getenv("CACHES_CONFIG_ADDRESS_STRING"), DefaultCacheConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	parsersConfig, err := LoadYAMLConfig[ParsersConfig](os.Getenv("PARSERS_CONFIG_ADDRESS_STRING"), DefaultParsersConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	parsersManagerConfig, err := LoadYAMLConfig[ParserManagerConfig](os.Getenv("PARSERS_CONFIG_ADDRESS_STRING"), DefaultParsersManagerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	healthCheckConfig, err := LoadYAMLConfig[HealthCheckConfig](os.Getenv("HEALTH_CHECK_CONFIG_ADDRESS_STRING"), DefaultHealthCheckConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	return &Config{
		API: APIConfig{
			ConcSearchTimeout: time.Duration(concSearchTimeOut) * time.Second,
		},
		Cache:       cacheConfig,
		Parsers:     parsersConfig,
		Manager:     parsersManagerConfig,
		HealthChech: healthCheckConfig,
	}, nil
}

// универсальня функция загрузки конфига из .yml файла (используем дженерики, так как будут ещё парсеры)
func LoadYAMLConfig[T any](configPath string, fn func() *T) (*T, error) {
	config := fn()

	if configPath == "" {
		return config, nil
	}

	if _, err := os.Stat(configPath); errors.Is(err, fs.ErrNotExist) {
		return config, nil
	}

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}

	return config, nil
}
