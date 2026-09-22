package database

import (
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

const EMPTY_GIFT = 1 // id of the blank prize ("Thanks for playing")

type Gift struct {
	Id      int `gorm:"primaryKey"`
	Name    string
	Price   int
	Picture string // where the image lives
	Count   int    // units in stock
}

func (Gift) TableName() string {
	return "inventory"
}

// Read the whole inventory table. Fine to select everything while the table stays small.
func GetAllGifts() []*Gift {
	var gifts []*Gift
	err := GiftDB.Order("id").Find(&gifts).Error
	if err != nil {
		slog.Error("scan table inventory failed", "error", err)
	}
	return gifts
}

func GetGift(id int) *Gift {
	var gift Gift
	err := GiftDB.First(&gift, id).Error // look up by primary key; returns ErrRecordNotFound when there is no such row
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gift by id failed", "error", err, "gid", id)
		}
		return nil
	}
	return &gift
}
