package vote

import "github.com/Borislavv/go-ddos/internal/flooder/domain/enum"

type Strategy interface {
	For() enum.Action
}
