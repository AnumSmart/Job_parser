package parser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"parser/internal/circuitbreaker"
	"parser/internal/domain/models"
	"parser/internal/interfaces"
	"parser/internal/model"
	ratelimiter "parser/internal/rate_limiter"

	"net/http"
	"net/url"

	"strconv"
	"time"
)

const (
	hhRateLimit = 2 * time.Second
)

// HHParser представляет парсер для HH.ru API
type HHParser struct {
	baseURL          string
	httpClient       *http.Client
	hhRateLimiter    interfaces.RateLimiter
	requestSemaphore chan struct{} // буфер: 10-15, Дополнительный семафор для парсера
	hhCircuitBreaker interfaces.CBInterface
}

// NewHHParser создаёт новый экземпляр парсера (конструктор для парсера)
func NewHHParser() *HHParser {
	// создаём конфиг для HH circuit breaker
	cbConfig := circuitbreaker.NewCircuitBreakerConfig(5, 3, 2, 10, 10) // [хардкодинг ---- плохо, нужно доделать!]
	return &HHParser{
		baseURL: "https://api.hh.ru/vacancies",
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
		hhRateLimiter:    ratelimiter.NewChannelRateLimiter(hhRateLimit),
		requestSemaphore: make(chan struct{}, 10),
		hhCircuitBreaker: circuitbreaker.NewCircutBreaker(cbConfig),
	}
}

func (p *HHParser) SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.Vacancy, error) {
	// Строим URL с параметрами
	apiURL, err := p.buildURL(params)
	if err != nil {
		return nil, fmt.Errorf("build URL failed: %w", err)
	}

	// Используем Circuit Breaker для выполнения запроса
	var vacancies []models.Vacancy

	err = p.hhCircuitBreaker.Execute(func() error {
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
		p.hhRateLimiter.Wait()

		// Выполняем HTTP запрос
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return fmt.Errorf("create request failed: %w", err)
		}

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

		var searchResponse model.SearchResponse
		if err := json.Unmarshal(body, &searchResponse); err != nil {
			return fmt.Errorf("parse JSON failed: %w", err)
		}

		vacancies = p.ConvertToUniversal(searchResponse.Items)

		return nil
	})

	if err != nil {
		// Проверяем, это ошибка Circuit Breaker или ошибка API
		if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
			// Логируем состояние Circuit Breaker (пока в консоль) ------------------ ЛООООООООООООООГГГГГГГИИИИИИИИ

			tR, tS, tF := p.hhCircuitBreaker.GetStats()
			fmt.Printf("totalReq = %d, totalSuccess = %d, totalFailures = %d\n", tR, tS, tF)
			return nil, fmt.Errorf("HH API is temporarily unavailable (circuit breaker open). Please try again later")
		}
		return nil, fmt.Errorf("search vacancies failed: %w", err)
	}

	return vacancies, nil
}

// Приводит структуры найденных результатов к универсальной структуре для всех парсеров
func (p *HHParser) ConvertToUniversal(hhVavancies []model.HHVacancy) []models.Vacancy {
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
	return universalVacancies
}

// buildURL строит URL для API запроса
func (p *HHParser) buildURL(params models.SearchParams) (string, error) {
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

// GetVacancyByID получает детальную информацию о вакансии по ID
func (p *HHParser) GetVacancyByID(vacancyID string) (*model.HHVacancy, error) {
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

	var vacancy model.HHVacancy
	//анмаршалим успешное тело ответа в в нужную структуру
	if err := json.Unmarshal(body, &vacancy); err != nil {
		return nil, fmt.Errorf("parse HH-JSON failed: %w", err)
	}

	return &vacancy, nil
}

// получаем имя парсера
func (p *HHParser) GetName() string {
	return "HH.ru"
}
