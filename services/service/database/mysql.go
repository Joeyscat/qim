package database

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB(driver string, dsn string) (db *gorm.DB, err error) {

	defaultLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Colorful:      true,
	})

	var dialector gorm.Dialector
	if driver == "mysql" {
		dialector = mysql.Open(dsn)
	} else if driver == "sqlite" {
		dialector = sqlite.Open(dsn)
	} else {
		return nil, errors.New("driver not support")
	}

	db, err = gorm.Open(dialector, &gorm.Config{
		Logger: defaultLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
			NameReplacer:  strings.NewReplacer("CID", "Cid"),
		},
	})

	return
}
