package websocket

import (
	"bufio"
	"net"

	"github.com/gobwas/ws"
	"github.com/joeyscat/qim"
)

type Frame struct {
	raw ws.Frame
}

// GetOpCode implements qim.Frame
func (f *Frame) GetOpCode() qim.OpCode {
	return qim.OpCode(f.raw.Header.OpCode)
}

// GetPayload implements qim.Frame
func (f *Frame) GetPayload() []byte {
	if f.raw.Header.Masked {
		ws.Cipher(f.raw.Payload, f.raw.Header.Mask, 0)
	}
	f.raw.Header.Masked = false
	return f.raw.Payload
}

// SetOpCode implements qim.Frame
func (f *Frame) SetOpCode(opcode qim.OpCode) {
	f.raw.Header.OpCode = ws.OpCode(opcode)
}

// SetPayload implements qim.Frame
func (f *Frame) SetPayload(payload []byte) {
	f.raw.Payload = payload
}

var _ qim.Frame = (*Frame)(nil)

type WsConn struct {
	net.Conn
	rd *bufio.Reader
	wr *bufio.Writer
}

func NewConn(conn net.Conn) qim.Conn {
	return &WsConn{
		Conn: conn,
		rd:   bufio.NewReaderSize(conn, 4096),
		wr:   bufio.NewWriterSize(conn, 1024),
	}
}

func NewConnWithRW(conn net.Conn, rd *bufio.Reader, wr *bufio.Writer) *WsConn {
	return &WsConn{
		Conn: conn,
		rd:   rd,
		wr:   wr,
	}
}

// Flush implements qim.Conn
func (c *WsConn) Flush() error {
	return c.wr.Flush()
}

// ReadFrame implements qim.Conn
func (c *WsConn) ReadFrame() (qim.Frame, error) {
	f, err := ws.ReadFrame(c.rd)
	if err != nil {
		return nil, err
	}
	return &Frame{raw: f}, nil
}

// WriteFrame implements qim.Conn
func (c *WsConn) WriteFrame(opcode qim.OpCode, payload []byte) error {
	f := ws.NewFrame(ws.OpCode(opcode), true, payload)
	return ws.WriteFrame(c.wr, f)
}

var _ qim.Conn = (*WsConn)(nil)
