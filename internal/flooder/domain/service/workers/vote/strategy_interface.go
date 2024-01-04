package vote

type Strategy interface {
	IsFor() bool
}
