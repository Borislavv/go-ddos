package workers

type Balancer interface {
	IsMustBeSpawned() bool
	IsMustBeClosed() bool
}
