package enum

type VoteStrategy string

const (
	AllVotersStrategy       VoteStrategy = "all"
	ManyVotersStrategy      VoteStrategy = "many"
	AtLeastOneVoterStrategy VoteStrategy = "at_least_one"
)

func (s VoteStrategy) String() string {
	return string(s)
}
