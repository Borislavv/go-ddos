package reqbalancer

type SenderInterface interface {
	IsMustBeSpawned() bool
	IsMustBeClosed() bool
}
