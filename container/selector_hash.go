package container

import (
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/wire/pkt"
)

type HashSelector struct {
}

// Lookup implements Selector
func (*HashSelector) Lookup(header *pkt.Header, srvs []qim.Service) string {
	ll := len(srvs)
	code := HashCode(header.ChannelId)
	return srvs[code%ll].ServiceID()
}

var _ Selector = (*HashSelector)(nil)
