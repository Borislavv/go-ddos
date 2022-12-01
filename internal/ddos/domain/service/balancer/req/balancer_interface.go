package reqsender

type SenderInterface interface {
	IsMustBeSpawned() bool
	IsMustBeClosed() bool
}
