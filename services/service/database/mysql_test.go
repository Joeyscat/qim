package database

import (
	"fmt"
	"log"
	"testing"
	"time"

	"gorm.io/gorm"
)

var db *gorm.DB
var idgen *IDGenerator

func init() {
	var err error
	db, err = InitDB("sqlite", "msg.db")
	if err != nil {
		log.Fatalln(err)
	}

	err = db.AutoMigrate(&MessageIndex{}, &MessageContent{})
	if err != nil {
		log.Fatalln(err)
	}

	idgen, err = NewIDGenerator(1)
	if err != nil {
		log.Fatalln(err)
	}
}

// Benchmark_Insert-8           144          10588232 ns/op           0.10 MB/s      293359 B/op       3373 allocs/op
// PASS
// ok      github.com/joeyscat/qim/services/service/database       2.081s
func Benchmark_Insert(b *testing.B) {
	sendTime := time.Now().UnixNano()
	b.ResetTimer()
	b.SetBytes(1024)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idxs := make([]MessageIndex, 100)
			cid := idgen.Next().Int64()
			for i := 0; i < len(idxs); i++ {
				idxs[i] = MessageIndex{
					ID:        idgen.Next().Int64(),
					AccountA:  fmt.Sprintf("test_%d", cid),
					AccountB:  fmt.Sprintf("test_%d", i),
					SendTime:  sendTime,
					MessageID: cid,
				}
			}
			db.Create(&idxs)
		}
	})
}
