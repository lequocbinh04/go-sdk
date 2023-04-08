/*
 * @author           Viet Tran <viettranx@gmail.com>
 * @copyright       2019 Viet Tran <viettranx@gmail.com>
 * @license           Apache-2.0
 */

package sdkgorm

import (
	"errors"
	"flag"
	"fmt"
	"github.com/lequocbinh04/go-sdk/logger"
	"github.com/lequocbinh04/go-sdk/plugin/storage/sdkgorm/gormdialects"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"strings"
	"sync"
	"time"
)

type GormDBType int

const (
	GormDBTypeMySQL GormDBType = iota + 1
	GormDBTypePostgres
	GormDBTypeSQLite
	GormDBTypeMSSQL
	GormDBTypeNotSupported
)

const retryCount = 10

type GormOpt struct {
	Uri                   string
	Prefix                string
	DBType                string
	PingInterval          int // in seconds
	MaxOpenConnections    int
	MaxIdleConnections    int
	MaxConnectionIdleTime int
}

type gormDB struct {
	name      string
	logger    logger.Logger
	db        *gorm.DB
	isRunning bool
	once      *sync.Once
	*GormOpt
}

func NewGormDB(name, prefix string) *gormDB {
	return &gormDB{
		GormOpt: &GormOpt{
			Prefix: prefix,
		},
		name:      name,
		isRunning: false,
		once:      new(sync.Once),
	}
}

func (gdb *gormDB) GetPrefix() string {
	return gdb.Prefix
}

func (gdb *gormDB) Name() string {
	return gdb.name
}

func (gdb *gormDB) InitFlags() {
	prefix := gdb.Prefix
	if gdb.Prefix != "" {
		prefix += "-"
	}

	flag.StringVar(&gdb.Uri, prefix+"gorm-db-uri", "", "Gorm database connection-string.")
	flag.StringVar(&gdb.DBType, prefix+"gorm-db-type", "mysql", "Gorm database type (mysql, postgres, sqlite, mssql)")
	flag.IntVar(&gdb.PingInterval, prefix+"gorm-db-ping-interval", 5, "Gorm database ping check interval")
	flag.IntVar(
		&gdb.MaxOpenConnections,
		fmt.Sprintf("%sdb-max-conn", prefix),
		50,
		"maximum number of open connections to the database - Default 50",
	)

	flag.IntVar(
		&gdb.MaxIdleConnections,
		fmt.Sprintf("%sdb-max-ide-conn", prefix),
		15,
		"maximum number of database connections in the idle - Default 10",
	)

	flag.IntVar(
		&gdb.MaxConnectionIdleTime,
		fmt.Sprintf("%sdb-max-conn-ide-time", prefix),
		3600,
		"maximum amount of time a connection may be idle in seconds - Default 3600",
	)
}

func (gdb *gormDB) isDisabled() bool {
	return gdb.Uri == ""
}

func (gdb *gormDB) Configure() error {
	if gdb.isDisabled() || gdb.isRunning {
		return nil
	}

	gdb.logger = logger.GetCurrent().GetLogger(gdb.name)

	dbType := getDBType(gdb.DBType)
	if dbType == GormDBTypeNotSupported {
		return errors.New("gorm database type is not supported")
	}

	gdb.logger.Info("Connect to Gorm DB at ", gdb.Uri, " ...")

	var err error
	gdb.db, err = gdb.getDBConn(dbType)
	if err != nil {
		gdb.logger.Error("Error connect to gorm database at ", gdb.Uri, ". ", err.Error())
		return err
	}
	gdb.isRunning = true

	return nil
}

func (gdb *gormDB) Run() error {
	return gdb.Configure()
}

func (gdb *gormDB) Stop() <-chan bool {
	gdb.isRunning = false

	c := make(chan bool)
	go func() {
		c <- true
		gdb.logger.Infoln("Stopped")
	}()
	return c
}

func (gdb *gormDB) Get() interface{} {
	if gdb.logger.GetLevel() == "debug" || gdb.logger.GetLevel() == "trace" {
		return gdb.db.Session(&gorm.Session{NewDB: true}).Debug()
	}
	newSessionDB := gdb.db.Session(&gorm.Session{NewDB: true, Logger: gdb.db.Logger.LogMode(glogger.Silent)})
	if db, err := newSessionDB.DB(); err == nil {
		db.SetMaxOpenConns(gdb.MaxOpenConnections)
		db.SetMaxIdleConns(gdb.MaxIdleConnections)
		db.SetConnMaxIdleTime(time.Second * time.Duration(gdb.MaxConnectionIdleTime))
	}
	return newSessionDB
}

func getDBType(dbType string) GormDBType {
	switch strings.ToLower(dbType) {
	case "mysql":
		return GormDBTypeMySQL
	case "postgres":
		return GormDBTypePostgres
	case "sqlite":
		return GormDBTypeSQLite
	case "mssql":
		return GormDBTypeMSSQL
	}

	return GormDBTypeNotSupported
}

func (gdb *gormDB) getDBConn(t GormDBType) (dbConn *gorm.DB, err error) {
	switch t {
	case GormDBTypeMySQL:
		return gormdialects.MySqlDB(gdb.Uri)
	case GormDBTypePostgres:
		return gormdialects.PostgresDB(gdb.Uri)
	case GormDBTypeSQLite:
		return gormdialects.SQLiteDB(gdb.Uri)
	case GormDBTypeMSSQL:
		return gormdialects.MSSqlDB(gdb.Uri)
	}

	return nil, nil
}
