package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"parser/internal/domain/models"
	"parser/internal/interfaces"
	"parser/internal/model"
	ratelimiter "parser/internal/rate_limiter"

	"strconv"
	"time"
)

const (
	sjRateLimit = 2 * time.Second
)

type SuperJobParser struct {
	baseURL          string
	apiKey           string
	httpClient       *http.Client
	sjRateLimiter    interfaces.RateLimiter
	requestSemaphore chan struct{} // буфер: 10-15, Дополнительный семафор для парсера
}

func NewSuperJobParser(apiKey string) *SuperJobParser {
	return &SuperJobParser{
		baseURL: "https://api.superjob.ru/2.0/vacancies/",
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				// Согласовано с размером семафора!
				MaxConnsPerHost:       10, // Столько же, сколько семафор
				MaxIdleConnsPerHost:   5,  // Половина от активных
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		sjRateLimiter:    ratelimiter.NewChannelRateLimiter(sjRateLimit),
		requestSemaphore: make(chan struct{}, 10), // буфер: 10-15, Дополнительный семафор для парсера
	}
}

func (p *SuperJobParser) GetName() string {
	return "SuperJob.ru"
}

func (p *SuperJobParser) SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.Vacancy, error) {
	apiURL, err := p.buildURL(params)
	if err != nil {
		return nil, fmt.Errorf("build URL failed: %w", err)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Добавляем заголовки для SuperJob API
	req.Header.Add("X-Api-App-Id", p.apiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// обрабатываем семафор
	select {
	case p.requestSemaphore <- struct{}{}:
		defer func() { <-p.requestSemaphore }()
	case <-ctx.Done():
		return nil, fmt.Errorf("Context HH error: %w", ctx.Err())
	case <-time.After(2 * time.Second):
		return nil, fmt.Errorf("SJ API is busy, try again later")
	}

	// вызываем метод rate limiter до обращения к внешнему сервису
	p.sjRateLimiter.Wait()

	// Выполняем HTTP запрос
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var sjResponse model.SuperJobResponse
	if err := json.Unmarshal(body, &sjResponse); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	return p.convertToUniversal(sjResponse.Items), nil
}

func (p *SuperJobParser) buildURL(params models.SearchParams) (string, error) {
	u, err := url.Parse(p.baseURL)
	if err != nil {
		return "", err
	}

	query := u.Query()

	if params.Text != "" {
		query.Set("keyword", params.Text)
	}
	if params.Area != "" {
		query.Set("town", p.convertArea(params.Area))
	}
	if params.PerPage > 0 {
		query.Set("count", strconv.Itoa(params.PerPage))
	}
	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page-1)) // SuperJob использует 0-based
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (p *SuperJobParser) convertArea(area string) string {
	// Конвертируем коды регионов HH.ru в названия SuperJob
	areas := map[string]string{
		"1": "Москва",
		"2": "Санкт-Петербург",
	}
	if name, ok := areas[area]; ok {
		return name
	}
	return ""
}

func (p *SuperJobParser) convertToUniversal(sjVacancies []model.SJVacancy) []models.Vacancy {
	vacancies := make([]models.Vacancy, len(sjVacancies))
	for i, sjv := range sjVacancies {
		salary := sjv.GetSalaryString()
		vacancies[i] = models.Vacancy{
			ID:          strconv.Itoa(sjv.ID),
			Job:         sjv.Profession,
			Company:     sjv.FirmName,
			Currency:    sjv.Currency,
			Salary:      &salary,
			Area:        sjv.Town.Title,
			URL:         sjv.Link,
			Seeker:      p.GetName(),
			Description: sjv.VacancyRichText,
		}
	}
	return vacancies
}

func (p *SuperJobParser) GetVacancyByID(vacancyID string) (*model.SJVacancy, error) {
	// Реализация получения деталей вакансии по ID
	if vacancyID == "" {
		return nil, fmt.Errorf("vacancy ID cannot be empty")
	}

	apiURL := p.baseURL + "/" + vacancyID
	resp, err := p.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// если API - вернул ошибку, прерываем функцию
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	//анмаршалим успешное тело ответа в в нужную структуру
	var vacancy model.SJVacancy
	if err := json.Unmarshal(body, &vacancy); err != nil {
		return nil, fmt.Errorf("parse SJ-JSON failed: %w", err)
	}

	return &vacancy, nil
}
