package model

// Структуры для SuperJob API
type SuperJobResponse struct {
	Items []SuperJobVacancy `json:"objects"`
	Total int               `json:"total"`
}

type SuperJobVacancy struct {
	ID              int    `json:"id"`
	Profession      string `json:"profession"`
	FirmName        string `json:"firm_name"`
	PaymentFrom     int    `json:"payment_from"`
	PaymentTo       int    `json:"payment_to"`
	Currency        string `json:"currency"`
	Town            Town   `json:"town"`
	Link            string `json:"link"`
	VacancyRichText string `json:"vacancyRichText"`
}

type Town struct {
	Title string `json:"title"`
}
