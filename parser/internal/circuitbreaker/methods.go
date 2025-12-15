package circuitbreaker

import "time"

// Execute выполняет операцию с защитой Circuit Breaker
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// так как этот метод будут использовать асинхронно, делаем дальнейшие операции из-под мьютекса
	cb.mu.Lock()

	// проеряем состояние (или идём дальше, или возвращаем ошибку)
	switch cb.state {
	// случай - открытого circut breaker
	case StateOpen:
		// Проверяем, не истекло ли время resetTimeout
		if time.Since(cb.lastFailureTime) < cb.resetTimeout {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}

		// если время - истекло, то переходим в Half-Open state
		// Переходим в Half-Open
		cb.state = StateHalfOpen // меняем состояние на полу-открытое
		cb.halfOpenAttempts = 0  // утсанавливаем счётчик попыток в полу-открытом состоянии в 0
		cb.successes = 0         // утсанавливаем счётчик успешных попыток попыток в полу-открытом состоянии в 0
		//// случай - полу-открытого circut breaker
	case StateHalfOpen:
		if cb.halfOpenAttempts >= cb.halfOpenMaxRequests {
			cb.mu.Unlock()
			return ErrTooManyRequests
		}
		cb.halfOpenAttempts++
	}

	// если нет ошибок, circut breaker или в закрытом состоянии или в полу-открытом
	// добавляем +1 к статистике запросов
	cb.totalRequests++
	cb.mu.Unlock()

	// Выполняем операцию
	err := fn()

	// вызываем мьютекс, так как меняем статус circut breaker фсинхронно
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Обрабатываем результат
	if err != nil {
		cb.totalFailures++
		cb.onFailure()
		return err
	}

	cb.totalSuccesses++
	cb.onSuccess()
	return nil
}

// onFailure обрабатывает неудачное выполнение
func (cb *CircuitBreaker) onFailure() {
	// проеряем статус circut breaker.
	// тут мьютекс не нужен, так как внешний вызов уже из-под мьютекса
	switch cb.state {
	// если circut breaker - закрыт (все в порядке)
	case StateClosed:
		cb.failures++ // увеличиваем счётчик ошибок
		if cb.failures >= cb.failureThreshold {
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
		}

	case StateHalfOpen:
		// При ошибке в Half-Open возвращаемся в Open
		cb.state = StateOpen
		cb.lastFailureTime = time.Now()
		cb.halfOpenAttempts = 0
	}
}

// onFailure обрабатывает удачное выполнение
func (cb *CircuitBreaker) onSuccess() {
	// проеряем статус circut breaker.
	// тут мьютекс не нужен, так как внешний вызов уже из-под мьютекса
	switch cb.state {
	case StateClosed:
		// Сбрасываем счетчик ошибок при успешном выполнении
		cb.failures = 0

	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.successThreshold {
			// Переходим в Closed состояние
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.halfOpenAttempts = 0
		}
	}
}

// GetState возвращает текущее состояние
func (cb *CircuitBreaker) getState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats возвращает статистику
func (cb *CircuitBreaker) GetStats() (total, success, failure uint) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.totalRequests, cb.totalSuccesses, cb.totalFailures
}
