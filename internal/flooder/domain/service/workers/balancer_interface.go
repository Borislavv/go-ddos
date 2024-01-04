package workers

import "github.com/Borislavv/go-ddos/internal/flooder/domain/enum"

type Balancer interface {
	CurrentAction() enum.Action
}
