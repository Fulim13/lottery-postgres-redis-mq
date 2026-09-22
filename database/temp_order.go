package database

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
)

/*
Temporary orders: a user who wins a prize holds it here until they pay for it.
*/
const (
	TEMP_ORDER_PREFIX = "porder_"
)

// Create a temporary order.
func CreateTempOrder(uid int, GiftId int) error {
	key := TEMP_ORDER_PREFIX + strconv.Itoa(uid)
	// store the temporary order in redis, keyed by uid, with the giftId as the value
	if err := GiftRedis.Set(context.Background(), key, GiftId, 0).Err(); err != nil {
		return err
	}
	return nil
}

// Look up a temporary order and return the gift id.
func GetTempOrder(uid int) int {
	key := TEMP_ORDER_PREFIX + strconv.Itoa(uid)
	giftId, err := GiftRedis.Get(context.Background(), key).Int()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			slog.Error("query redis fail", "key", key, "error", err)
		}
		return 0
	} else {
		return giftId
	}
}

// Delete a temporary order and return how many were deleted.
func DeleteTempOrder(uid int, GiftId int) int64 {
	key := TEMP_ORDER_PREFIX + strconv.Itoa(uid)
	if n, err := GiftRedis.Del(context.Background(), key).Result(); err != nil { // n is the number of keys deleted
		slog.Error("delete TempOrder failed", "error", err)
		return -1
	} else {
		return n
	}
}
