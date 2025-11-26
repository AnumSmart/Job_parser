package ratelimiter

import (
	"context"
	"errors"
	"sync"
	"time"
)

// структура rate limiter, онованная на канале
type ChannelRateLimiter struct {
	limiter chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
	stopped bool
}

// создаём конструктор rate limiter, с указанием интервала между запросами
// внутри запускается 1 горутина, внутри которой работает тикер
// он через заданный rate пишет в канал пустую стуктуру.
func NewChannelRateLimiter(rate time.Duration) *ChannelRateLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	rl := &ChannelRateLimiter{
		limiter: make(chan struct{}, 1),
		ctx:     ctx,
		cancel:  cancel,
	}

	go rl.run(rate)

	return rl
}

// метод rate limiter, в котором запускается тикер, где идёт ограничение через заданные интервал
func (rl *ChannelRateLimiter) run(rate time.Duration) {
	ticker := time.NewTicker(rate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Проверяем не остановлен ли лимитер
			rl.mu.RLock()
			stopped := rl.stopped
			rl.mu.RUnlock()

			if stopped {
				return
			}

			// Non-blocking send благодаря буферу и select+default
			select {
			case rl.limiter <- struct{}{}:
				// успешно добавили токен в буфер
			case <-rl.ctx.Done():
				return
			default:
				// буфер полон (уже есть токен), пропускаем
				// это предотвращает накопление "долгов"
			}
		}
	}
}

// метод rate limiter, ожидание, пока не будет доступен токен для чтения
func (rl *ChannelRateLimiter) Wait() error {
	// проверяем, что rate limiter - не остановлен
	// делаем это из-под мьютекса (конкурентный доступ)
	rl.mu.RLock()
	stopped := rl.stopped
	rl.mu.RUnlock()

	if stopped {
		return errors.New("rate limiter stopped")
	}

	select {
	// проверяем отмену контекста
	case <-rl.ctx.Done():
		return errors.New("rate limiter stopped")
		// пробуем читать из канала rate limiter и проверяем закрыт ли этот канал
	case _, ok := <-rl.limiter:
		if !ok {
			return errors.New("rate limiter stopped")
		}
		return nil
	}
}

// метод rate limiter,который останаваливает rate limiter
func (rl *ChannelRateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if !rl.stopped {
		rl.stopped = true
		rl.cancel()
		close(rl.limiter) // закрываем канал
	}
}
