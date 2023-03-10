package pkt

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/joeyscat/qim/wire"
	"github.com/joeyscat/qim/wire/endian"
	"google.golang.org/protobuf/proto"
)

// LogicPkt 定义了网关对外的client消息结构
type LogicPkt struct {
	Header
	Body []byte `json:"body,omitempty"`
}

type HeaderOption func(*Header)

func WithStatus(status Status) HeaderOption {
	return func(h *Header) {
		h.Status = status
	}
}

func WithSeq(seq uint32) HeaderOption {
	return func(h *Header) {
		h.Sequence = seq
	}
}

func WithChannel(channelID string) HeaderOption {
	return func(h *Header) {
		h.ChannelId = channelID
	}
}

func WithDest(dest string) HeaderOption {
	return func(h *Header) {
		h.Dest = dest
	}
}

func New(command string, options ...HeaderOption) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Command = command

	for _, option := range options {
		option(&pkt.Header)
	}
	if pkt.Sequence == 0 {
		pkt.Sequence = wire.Seq.Next()
	}
	return pkt
}

func NewFrom(header *Header) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Header = Header{
		Command:   header.Command,
		Sequence:  header.Sequence,
		ChannelId: header.ChannelId,
		Status:    header.Status,
		Dest:      header.Dest,
	}
	return pkt
}

// Decode implements Packet
func (p *LogicPkt) Decode(r io.Reader) error {
	headerBytes, err := endian.ReadBytes(r)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(headerBytes, &p.Header); err != nil {
		return err
	}

	p.Body, err = endian.ReadBytes(r)
	return err
}

// Encode implements Packet
func (p *LogicPkt) Encode(w io.Writer) error {
	headerBytes, err := proto.Marshal(&p.Header)
	if err != nil {
		return err
	}
	if err := endian.WriteBytes(w, headerBytes); err != nil {
		return err
	}
	return endian.WriteBytes(w, p.Body)
}

var _ Packet = (*LogicPkt)(nil)

func (p *LogicPkt) ReadBody(val proto.Message) error {
	return proto.Unmarshal(p.Body, val)
}

func (p *LogicPkt) WriteBody(val proto.Message) *LogicPkt {
	if val == nil {
		return p
	}
	p.Body, _ = proto.Marshal(val)
	return p
}

func (p *LogicPkt) StringBody() string {
	return string(p.Body)
}

func (p *LogicPkt) String() string {
	return fmt.Sprintf("header: %v, body: %d bits", &p.Header, len(p.Body))
}

func (p *LogicPkt) AddMeta(m ...*Meta) {
	p.Meta = append(p.Meta, m...)
}

func (p *LogicPkt) AddStringMeta(key, value string) {
	p.AddMeta(&Meta{
		Key:   key,
		Value: value,
		Type:  MetaType_string,
	})
}

func (p *LogicPkt) GetMeta(key string) (interface{}, bool) {
	return FindMeta(p.Meta, key)
}

func (p *LogicPkt) DelMeta(key string) {
	for i, m := range p.Meta {
		if m.Key == key {
			length := len(p.Meta)
			if i < length-1 {
				copy(p.Meta[i:], p.Meta[i+1:])
			}
			p.Meta = p.Meta[:length-1]
		}
	}
}

func (h *Header) ServiceName() string {
	arr := strings.SplitN(h.Command, ".", 2)
	if len(arr) <= 1 {
		return "default"
	}
	return arr[0]
}

func FindMeta(meta []*Meta, key string) (interface{}, bool) {
	for _, m := range meta {
		if m.Key == key {
			switch m.Type {
			case MetaType_int:
				v, _ := strconv.Atoi(m.Value)
				return v, true
			case MetaType_float:
				v, _ := strconv.ParseFloat(m.Value, 64)
				return v, true
			}
			return m.Value, true
		}
	}
	return nil, false
}
