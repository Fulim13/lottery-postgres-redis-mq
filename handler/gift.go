package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Fulim13/lottery/database"
	"github.com/Fulim13/lottery/mq"
	"github.com/Fulim13/lottery/util"
	"github.com/gin-gonic/gin"
)

const (
	PAY_DELAY = 600 // the user has this many seconds to pay, after that the prize goes back
)

// Every prize, used to fill in the wheel.
func GetAllGifts(ctx *gin.Context) {
	gifts := database.GetAllGifts()
	if len(gifts) == 0 {
		ctx.JSON(http.StatusInternalServerError, nil)
	} else {
		// hide anything the client has no business knowing
		for _, gift := range gifts {
			gift.Count = 1
		}
		ctx.JSON(http.StatusOK, gifts)
	}
}

// Draw a prize.
func Lottery(ctx *gin.Context) {
	for try := 0; try < 10; try++ { // at most 10 attempts
		gifts := database.GetAllGiftInventory() // remaining stock of every prize
		ids := make([]int, 0, len(gifts))
		probs := make([]float64, 0, len(gifts))
		for _, gift := range gifts {
			if gift.Count > 0 { // the draw doesn't support prizes with zero odds, so skip whatever Redis reports as empty
				ids = append(ids, gift.Id)
				probs = append(probs, float64(gift.Count))
			}
		}
		if len(ids) == 0 {
			ctx.String(http.StatusOK, strconv.Itoa(0)) // 0 means every prize has been handed out
			return
		}
		index := util.Lottery(probs) // the index of the prize that was drawn
		giftId := ids[index]
		err := database.ReduceInventory(giftId) // take it out of stock in Redis first
		if err != nil {
			// Say a prize has one unit left and several goroutines all draw it at the same time.
			// The first decrement lands on 0, the next goes negative, which means that draw lost
			// the race: it failed, so we loop around and try again.
			slog.Error("out of stock, could not decrement", "gift id", giftId)
			continue
		} else {
			uid := 1 // there is no login system, so the user id is hardcoded
			inst := database.GetGift(giftId)
			if inst == nil {
				slog.Error("gift not found", "gid", giftId)
				continue
			}
			database.CreateTempOrder(uid, giftId)                                      // create the temporary order
			mq.SendCancelOrder(database.Order{UserId: uid, GiftId: giftId}, PAY_DELAY) // delayed message telling the consumer to drop that temporary order
			slog.Info("prize drawn", "user", uid, "gift", giftId)

			// cookies first
			ctx.SetCookie("name", inst.Name, PAY_DELAY, "/", "localhost", false, false)                 // name of the prize
			ctx.SetCookie("price", strconv.Itoa(inst.Price), PAY_DELAY, "/", "localhost", false, false) // its price
			ctx.SetCookie("uid", strconv.Itoa(uid), PAY_DELAY, "/", "localhost", false, false)          // user id
			ctx.SetCookie("gid", strconv.Itoa(giftId), PAY_DELAY, "/", "localhost", false, false)       // gift id

			// then the body
			ctx.String(http.StatusOK, strconv.Itoa(giftId)) // only hand the prize id back once the stock was successfully decremented

			return
		}
	}
	ctx.String(http.StatusOK, strconv.Itoa(database.EMPTY_GIFT)) // still failing after 10 attempts, so return "Thanks for playing"
}
