package vote

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"time"
)

type Strategy interface {
	For() (action enum.Action, sleep time.Duration)
}
