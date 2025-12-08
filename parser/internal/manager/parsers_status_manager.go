// описание мэжнеджера состояния всех парсеров
// агрегирует состояния всех парсеров для дальнейшего использования этой информации в глобальном circuit breaker
package manager

import (
	"sync"
	"time"
)

// структура статуса отдельного парсера
type ParserStatus struct {
	Name         string    // имя парсера
	LastCheck    time.Time // время последней проверки статуса
	ErrorCount   int       // количество состояний, что парсер в ошибке
	SuccessCount int       // количество состояний, что парсер - без ошибок
	IsHealthy    bool      // состояние
	LastError    error     // последняя ошибка
	CircuitState string    // "closed", "open", "half-open" (состояние внутреннего circuit breaker)
}

// ParserStatusManager управляет статусами всех парсеров
type ParserStatusManager struct {
	mu      sync.RWMutex
	parsers map[string]*ParserStatus
}

// конструктор для нового менеджера парсеров
func NewParserStatusManager() *ParserStatusManager {
	return &ParserStatusManager{
		parsers: make(map[string]*ParserStatus),
	}
}

// UpdateStatus обновляет статус парсера в менеджере статуса парсеров
func (m *ParserStatusManager) UpdateStatus(name string, success bool, err error) {
	// так как мэнеджер статуса парсеров основан на мапе, все панипуляции проводит под мьютексом
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.parsers[name] // пытаемся получить статус парсера по ключу
	// если его нету, то добавляем новый в менеджер статуса парсеров
	if !exists {
		status = &ParserStatus{
			Name:      name,
			LastCheck: time.Now(),
		}
		m.parsers[name] = status
	}

	status.LastCheck = time.Now()

	if success {
		status.SuccessCount++
		status.ErrorCount = 0
		status.IsHealthy = true
		status.LastError = nil
	} else {
		status.ErrorCount++
		status.SuccessCount = 0
		status.IsHealthy = false
		status.LastError = err
	}
}

// GetHealthyParsers возвращает список здоровых парсеров
func (m *ParserStatusManager) GetHealthyParsers() []string {
	// так как мэнеджер статуса парсеров основан на мапе, все панипуляции проводит под мьютексом
	m.mu.Lock()
	defer m.mu.Unlock()

	var healthy []string
	for name, status := range m.parsers {
		// проверяем, что статус парсера IsHeathy==true,Lastcheck бы не позднее 5 мин
		if status.IsHealthy && time.Since(status.LastCheck) < 5*time.Minute {
			healthy = append(healthy, name)
		}
	}
	return healthy
}

// GetStatus возвращает статус конкретного парсера
func (m *ParserStatusManager) GetParserStatus(name string) (ParserStatus, bool) {
	// так как мэнеджер статуса парсеров основан на мапе, все панипуляции проводит под мьютексом
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.parsers[name]
	if !exists {
		return ParserStatus{}, false
	}
	return *status, true
}
