package vote_enum

type VoteStrategy string

const (
	AllVoters       VoteStrategy = "all"
	ManyVoters      VoteStrategy = "many"
	AtLeastOneVoter VoteStrategy = "at_least_one"
)

func (s VoteStrategy) String() string {
	return string(s)
}
