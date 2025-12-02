package parser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"parser/internal/circuitbreaker"
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
	sjCircuitBreaker interfaces.CBInterface
}

func NewSuperJobParser(apiKey string) *SuperJobParser {
	// создаём конфиг для SJ circuit breaker
	cbConfig := circuitbreaker.NewCircuitBreakerConfig(5, 3, 2, 10, 10) // [хардкодинг ---- плохо, нужно доделать!]
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
		sjCircuitBreaker: circuitbreaker.NewCircutBreaker(cbConfig),
	}
}

func (p *SuperJobParser) GetName() string {
	return "SuperJob.ru"
}

func (p *SuperJobParser) SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.Vacancy, error) {
	// Строим URL с параметрами
	apiURL, err := p.buildURL(params)
	if err != nil {
		return nil, fmt.Errorf("build URL failed: %w", err)
	}

	// Используем Circuit Breaker для выполнения запроса
	var vacancies []models.Vacancy

	err = p.sjCircuitBreaker.Execute(func() error {
		// ВСЁ, что связано с внешним вызовом API, внутри Execute

		// обрабатываем семафор
		select {
		case p.requestSemaphore <- struct{}{}:
			defer func() { <-p.requestSemaphore }()
		case <-ctx.Done():
			return fmt.Errorf("context canceled while waiting for semaphore: %w", ctx.Err())
		case <-time.After(2 * time.Second):
			return fmt.Errorf("semaphore timeout: HH API is busy, try again later")
		}

		// вызываем метод rate limiter до обращения к внешнему сервису
		p.sjRateLimiter.Wait()

		// Выполняем HTTP запрос
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return fmt.Errorf("create request failed: %w", err)
		}

		// Добавляем заголовки для SuperJob API
		req.Header.Add("X-Api-App-Id", p.apiKey)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := p.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("HTTP request failed: %w", err)
		}

		defer func() {
			io.Copy(io.Discard, resp.Body) // Сбрасываем тело для повторного использования соединения
			resp.Body.Close()
		}()

		// Проверяем статус ответа
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)

			// 5xx ошибки считаем как сбои для Circuit Breaker
			if resp.StatusCode >= 500 && resp.StatusCode < 600 {
				return fmt.Errorf("API server error %d: %s", resp.StatusCode, string(body))
			}
		}
		// Читаем и парсим ответ
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response failed: %w", err)
		}

		var searchResponse model.SuperJobResponse
		if err := json.Unmarshal(body, &searchResponse); err != nil {
			return fmt.Errorf("parse JSON failed: %w", err)
		}

		vacancies = p.convertToUniversal(searchResponse.Items)

		return nil
	})

	if err != nil {
		// Проверяем, это ошибка Circuit Breaker или ошибка API
		if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
			// Логируем состояние Circuit Breaker (пока в консоль) ------------------ ЛООООООООООООООГГГГГГГИИИИИИИИ

			tR, tS, tF := p.sjCircuitBreaker.GetStats()
			fmt.Printf("totalReq = %d, totalSuccess = %d, totalFailures = %d\n", tR, tS, tF)
			return nil, fmt.Errorf("HH API is temporarily unavailable (circuit breaker open). Please try again later")
		}
		return nil, fmt.Errorf("search vacancies failed: %w", err)
	}

	return vacancies, nil
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
