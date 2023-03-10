package pkt

import (
	"bytes"
	"testing"

	"github.com/joeyscat/qim/wire"
	"github.com/stretchr/testify/assert"
)

func TestProtoEncode(t *testing.T) {
	arr := []byte{195, 17, 163, 101, 0, 0, 0, 16, 10, 12, 108, 111, 103, 105, 110, 46, 115, 105, 103, 110, 105, 110, 24, 1, 0, 0, 0, 148, 10, 140, 1, 101, 121, 74, 104, 98, 71, 99, 105, 79, 105, 74, 73, 85, 122, 73, 49, 78, 105, 73, 115, 73, 110, 82, 53, 99, 67, 73, 54, 73, 107, 112, 88, 86, 67, 74, 57, 46, 101, 121, 74, 104, 89, 50, 77, 105, 79, 105, 74, 48, 90, 88, 78, 48, 77, 83, 73, 115, 73, 109, 70, 119, 99, 67, 73, 54, 73, 109, 116, 112, 98, 83, 73, 115, 73, 109, 86, 52, 99, 67, 73, 54, 77, 84, 89, 121, 79, 84, 65, 53, 77, 122, 85, 48, 79, 88, 48, 46, 80, 95, 121, 107, 49, 75, 77, 66, 53, 118, 57, 114, 105, 85, 121, 48, 121, 87, 52, 101, 79, 84, 103, 67, 48, 107, 48, 113, 101, 66, 54, 88, 82, 106, 105, 104, 52, 100, 76, 49, 120, 71, 107, 34, 3, 119, 101, 98}
	buf := bytes.NewBuffer(arr)
	got, err := Read(buf)
	assert.Nil(t, err)
	t.Log(got)

	p, ok := got.(*LogicPkt)
	assert.True(t, ok)
	assert.Equal(t, wire.CommandLoginSignIn, p.Command)
	t.Log(p.Body)

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2MiOiJ0ZXN0MSIsImFwcCI6ImtpbSIsImV4cCI6MTYyOTA5MzU0OX0.P_yk1KMB5v9riUy0yW4eOTgC0k0qeB6XRjih4dL1xGk"
	var req LoginReq
	err = p.ReadBody(&req)
	t.Log(&req)
	assert.Nil(t, err)
	assert.Equal(t, token, req.Token)
}

func TestReadPkt(t *testing.T) {
	seq := wire.Seq.Next()

	packet := New("auth.login.aa", WithSeq(seq), WithStatus(Status_Success))
	assert.Equal(t, "auth", packet.ServiceName())

	packet = New(wire.CommandLoginSignIn, WithSeq(seq), WithStatus(Status_Success))
	packet.WriteBody(&LoginReq{
		Token: "test token",
	})
	packet.AddMeta(
		&Meta{Key: "test", Value: "test"},
		&Meta{Key: wire.MetaDestServer, Value: "test"},
		&Meta{Key: wire.MetaDestChannels, Value: "test1,test2"},
	)

	buf := bytes.NewBuffer(Marshal(packet))
	t.Log(buf.Bytes())

	got, err := Read(buf)
	assert.Nil(t, err)
	p, ok := got.(*LogicPkt)
	assert.True(t, ok)
	assert.Equal(t, wire.CommandLoginSignIn, p.Command)
	assert.Equal(t, seq, p.Sequence)
	assert.Equal(t, Status_Success, p.Status)

	assert.Equal(t, 3, len(packet.Meta))

	packet.DelMeta(wire.MetaDestServer)
	assert.Equal(t, 2, len(packet.Meta))
	assert.Equal(t, wire.MetaDestChannels, packet.Meta[1].Key)

	packet.DelMeta(wire.MetaDestChannels)
	assert.Equal(t, 1, len(packet.Meta))

}
