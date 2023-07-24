package database

import (
	"fmt"
	"github.com/chang144/gotalk/internal/pkg/snowflake"
	"gorm.io/gorm"
	"testing"
	"time"
)

var db *gorm.DB
var idgen *snowflake.IDGenerator

func init() {
	db, _ = InitMysqlDb("root:123456@tcp(127.0.0.1:3306)/gotalk?charset=utf8mb4&parseTime=True&loc=Local")

	//_ = db.AutoMigrate(&MessageIndex{})
	//_ = db.AutoMigrate(&MessageContent{})

	idgen, _ = snowflake.NewIDGenerator(1)
}

func Test_auto_migrate(t *testing.T) {
	_ = db.AutoMigrate(&MessageContent{})
	_ = db.AutoMigrate(&MessageIndex{})
	_ = db.AutoMigrate(&User{})
	_ = db.AutoMigrate(&Group{})
	_ = db.AutoMigrate(&GroupMember{})
}

func Benchmark_insert(b *testing.B) {
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
