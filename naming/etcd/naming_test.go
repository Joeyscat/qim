package etcd

import (
	"strings"
	"testing"
	"time"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/naming"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

func Test_etcdNaming_Register(t *testing.T) {
	log, err := zap.NewDevelopment()
	assert.Nil(t, err)
	n, err := NewNaming([]string{"127.0.0.1:2379"}, log)
	assert.Nil(t, err)

	s1 := naming.NewEntry("s1", "hello", "tcp", "127.0.0.1", 8001)
	s2 := naming.NewEntry("s2", "hello", "tcp", "127.0.0.1", 8002)
	s3 := naming.NewEntry("s3", "hello", "tcp", "127.0.0.1", 8003)
	s1Dup := naming.NewEntry("s1", "hello", "tcp", "127.0.0.1", 8001)

	ss, err := n.Find(s1.ServiceName())
	assert.Nil(t, err)
	assert.Empty(t, ss)

	err = n.Register(s1)
	assert.Nil(t, err)

	ss, err = n.Find(s1.ServiceName())
	assert.Nil(t, err)
	assert.NotEmpty(t, ss)

	err = n.Register(s1Dup)
	assert.NotNil(t, err)
	t.Log(err)
	ss, err = n.Find(s1.ServiceName())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ss))

	err = n.Register(s2)
	assert.Nil(t, err)
	ss, err = n.Find(s1.ServiceName())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ss))

	err = n.Register(s3)
	assert.Nil(t, err)
	ss, err = n.Find(s1.ServiceName())
	assert.Nil(t, err)
	assert.Equal(t, 3, len(ss))

	err = n.Deregister(s3.ServiceID())
	assert.Nil(t, err)
	ss, err = n.Find(s1.ServiceName())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ss))
}

func Test_etcdNaming_Subscribe(t *testing.T) {
	log, err := zap.NewDevelopment()
	assert.Nil(t, err)
	n, err := NewNaming([]string{"127.0.0.1:2379"}, log)
	assert.Nil(t, err)

	s1 := naming.NewEntry("s1", "hello", "tcp", "127.0.0.1", 8001)
	s2 := naming.NewEntry("s2", "hello", "tcp", "127.0.0.1", 8002)
	s3 := naming.NewEntry("s3", "hello", "tcp", "127.0.0.1", 8002)
	s1Dup := naming.NewEntry("s1", "hello", "tcp", "127.0.0.1", 8001)

	subscribedSvc := []string{}

	n.Subscribe(s1.ServiceName(), func(services []qim.ServiceRegistration) {
		ids := []string{}
		for _, v := range services {
			ids = append(ids, v.ServiceID())
		}
		t.Logf("Service(%s) Change", s1.ServiceName())
		slices.Sort(ids)
		t.Logf("[%s]", strings.Join(ids, ","))
		subscribedSvc = ids
	})
	assert.Equal(t, []string{}, subscribedSvc)

	err = n.Register(s1)
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, []string{"s1"}, subscribedSvc)

	err = n.Register(s1Dup)
	assert.NotNil(t, err)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, []string{"s1"}, subscribedSvc)

	err = n.Register(s2)
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, []string{"s1", "s2"}, subscribedSvc)

	err = n.Register(s3)
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, []string{"s1", "s2", "s3"}, subscribedSvc)

	err = n.Deregister(s1.ServiceID())
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, []string{"s2", "s3"}, subscribedSvc)

	err = n.Unsubscribe(s1.ServiceName())
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, []string{"s2", "s3"}, subscribedSvc)

	err = n.Deregister(s2.ServiceID())
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 100)
	err = n.Deregister(s2.ServiceID())
	assert.NotNil(t, err)

	time.Sleep(time.Millisecond * 300)
}

func Test_etcdNaming_Find(t *testing.T) {
	log, err := zap.NewDevelopment()
	assert.Nil(t, err)
	n, err := NewNaming([]string{"127.0.0.1:2379"}, log)
	assert.Nil(t, err)

	s1 := naming.NewEntry("s1", "hello", "tcp", "127.0.0.1", 8001)
	s2 := &naming.DefaultService{
		ID:        "s2",
		Name:      "hello",
		Namespace: "",
		Address:   "localhost",
		Port:      8001,
		Protocol:  "ws",
		Tags:      []string{"tab1", "gate"},
	}
	s3 := &naming.DefaultService{
		ID:        "s3",
		Name:      "hello",
		Namespace: "",
		Address:   "localhost",
		Port:      8001,
		Protocol:  "ws",
		Tags:      []string{"tab2", "gate"},
	}

	ss, err := n.Find(s1.ServiceName())
	assert.Nil(t, err)
	assert.Empty(t, ss)

	err = n.Register(s1)
	assert.Nil(t, err)
	err = n.Register(s2)
	assert.Nil(t, err)
	err = n.Register(s3)
	assert.Nil(t, err)

	ss, err = n.Find(s1.ServiceName())
	assert.Nil(t, err)
	assert.Equal(t, 3, len(ss))

	ss, err = n.Find(s1.ServiceName(), "gate")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ss))

	ss, err = n.Find(s1.ServiceName(), "tab2")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ss))
}
