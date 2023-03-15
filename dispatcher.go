package qim

import "github.com/joeyscat/qim/wire/pkt"

// The Dispatcher is responsible for dispatch messages to the gateway.
type Dispatcher interface {
	Push(gateway string, channels []string, p *pkt.LogicPkt) error
}
