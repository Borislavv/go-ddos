package enum

type Action string

const (
	Spawn Action = "spawn"
	Close Action = "close"
	Await Action = "await"
)

func (a Action) String() string {
	return string(a)
}
