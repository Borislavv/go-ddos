package voter

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
)

type Voter interface {
	Vote() (weight enum.Weight)
}
