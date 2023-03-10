package qim

import (
	"bytes"
	"errors"

	"github.com/joeyscat/qim/wire/endian"
)

type Location struct {
	ChannelID string
	GateID    string
}

func (loc *Location) Bytes() []byte {
	if loc == nil {
		return []byte{}
	}
	buf := new(bytes.Buffer)
	_ = endian.WriteShortBytes(buf, []byte(loc.ChannelID))
	_ = endian.WriteShortBytes(buf, []byte(loc.GateID))
	return buf.Bytes()
}

func (loc *Location) Unmarshal(data []byte) (err error) {
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	buf := bytes.NewBuffer(data)
	loc.ChannelID, err = endian.ReadShortString(buf)
	if err != nil {
		return
	}
	loc.GateID, err = endian.ReadShortString(buf)
	return
}
