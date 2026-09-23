package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Fulim13/lottery/database"
	"github.com/gin-gonic/gin"
)

// The user paid.
func Pay(ctx *gin.Context) {
	uid, err := strconv.Atoi(ctx.PostForm("uid"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	gid, err := strconv.Atoi(ctx.PostForm("gid"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Finding the temporary order proves they really won this prize.
	tempOrderGid := database.GetTempOrder(uid)
	if tempOrderGid != gid {
		ctx.String(http.StatusForbidden, "You did not win this prize, or the payment window has closed")
		return
	}

	// Turn it into a real order and drop the temporary one.
	if database.CreateOrder(uid, gid) > 0 {
		database.DeleteTempOrder(uid, gid)
		slog.Info("payment accepted, temp order removed", "uid", uid, "gid", gid)
	} else {
		ctx.String(http.StatusInternalServerError, "Sorry, something went wrong. Please contact support")
	}
}

// The user gave up the prize they won.
func GiveUp(ctx *gin.Context) {
	uid, err := strconv.Atoi(ctx.PostForm("uid"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	gid, err := strconv.Atoi(ctx.PostForm("gid"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// drop the temporary order
	database.DeleteTempOrder(uid, gid)
	// put the stock back
	database.IncreaseInventory(gid)
	slog.Info("user gave up the prize", "uid", uid, "gid", gid)
}
