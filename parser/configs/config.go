package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Api_conf        apiConfig
	ParsConfAddress string // путь к .yml конфиг-файлу для экземпляров парсеров
	Cache_conf      cachesConfig
}

type apiConfig struct {
	ConcSearchCtxTimeOut string
}

type cachesConfig struct {
	NumOfShards     int
	SearchCacheTTL  time.Duration
	VacancyCacheTTL time.Duration
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("c:\\Son_Alex\\GO_projects\\go_v_1_23\\Job_Parser\\parser\\.env")
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// получаем количество шардов из .env (string ---> int)
	cachesNumOfShards, err := strconv.Atoi(os.Getenv("NUM_OF_SHARDS"))
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	searchCacheTTL, err := strconv.Atoi(os.Getenv("SEARCH_CACHE_TTL"))
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	vacancyCacheTTL, err := strconv.Atoi(os.Getenv("VACANCY_CACHE_TTL"))
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	return &Config{
		Api_conf: apiConfig{
			ConcSearchCtxTimeOut: os.Getenv("CONC_SEARCH_TIMEOUT"),
		},
		ParsConfAddress: os.Getenv("PARSES_CONFIG_ADDRESS_STRING"),
		Cache_conf: cachesConfig{
			NumOfShards:     cachesNumOfShards,
			SearchCacheTTL:  time.Duration(searchCacheTTL) * time.Second,
			VacancyCacheTTL: time.Duration(vacancyCacheTTL) * time.Second,
		},
	}, nil
}
