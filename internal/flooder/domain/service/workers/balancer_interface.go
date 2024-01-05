package workers

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"time"
)

type Balancer interface {
	CurrentAction() (action enum.Action, sleep time.Duration)
}
