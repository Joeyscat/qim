package wire

import (
	"math"
	"sync/atomic"
)

type seqeunce struct {
	num uint32
}

func (s *seqeunce) Next() uint32 {
	next := atomic.AddUint32(&s.num, 1)
	if next == math.MaxUint32 {
		if atomic.CompareAndSwapUint32(&s.num, next, 1) {
			return 1
		}
		return s.Next()
	}
	return next
}

var Seq = seqeunce{num: 1}
