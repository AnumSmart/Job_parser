// описание мэжнеджера состояния всех парсеров
// агрегирует состояния всех парсеров для дальнейшего использования этой информации в глобальном circuit breaker
package parsers_status_manager

import (
	"parser/configs"
	"parser/internal/interfaces"
	"sync"
	"time"
)

// структура статуса отдельного парсера
type ParserStatus struct {
	Name           string        // имя парсера
	LastCheck      time.Time     // время последней проверки статуса
	LastSuccess    time.Time     // время последней успешной проверки
	ErrorCount     int           // количество состояний, что парсер в ошибке
	SuccessCount   int           // количество состояний, что парсер - без ошибок
	IsHealthy      bool          // состояние
	LastError      error         // последняя ошибка
	CircuitState   string        // "closed", "open", "half-open" (состояние внутреннего circuit breaker)
	initialized    bool          // false - просто создан парсер, true - была попытка запроса
	HealthEndpoint string        // URL для health check
	ResponseTime   time.Duration // время ответа от парсера
	mu             sync.RWMutex
}

// ParserStatusManager управляет статусами всех парсеров
// ключ - это имя экземпляра парсера
type ParserStatusManager struct {
	parsers      map[string]*ParserStatus   // мапа статусов парсеров
	config       *configs.HealthCheckConfig // конфиг мэнеджера статусов
	client       interfaces.HealthClient    // клиент для проверки
	initComplete chan struct{}              // Сигнал завершения инициализации
	mu           sync.RWMutex
}

// конструктор для нового менеджера статусов парсеров--------------------------------!!!!!!!!!!!!!!!!!!!!!!!!!!!
func NewParserStatusManager(conf *configs.HealthCheckConfig, parsers ...interfaces.Parser) *ParserStatusManager {
	psm := &ParserStatusManager{
		parsers:      make(map[string]*ParserStatus),
		config:       conf, // конфиг для коиента health check
		client:       NewHttpHealthCheckClient(conf),
		initComplete: make(chan struct{}),
	}
	var parsersNames []string

	// собираем имена всех парсеров
	for _, parser := range parsers {
		parsersNames = append(parsersNames, parser.GetName())
	}

	// инициализируем статусы парсеров в менеджере статусов
	for _, name := range parsersNames {
		psm.parsers[name] = &ParserStatus{
			Name:         name,
			LastCheck:    time.Now(),
			initialized:  false,
			CircuitState: "closed",
		}
	}
	return psm
}

// UpdateStatus обновляет статус парсера в менеджере статуса парсеров (потокобезопасен, есть мьютекс внутри)
func (m *ParserStatusManager) UpdateStatus(name string, success bool, err error) {
	// так как мэнеджер статуса парсеров основан на мапе, все панипуляции проводит под мьютексом
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.parsers[name] // пытаемся получить статус парсера по ключу
	// если его нету, то добавляем новый в менеджер статуса парсеров
	if !exists {
		status = &ParserStatus{
			Name:        name,
			initialized: true,
			LastCheck:   time.Now(),
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
		if !status.initialized {
			// Если статус неизвестен, считаем доступным для первой попытки
			healthy = append(healthy, name)
		} else if status.IsHealthy && time.Since(status.LastCheck) < 5*time.Minute {
			// проверяем, что статус парсера IsHeathy==true,Lastcheck бы не позднее 5 мин
			healthy = append(healthy, name)
		}

	}
	return healthy
}

// GetStatus возвращает статус конкретного парсера
func (m *ParserStatusManager) GetParserStatus(name string) (*ParserStatus, bool) {
	// так как мэнеджер статуса парсеров основан на мапе, все панипуляции проводит под мьютексом
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.parsers[name]
	if !exists {
		return nil, false
	}
	return status, true
}
