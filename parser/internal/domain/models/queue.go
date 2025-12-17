package models

import "time"

// Структура задания для очереди в мэнеджере парсеров
type SearchJob struct {
	ID         string
	Params     SearchParams
	ResultChan chan *JobResult
	CreatedAt  time.Time
}

// Структура результата по выполнении работы
type JobResult struct {
	Results []SearchVacanciesResult
	Error   error
}
