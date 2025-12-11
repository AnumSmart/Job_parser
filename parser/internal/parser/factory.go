// описание структуры для создания механизма фабрики для парсеров
// на вход фабрики подаётся тип парсера, конфиг для нужного парсера и конструктор для нужного парсера
package parser

import (
	"fmt"

	"parser/configs"
	"parser/internal/interfaces"
	"sync"
)

//Фабрика с регистрацией парсеров

// ParserType тип парсера
type ParserType string

const (
	ParserTypeHH ParserType = "hh"
	ParserTypeSJ ParserType = "superjob"
	// можно добавить: ParserTypeRabotaRu ParserType = "rabota.ru"
)

// ParserConstructor функция-конструктор парсера
type ParserConstructor func(config *configs.ParserInstanceConfig) interfaces.Parser

// ParserFactory фабрика парсеров
type ParserFactory struct {
	constructors map[ParserType]ParserConstructor
	configs      map[ParserType]*configs.ParserInstanceConfig
	mu           sync.RWMutex
}

// NewParserFactory создает новую фабрику
func NewParserFactory() *ParserFactory {
	return &ParserFactory{
		constructors: make(map[ParserType]ParserConstructor),
		configs:      make(map[ParserType]*configs.ParserInstanceConfig),
	}
}

// Register регистрирует конструктор парсера и конфиг
func (f *ParserFactory) Register(parserType ParserType, config *configs.ParserInstanceConfig, constructor ParserConstructor) {
	// так как есть конкурентный доступ к мапе - делаем черезе мьютекс
	f.mu.Lock()
	defer f.mu.Unlock()

	f.constructors[parserType] = constructor
	f.configs[parserType] = config
}

// Create - создает парсер, если вся инфа до этого была зарегестрирована в фабрике
func (f *ParserFactory) Create(parserType ParserType) (interfaces.Parser, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	// под защитой мьютекса проверяем, есть ли в фабрике зарегестрированный конструктор для данного типа парсера
	constructor, ok := f.constructors[parserType]
	if !ok {
		return nil, fmt.Errorf("parser type not registered: %s", parserType)
	}
	// под защитой мьютекса проверяем, есть ли в фабрике зарегестрированный конфиг для данного типа парсера
	config, ok := f.configs[parserType]
	if !ok {
		return nil, fmt.Errorf("config not found for parser: %s", parserType)
	}
	return constructor(config), nil
}

// CreateEnabled создает только включенные парсеры
func (f *ParserFactory) CreateEnabled(enabled []ParserType) ([]interfaces.Parser, error) {
	parsers := make([]interfaces.Parser, len(enabled))

	for i, parserType := range enabled {
		parser, err := f.Create(parserType)
		if err != nil {
			return nil, fmt.Errorf("failed to create parser %s: %w", parserType, err)
		}
		parsers[i] = parser
	}

	return parsers, nil
}
