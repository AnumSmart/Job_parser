package configs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Api_conf apiConfig
}

type apiConfig struct {
	SJAPIKey             string
	ConcSearchCtxTimeOut string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("c:\\Son_Alex\\Go_projects\\go_v_1_23\\Job_Parser\\parser\\.env")
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}
	return &Config{
		Api_conf: apiConfig{
			SJAPIKey:             os.Getenv("SJ_API_KEY"),
			ConcSearchCtxTimeOut: os.Getenv("CONC_SEARCH_TIMEOUT"),
		},
	}, nil
}
