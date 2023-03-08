package qim

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

// websocket implementation of Channel
type ChannelImpl struct {
	id string
	Conn
	meta      Meta
	writechan chan []byte
	writewait time.Duration
	readwait  time.Duration
	gpool     *ants.Pool
	state     int32 // 0 init 1 started 2 closed
	lg        *zap.Logger
}

var _ Channel = (*ChannelImpl)(nil)

func (ch *ChannelImpl) Close() error {
	if !atomic.CompareAndSwapInt32(&ch.state, 1, 2) {
		return fmt.Errorf("channel state not started")
	}
	close(ch.writechan)
	return nil
}

// GetMeta implements Channel
func (ch *ChannelImpl) GetMeta() Meta {
	return ch.meta
}

// ID implements Channel
func (ch *ChannelImpl) ID() string {
	return ch.id
}

// Push implements Channel
// 异步写入消息
func (ch *ChannelImpl) Push(payload []byte) error {
	if atomic.LoadInt32(&ch.state) != 1 {
		return fmt.Errorf("channel %s has closed", ch.id)
	}

	ch.writechan <- payload
	return nil
}

// Readloop implements Channel
// 负责读取消息与心跳处理(SetReadDeadline)
// 利用atomic保证只被一个线程调用
func (ch *ChannelImpl) Readloop(lst MessageListener) error {
	if !atomic.CompareAndSwapInt32(&ch.state, 0, 1) {
		return fmt.Errorf("channel has started")
	}
	log := ch.lg.With(zap.String("func", "Readloop"))

	for {
		_ = ch.SetReadDeadline(time.Now().Add(ch.readwait))

		frame, err := ch.ReadFrame()
		if err != nil {
			log.Warn("ReadFrame error", zap.Error(err))
			return err
		}
		if frame.GetOpCode() == OpClose {
			return errors.New("remote side close the channel")
		}
		if frame.GetOpCode() == OpPing {
			log.Debug("receive a ping, response with a pong")
			_ = ch.WriteFrame(OpPong, nil)
			_ = ch.Flush()
			continue
		}

		payload := frame.GetPayload()
		if len(payload) == 0 {
			continue
		}
		err = ch.gpool.Submit(func() {
			lst.Receive(ch, payload)
		})
		if err != nil {
			return err
		}
	}
}

// SetReadwait implements Channel
func (ch *ChannelImpl) SetReadwait(timeout time.Duration) {
	if timeout == 0 {
		return
	}
	ch.readwait = timeout
}

// SetWritewait implements Channel
func (ch *ChannelImpl) SetWritewait(timeout time.Duration) {
	if timeout == 0 {
		return
	}
	ch.writewait = timeout
}

func NewChannel(id string, meta Meta, conn Conn, gpool *ants.Pool, logger *zap.Logger) Channel {
	logger = logger.With(zap.String("module", "ChannelImpl"), zap.String("id", id))

	ch := &ChannelImpl{
		id:        id,
		Conn:      conn,
		meta:      meta,
		writechan: make(chan []byte, 5),
		writewait: DefaultWritewait,
		readwait:  DefaultReadwait,
		gpool:     gpool,
		state:     0,
		lg:        logger,
	}

	go func() {
		err := ch.writeloop()
		if err != nil {
			logger.Info(err.Error())
		}
	}()

	return ch
}

func (ch *ChannelImpl) writeloop() error {
	defer func() {
		ch.lg.Debug("channel writeloop exited")
	}()

	for payload := range ch.writechan {
		err := ch.WriteFrame(OpBinary, payload)
		if err != nil {
			return err
		}
		chanlen := len(ch.writechan)
		for i := 0; i < chanlen; i++ {
			payload = <-ch.writechan
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}
		}
		err = ch.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}
