package workers

type SenderInterface interface {
	IsMustBeSpawned() bool
	IsMustBeClosed() bool
}
