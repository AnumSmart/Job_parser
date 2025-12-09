package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"parser/internal/domain/models"
	"parser/internal/model"
	"parser/internal/parser/config"
	"strconv"
)

type HH2Parser struct {
	*BaseParser
}

func NewHH2Parser(cfg *config.ParserInstanceConfig) *HH2Parser {
	if cfg == nil {
		cfg = config.DefaultConfig().HH
	}

	baseCfg := BaseConfig{
		Name:                  "HH.ru",
		BaseURL:               cfg.BaseURL,
		Timeout:               cfg.Timeout,
		RateLimit:             cfg.RateLimit,
		MaxConcurrent:         cfg.MaxConcurrent,
		CircuitBreakerCfg:     cfg.CircuitBreaker,
		MaxIdleConns:          cfg.MaxIdleConns,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout,
	}

	return &HH2Parser{
		BaseParser: NewBaseParser(baseCfg),
	}
}

func (p *HH2Parser) SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.Vacancy, error) {
	return p.BaseParser.SearchVacancies(
		ctx,
		params,
		ParserFuncs{
			BuildURL: p.buildURL,
			Parse:    p.parseResponse,
			Convert:  p.convertToUniversal,
		},
	)
}

// buildURL строит URL для API запроса
func (p *HH2Parser) buildURL(params models.SearchParams) (string, error) {
	u, err := url.Parse(p.baseURL)
	if err != nil {
		return "", err
	}

	query := u.Query()

	if params.Text != "" {
		query.Set("text", params.Text)
	}
	if params.Area != "" {
		query.Set("area", params.Area)
	}

	// Устанавливаем количество вакансий на странице
	perPage := params.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 20 // Значение по умолчанию
	}
	query.Set("per_page", strconv.Itoa(perPage))

	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page))
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

// метод парсера обработки тела запроса
func (p *HH2Parser) parseResponse(body []byte) (interface{}, error) {
	var searchResponse model.SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("[Parser name: %s] parse reaponse body - failed: %w", p.name, err)
	}
	return &searchResponse, nil
}

// метод приведения результатов поиска у унифицированной структуре + проверка данных их интерфейса
func (p *HH2Parser) convertToUniversal(searchResponse interface{}) ([]models.Vacancy, error) {
	hhVavancies, ok := searchResponse.([]model.HHVacancy)
	if !ok {
		return nil, fmt.Errorf("[Parser name: %s], wrong data type in the response body\n", p.name)
	}

	// сразу инициализируем слайс универсальных вакансий, чтобы уменьшить количество переаалокаций, если выйдем за размер базового массива слайса
	universalVacancies := make([]models.Vacancy, len(hhVavancies))

	for i, hhvacancy := range hhVavancies {
		// получаем строку-описания вилки зарплаты для каждой найденной записи
		salary := hhvacancy.GetSalaryString()

		universalVacancies[i] = models.Vacancy{
			ID:          hhvacancy.ID,
			Job:         hhvacancy.Name,
			Company:     hhvacancy.Employer.Name,
			Currency:    hhvacancy.Salary.Currency,
			Salary:      &salary,
			Area:        hhvacancy.Area.Name,
			URL:         hhvacancy.URL,
			Seeker:      p.GetName(),
			Description: hhvacancy.Description,
		}
	}
	return universalVacancies, nil
}
