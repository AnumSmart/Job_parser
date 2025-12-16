package interfaces

// скорее всего названия методов  - поменяются !!!!!
type Job interface {
	Execute() (interface{}, error)
	GetSource() string
	GetPriority() int // приоритет для очереди
}
