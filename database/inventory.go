package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
)

/**
Live stock levels are kept in Redis.
When a lot of users join the draw at once we cannot decrement stock in Postgres for every
request: the database won't take that kind of concurrency, while Redis will.
*/

const (
	INVENTORY_PREFIX = "gift_count_" // shared prefix on every key, which makes it easy to scan them later
)

// Copy the starting stock of every prize out of Postgres and into Redis.
func InitGiftInventory() {
	for _, gift := range GetAllGifts() {
		if gift.Count <= 0 {
			slog.Warn("gift count is zero", "id", gift.Id, "name", gift.Name)
			continue // a prize with no stock doesn't take part in the draw
		}
		err := GiftRedis.Set(context.Background(), INVENTORY_PREFIX+strconv.Itoa(gift.Id), gift.Count, 0).Err()
		if err != nil {
			slog.Error("set gift count to redis failed", "gift id", gift.Id, "error", err)
		}
	}
}

// Remaining stock for every prize.
func GetAllGiftInventory() []*Gift {
	keys, err := GiftRedis.Keys(context.Background(), INVENTORY_PREFIX+"*").Result() // every prize key, found by prefix
	if err != nil {
		slog.Error("iterate all gift keys failed", "error", err)
		return nil
	}
	gifts := make([]*Gift, 0, len(keys))
	for _, key := range keys { // read the stock count behind each prize key
		if id, err := strconv.Atoi(key[len(INVENTORY_PREFIX):]); err == nil {
			count, err := GiftRedis.Get(context.Background(), key).Int()
			if err == nil {
				gifts = append(gifts, &Gift{Id: id, Count: count})
			} else {
				slog.Error("gift count is not int", "key", key)
			}
		} else {
			slog.Error("gift id is not int", "key", key)
		}
	}

	return gifts
}

// Remaining stock for one prize.
func GetGiftInventory(GiftId int) int {
	key := INVENTORY_PREFIX + strconv.Itoa(GiftId)
	count, err := GiftRedis.Get(context.Background(), key).Int()
	if err == nil {
		return count
	} else {
		slog.Error("gift count is not int", "key", key)
		return -1
	}
}

// Take one unit of a prize out of stock.
func ReduceInventory(GiftId int) error {
	key := INVENTORY_PREFIX + strconv.Itoa(GiftId)
	n, err := GiftRedis.Decr(context.Background(), key).Result() // atomic; returns the value after decrementing, or -1 when the key is missing
	if err != nil {
		slog.Error("decr key failed", "key", key, "error", err)
		return err
	} else {
		if n < 0 {
			msg := fmt.Sprintf("gift %d is out of stock, decrement failed", GiftId)
			slog.Error(msg)
			return errors.New(msg)
		}
		return nil
	}
}

// Put one unit of a prize back into stock.
func IncreaseInventory(GiftId int) error {
	key := INVENTORY_PREFIX + strconv.Itoa(GiftId)
	_, err := GiftRedis.Incr(context.Background(), key).Result() // atomic increment
	if err != nil {
		slog.Error("incr key failed", "key", key, "error", err)
		return err
	} else {
		return nil
	}
}
