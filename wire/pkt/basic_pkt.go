package pkt

import (
	"io"

	"github.com/joeyscat/qim/wire/endian"
)

// basic pkt code
const (
	CodePing = uint16(1)
	CodePong = uint16(2)
)

type BasicPkt struct {
	Code   uint16
	Length uint16
	Body   []byte
}

// Decode implements Packet
func (p *BasicPkt) Decode(r io.Reader) error {
	var err error
	if p.Code, err = endian.ReadUint16(r); err != nil {
		return err
	}
	if p.Length, err = endian.ReadUint16(r); err != nil {
		return err
	}
	if p.Length > 0 {
		p.Body, err = endian.ReadFixedBytes(int(p.Length), r)
		return err
	}
	return nil
}

// Encode implements Packet
func (p *BasicPkt) Encode(w io.Writer) error {
	if err := endian.WriteUint16(w, p.Code); err != nil {
		return err
	}
	if err := endian.WriteUint16(w, p.Length); err != nil {
		return err
	}
	if p.Length > 0 {
		_, err := w.Write(p.Body)
		return err
	}

	return nil
}

var _ Packet = (*BasicPkt)(nil)
