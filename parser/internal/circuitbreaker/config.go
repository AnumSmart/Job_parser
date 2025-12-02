package circuitbreaker

import "time"

// CircuitBreakerConfig - конфигурация Circuit Breaker
type CircuitBreakerConfig struct {
	FailureThreshold    uint
	SuccessThreshold    uint
	HalfOpenMaxRequests uint
	ResetTimeout        time.Duration
	WindowDuration      time.Duration
}

// создаём конструктор для конфига circuit breaker
// будем возвращать копию структуры, так как буту разные конфиги
func NewCircuitBreakerConfig(fTreshold, sTreshold, halfTreshold uint, resetTimeout, winDuration time.Duration) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:    fTreshold,
		SuccessThreshold:    sTreshold,
		HalfOpenMaxRequests: halfTreshold,
		ResetTimeout:        resetTimeout,
		WindowDuration:      winDuration,
	}
}
