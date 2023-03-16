package serv

import (
	"testing"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/naming"
	"github.com/joeyscat/qim/wire"
	"github.com/joeyscat/qim/wire/pkt"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRouteSelector_Lookup(t *testing.T) {
	srvs := []qim.Service{
		&naming.DefaultService{ID: "s1", Meta: map[string]string{"zone": "zone_01"}},
		&naming.DefaultService{ID: "s2", Meta: map[string]string{"zone": "zone_01"}},
		&naming.DefaultService{ID: "s3", Meta: map[string]string{"zone": "zone_01"}},
		&naming.DefaultService{ID: "s4", Meta: map[string]string{"zone": "zone_02"}},
		&naming.DefaultService{ID: "s5", Meta: map[string]string{"zone": "zone_03"}},
		&naming.DefaultService{ID: "s6", Meta: map[string]string{"zone": "zone_03"}},
	}

	log, err := zap.NewDevelopment()
	assert.Nil(t, err)
	rs, err := NewRouteSelector("../route.json", log)
	assert.Nil(t, err)

	packet := pkt.New(wire.CommandChatUserTalk, pkt.WithChannel(ksuid.New().String()))
	packet.AddStringMeta(MetaKeyApp, "qim")
	packet.AddStringMeta(MetaKeyAccount, "test1")
	hit := rs.Lookup(&packet.Header, srvs)
	assert.Equal(t, "s4", hit)

	hits := make(map[string]int)
	for i := 0; i < 100; i++ {
		header := pkt.Header{
			ChannelId: ksuid.New().String(),
			Meta: []*pkt.Meta{
				{Type: pkt.MetaType_string, Key: MetaKeyApp, Value: ksuid.New().String()},
				{Type: pkt.MetaType_string, Key: MetaKeyAccount, Value: ksuid.New().String()},
			},
		}
		hit := rs.Lookup(&header, srvs)
		hits[hit]++
	}

	t.Log(hits)
}
