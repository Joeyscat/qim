package handler

import (
	"fmt"
	"testing"
	"time"

	"github.com/joeyscat/qim/services/service/database"
	"github.com/joeyscat/qim/wire/rpcc"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

var handler ServiceHandler

func init() {
	baseDB, err := database.InitDB("mysql", "root:123456@tcp(127.0.0.1:3306)/qim_base?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	messageDB, err := database.InitDB("mysql", "root:123456@tcp(127.0.0.1:3306)/qim_message?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	err = baseDB.AutoMigrate(&database.Group{}, &database.GroupMember{})
	if err != nil {
		panic(err)
	}
	err = messageDB.AutoMigrate(&database.MessageIndex{}, &database.MessageContent{})
	if err != nil {
		panic(err)
	}

	idgen, err := database.NewIDGenerator(1)
	if err != nil {
		panic(err)
	}
	handler = ServiceHandler{
		BaseDB:    baseDB,
		MessageDB: messageDB,
		IDgen:     idgen,
	}
}

// Benchmark_InsertUserMessage-8               2367            713581 ns/op           1.44 MB/s       19211 B/op        198 allocs/op
// PASS
// ok      github.com/joeyscat/qim/services/service/handler        1.795s
func Benchmark_InsertUserMessage(b *testing.B) {
	b.ResetTimer()
	b.SetBytes(1024)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = handler.insertUserMessage(&rpcc.InsertMessageReq{
				Sender:   "u1",
				Dest:     ksuid.New().String(),
				SendTime: time.Now().UnixNano(),
				Message:  &rpcc.Message{Type: 1, Body: "hello"},
			})
		}
	})
}

// Benchmark_InsertGroup10Message-8             894           1870230 ns/op           0.55 MB/s       52106 B/op        583 allocs/op
// PASS
// ok      github.com/joeyscat/qim/services/service/handler        1.871s
func Benchmark_InsertGroup10Message(b *testing.B) {
	memberCount := 10
	members := make([]string, memberCount)
	for i := 0; i < memberCount; i++ {
		members[i] = fmt.Sprintf("u_%d", i+1)
	}

	groupID, err := handler.groupCreate(&rpcc.CreateGroupReq{
		App:     "app1",
		Owner:   "u1",
		Name:    "test",
		Members: members,
	})
	assert.NoError(b, err)

	b.ResetTimer()
	b.SetBytes(1024)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = handler.insertGroupMessage(&rpcc.InsertMessageReq{
				Sender:   "u1",
				Dest:     groupID.Base36(),
				SendTime: time.Now().UnixNano(),
				Message:  &rpcc.Message{Type: 1, Body: "hello"},
			})
		}
	})
}

// Benchmark_InsertGroup50Message-8             260           5839543 ns/op           0.18 MB/s      176762 B/op       1876 allocs/op
// PASS
// ok      github.com/joeyscat/qim/services/service/handler        2.056s
func Benchmark_InsertGroup50Message(b *testing.B) {
	memberCount := 50
	members := make([]string, memberCount)
	for i := 0; i < memberCount; i++ {
		members[i] = fmt.Sprintf("u_%d", i+1)
	}

	groupID, err := handler.groupCreate(&rpcc.CreateGroupReq{
		App:     "app1",
		Owner:   "u1",
		Name:    "test",
		Members: members,
	})
	assert.NoError(b, err)

	b.ResetTimer()
	b.SetBytes(1024)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = handler.insertGroupMessage(&rpcc.InsertMessageReq{
				Sender:   "u1",
				Dest:     groupID.Base36(),
				SendTime: time.Now().UnixNano(),
				Message:  &rpcc.Message{Type: 1, Body: "hello"},
			})
		}
	})
}

func TestServiceHandler_insertUserMessage(t *testing.T) {
	type fields struct {
		handler *ServiceHandler
	}
	type args struct {
		req *rpcc.InsertMessageReq
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"insert user message",
			fields{&handler},
			args{&rpcc.InsertMessageReq{
				Sender: "u1", Dest: "u2", SendTime: time.Now().UnixNano(), Message: &rpcc.Message{Type: 1, Body: "hello"},
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.fields.handler
			got, err := h.insertUserMessage(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceHandler.insertUserMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("got:", got)
			if got == 0 {
				t.Errorf("ServiceHandler.insertUserMessage() = %v", got)
			}
		})
	}
}
