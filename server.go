package qim

import (
	"context"
	"net"
	"time"
)

const (
	DefaultReadwait  = time.Minute * 3
	DefaultWritewait = time.Second * 10
	DefaultLoginwait = time.Second * 10
	DefaultHeartbeat = time.Second * 55
)

const (
	DefaultMessageReadPool = 5000
	DefaultConnectionPool  = 5000
)

// 基础服务的抽象接口
type Service interface {
	ServiceID() string
	ServiceName() string
	GetMeta() map[string]string
}

// 服务注册的抽象接口
type ServiceRegistration interface {
	Service
	PublicAddress() string
	PublicPort() uint16
	DialURL() string
	GetTags() []string
	GetProtocol() string
	GetNamespace() string
	String() string
}

// 适用于不同底层协议（tcp、websocket）的通用服务端接口
type Server interface {
	ServiceRegistration

	SetAcceptor(acceptor Acceptor)

	SetMessageListener(messageListener MessageListener)

	SetStateListener(stateListener StateListener)

	SetReadwait(readwait time.Duration)

	SetChannelMap(channelMap ChannelMap)

	// Start 用于在内部实现网络端口的监听和接收连接，
	// 并完成一个Channel的初始化过程。
	Start() error
	// Push消息到指定的Channel中，
	Push(channelID string, payload []byte) error
	// 服务下线，关闭连接
	Shutdown(ctx context.Context) error
}

// 连接接收器
type Acceptor interface {
	// 返回有个握手完成的Channel对象或者一个error。
	// 业务层需要处理不同协议和网络环境下的连接握手协议。
	Accept(conn Conn, timeout time.Duration) (string, Meta, error)
}

// 监听消息
type MessageListener interface {
	// 收到消息回调
	Receive(agent Agent, payload []byte)
}

// 状态监听器
type StateListener interface {
	// 连接断开回调
	Disconnect(channelID string) error
}

type Meta map[string]string

type Agent interface {
	ID() string
	Push(payload []byte) error
	GetMeta() Meta
}

type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(opcode OpCode, payload []byte) error
	Flush() error
}

type Channel interface {
	Conn
	Agent
	Close() error
	Readloop(lst MessageListener) error
	SetWritewait(timeout time.Duration)
	SetReadwait(timeout time.Duration)
}

type Client interface {
	Service
	Connect(addr string) error
	SetDialer(dialer Dialer)
	Send(payload []byte) error
	Read() (Frame, error)
	Close()
}

type Dialer interface {
	DialAndHandshake(ctx DialerContext) (net.Conn, error)
}

type DialerContext struct {
	ID      string
	Name    string
	Address string
	Timeout time.Duration
}

type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

type Frame interface {
	SetOpCode(opcode OpCode)
	GetOpCode() OpCode
	SetPayload(payload []byte)
	GetPayload() []byte
}
