// parser/config/config.go
package config

import (
	"parser/internal/circuitbreaker"
	"time"
)

type ParserConfig struct {
	HH       *ParserInstanceConfig `yaml:"hh"`
	SuperJob *ParserInstanceConfig `yaml:"superjob"`
}

type ParserInstanceConfig struct {
	Enabled               bool                                `yaml:"enabled"`
	BaseURL               string                              `yaml:"base_url"`
	APIKey                string                              `yaml:"api_key"`
	Timeout               time.Duration                       `yaml:"timeout"`
	RateLimit             time.Duration                       `yaml:"rate_limit"`
	MaxConcurrent         int                                 `yaml:"max_concurrent"`
	CircuitBreaker        circuitbreaker.CircuitBreakerConfig `yaml:"circuit_breaker"`
	MaxIdleConns          int                                 `yaml:"max_idle_conns"`
	IdleConnTimeout       time.Duration                       `yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout   time.Duration                       `yaml:"tls_handshake_timeout"`
	ResponseHeaderTimeout time.Duration                       `yaml:"response_header_timeout"`
	ExpectContinueTimeout time.Duration                       `yaml:"expect_continue_timeout"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *ParserConfig {
	return &ParserConfig{
		HH: &ParserInstanceConfig{
			Enabled:       true,
			BaseURL:       "https://api.hh.ru/vacancies",
			Timeout:       30 * time.Second,
			RateLimit:     2 * time.Second,
			MaxConcurrent: 10,
			CircuitBreaker: circuitbreaker.CircuitBreakerConfig{
				FailureThreshold:    5,
				SuccessThreshold:    3,
				HalfOpenMaxRequests: 2,
				ResetTimeout:        10 * time.Second,
				WindowDuration:      10 * time.Second,
			},
			MaxIdleConns:          5,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		SuperJob: &ParserInstanceConfig{
			Enabled:       true,
			BaseURL:       "https://api.superjob.ru/2.0/vacancies/",
			Timeout:       30 * time.Second,
			RateLimit:     2 * time.Second,
			MaxConcurrent: 10,
			CircuitBreaker: circuitbreaker.CircuitBreakerConfig{
				FailureThreshold:    5,
				SuccessThreshold:    3,
				HalfOpenMaxRequests: 2,
				ResetTimeout:        10 * time.Second,
				WindowDuration:      10 * time.Second,
			},
			MaxIdleConns:          5,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
