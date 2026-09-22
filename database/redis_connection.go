package database

import (
	"context"
	"log/slog"

	"github.com/Fulim13/lottery/util"
	"github.com/redis/go-redis/v9"
)

var (
	GiftRedis *redis.Client
)

func ConnectGiftRedis(confDir, confFile, fileType string) {
	viper := util.InitViper(confDir, confFile, fileType)

	GiftRedis = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("addr"),
		Password: viper.GetString("pass"),
		DB:       viper.GetInt("db"),
	})
	if err := GiftRedis.Ping(context.Background()).Err(); err != nil {
		panic(err) // no Redis at startup means no stock to draw from, so exit instead of failing later
	}
	slog.Info("connect to redis", "addr", viper.GetString("addr"), "db", viper.GetInt("db"))
}

// Close the Redis connection.
func CloseGiftRedis() {
	if GiftRedis != nil {
		GiftRedis.Close()
		slog.Info("close redis")
	}
}
