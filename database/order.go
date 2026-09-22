package database

import "log/slog"

type Order struct {
	Id     int `gorm:"primaryKey"`
	GiftId int
	UserId int
}

// Write one order row and return its id. On Postgres the generated primary key comes back
// through INSERT ... RETURNING id, which GORM fills into order.Id for us.
func CreateOrder(userid, giftid int) int {
	order := Order{GiftId: giftid, UserId: userid}
	if err := GiftDB.Create(&order).Error; err != nil {
		slog.Error("create order failed", "error", err, "userid", userid, "giftid", giftid)
		return 0
	} else {
		return order.Id
	}
}

// Wipe every order row.
func ClearOrders() error {
	return GiftDB.Where("id>0").Delete(&Order{}).Error
}
