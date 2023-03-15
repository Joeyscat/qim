package qim

import (
	"errors"

	"github.com/joeyscat/qim/wire/pkt"
)

var ErrSessionNil = errors.New("err:session nil")

type SessionStorage interface {
	Add(session *pkt.Session) error
	Delete(account, channelID string) error
	Get(channelID string) (*pkt.Session, error)
	GetLocations(account ...string) ([]*Location, error)
	GetLocation(account, device string) (*Location, error)
}
