package qim

import (
	"sync"
)

type ChannelMap interface {
	// add channel
	Add(ch Channel)
	// remove channel by id
	Remove(id string)
	// get channel by id
	Get(id string) (Channel, bool)
	// return all channels
	All() []Channel
}

type ChannlesImpl struct {
	channels *sync.Map
}

// Add implements ChannelMap
func (cs *ChannlesImpl) Add(c Channel) {
	cs.channels.Store(c.ID(), c)
}

// All implements ChannelMap
func (cs *ChannlesImpl) All() []Channel {
	arr := make([]Channel, 0)
	cs.channels.Range(func(key, value any) bool {
		arr = append(arr, value.(Channel))
		return true
	})
	return arr
}

// Get implements ChannelMap
func (cs *ChannlesImpl) Get(id string) (Channel, bool) {
	val, ok := cs.channels.Load(id)
	if !ok {
		return nil, false
	}
	return val.(Channel), true
}

// Remove implements ChannelMap
func (cs *ChannlesImpl) Remove(id string) {
	cs.channels.Delete(id)
}

func NewChannels(num int) ChannelMap {
	return &ChannlesImpl{
		channels: new(sync.Map),
	}
}
