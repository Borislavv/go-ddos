package voter

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"time"
)

type Voter interface {
	Vote() (weight enum.Weight, sleep time.Duration)
}
