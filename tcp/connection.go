package tcp

import (
	"bufio"
	"io"
	"net"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/wire/endian"
)

type Frame struct {
	OpCode  qim.OpCode
	Payload []byte
}

// GetOpCode implements qim.Frame
func (f *Frame) GetOpCode() qim.OpCode {
	return qim.OpCode(f.OpCode)
}

// GetPayload implements qim.Frame
func (f *Frame) GetPayload() []byte {
	return f.Payload
}

// SetOpCode implements qim.Frame
func (f *Frame) SetOpCode(opcode qim.OpCode) {
	f.OpCode = opcode
}

// SetPayload implements qim.Frame
func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

var _ qim.Frame = (*Frame)(nil)

type TcpConn struct {
	net.Conn
	rd *bufio.Reader
	wr *bufio.Writer
}

func NewConn(conn net.Conn) qim.Conn {
	return &TcpConn{
		Conn: conn,
		rd:   bufio.NewReaderSize(conn, 4096),
		wr:   bufio.NewWriterSize(conn, 1024),
	}
}

func NewConnWithRW(conn net.Conn, rd *bufio.Reader, wr *bufio.Writer) *TcpConn {
	return &TcpConn{
		Conn: conn,
		rd:   rd,
		wr:   wr,
	}
}

// Flush implements qim.Conn
func (c *TcpConn) Flush() error {
	return c.wr.Flush()
}

// ReadFrame implements qim.Conn
func (c *TcpConn) ReadFrame() (qim.Frame, error) {
	opcode, err := endian.ReadUint8(c.rd)
	if err != nil {
		return nil, err
	}
	payload, err := endian.ReadBytes(c.rd)
	if err != nil {
		return nil, err
	}

	return &Frame{
		OpCode:  qim.OpCode(opcode),
		Payload: payload,
	}, nil
}

// WriteFrame implements qim.Conn
func (c *TcpConn) WriteFrame(opcode qim.OpCode, payload []byte) error {
	return WriteFrame(c.wr, opcode, payload)
}

var _ qim.Conn = (*TcpConn)(nil)

func WriteFrame(w io.Writer, code qim.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	return endian.WriteBytes(w, payload)
}
