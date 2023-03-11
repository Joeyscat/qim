package container

import (
	"hash/crc32"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/wire/pkt"
)

func HashCode(key string) int {
	hash32 := crc32.NewIEEE()
	hash32.Write([]byte(key))
	return int(hash32.Sum32())
}

type Selector interface {
	Lookup(header *pkt.Header, srvs []qim.Service) string
}
