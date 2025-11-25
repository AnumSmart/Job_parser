package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"parser/internal/domain/models"
	"parser/internal/model"

	"net/http"
	"net/url"

	"strconv"
	"time"
)

// HHParser представляет парсер для HH.ru API
type HHParser struct {
	baseURL    string
	httpClient *http.Client
}

// NewHHParser создаёт новый экземпляр парсера (конструктор для парсера)
func NewHHParser() *HHParser {
	return &HHParser{
		baseURL: "https://api.hh.ru/vacancies",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *HHParser) SearchVacancies(params models.SearchParams) ([]models.Vacancy, error) {
	// Строим URL с параметрами
	apiURL, err := p.buildURL(params)
	if err != nil {
		return nil, fmt.Errorf("build URL failed: %w", err)
	}

	fmt.Printf("Created URL with params: %s\n", apiURL)

	// Выполняем HTTP запрос
	resp, err := p.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Читаем и парсим ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var searchResponse model.SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	return p.ConvertToUniversal(searchResponse.Items), nil
}

// Приводит структуры найденных результатов к универсальной структуре для всех парсеров
func (p *HHParser) ConvertToUniversal(hhVavancies []model.HHVacancy) []models.Vacancy {
	// сразу инициализируем слайс универсальных вакансий, чтобы уменьшить количество переаалокаций, если выйдем за размер базового массива слайса
	universalVacancies := make([]models.Vacancy, len(hhVavancies))

	for i, hhvacancy := range hhVavancies {
		// получаем строку-описания вилки зарплаты для каждой найденной записи
		salary := hhvacancy.GetSalaryString()

		universalVacancies[i] = models.Vacancy{
			ID:       hhvacancy.ID,
			Job:      hhvacancy.Name,
			Company:  hhvacancy.Employer.Name,
			Currency: hhvacancy.Salary.Currency,
			Salary:   &salary,
			Area:     hhvacancy.Area.Name,
			URL:      hhvacancy.URL,
			Seeker:   p.GetName(),
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
