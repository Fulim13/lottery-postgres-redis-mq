package database

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/Fulim13/lottery/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	GiftDB *gorm.DB
)

func ConnectGiftDB(confDir, confFile, fileType, logDir string) {
	viper := util.InitViper(confDir, confFile, fileType)
	user := viper.GetString("lottery.user")
	pass := viper.GetString("lottery.pass")
	host := viper.GetString("lottery.host")
	port := viper.GetInt("lottery.port")
	dbname := viper.GetString("lottery.dbname")
	sslmode := viper.GetString("lottery.sslmode")
	logFileName := viper.GetString("lottery.log")

	DataSourceName := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, dbname, sslmode)

	// logging
	os.MkdirAll(logDir, os.ModePerm) // OpenFile below fails if the log directory doesn't exist yet
	logFile, err := os.OpenFile(path.Join(logDir, logFileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	newLogger := logger.New(
		log.New(logFile, "\r\n", log.LstdFlags), // an io.Writer, a file here, but os.Stdout works too
		logger.Config{
			SlowThreshold:             100 * time.Millisecond, // anything slower than this counts as a slow query
			LogLevel:                  logger.Warn,            // lowest level that still gets logged; Silent turns logging off
			IgnoreRecordNotFoundError: true,                   // don't log RecordNotFound, it isn't really an error
			Colorful:                  false,                  // no colors in a log file
		},
	)

	db, err := gorm.Open(postgres.Open(DataSourceName), &gorm.Config{
		PrepareStmt:            true,      // cache a prepared statement for every SQL statement, which makes later runs cheaper
		SkipDefaultTransaction: true,      // GORM wraps writes in a transaction for consistency; we don't need that here, and turning it off buys roughly 30% throughput
		Logger:                 newLogger, // logging
	})
	if err != nil {
		panic(err)
	}

	// connection pool settings
	sqlDB, _ := db.DB()
	// upper bound on idle connections kept around (anything beyond this gets closed)
	sqlDB.SetMaxIdleConns(10)
	// hard cap on open connections. Postgres defaults to max_connections=100, so leave room under load tests
	sqlDB.SetMaxOpenConns(50)
	// retire a connection after this long. The database may drop idle connections on its own,
	// so we either ping periodically or set a lifetime here.
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err := sqlDB.Ping(); err != nil {
		panic(fmt.Errorf("failed to connect to Postgres (%s:%d): %w", host, port, err))
	}
	GiftDB = db
	slog.Info("connect to postgres", "host", host, "port", port, "db", dbname)
}

// Ping now and then to keep the connection alive.
func PingGiftDB() {
	if GiftDB != nil {
		sqlDB, _ := GiftDB.DB()
		sqlDB.Ping()
		slog.Info("ping gift db")
	}
}

// Close the database connection.
func CloseGiftDB() {
	if GiftDB != nil {
		sqlDB, _ := GiftDB.DB()
		sqlDB.Close()
		slog.Info("close GiftDB")
	}
}
