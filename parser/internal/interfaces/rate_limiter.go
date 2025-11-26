package interfaces

type RateLimiter interface {
	Wait() error
	Stop()
}
