package interfaces

type CBInterface interface {
	Execute(fn func() error) error
	GetStats() (total, success, failure uint)
}
