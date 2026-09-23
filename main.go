package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fulim13/lottery/database"
	"github.com/Fulim13/lottery/handler"
	"github.com/Fulim13/lottery/mq"
	"github.com/Fulim13/lottery/util"
	"github.com/gin-gonic/gin"
)

var (
	server *http.Server
)

func Init() {
	util.InitSlog("./log/lottery.log")
	database.ConnectGiftDB("./conf", "postgres", util.YAML, "./log")
	database.ConnectGiftRedis("./conf", "redis", util.YAML)
	mq.InitMQLog() // send the amqp client's own messages to slog
	mq.InitMQ("./conf", "rabbitmq", util.YAML)
	go mq.ReceiveCancelOrder()
	database.InitGiftInventory()
}

func ListenTermSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	slog.Info("receive term signal " + sig.String() + ", going to exit")

	// release everything we hold
	database.CloseGiftDB()
	database.CloseGiftRedis()
	mq.Close() // stop the consumer, close the producer, drop the connection

	// wait for the web server to finish shutting down
	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx) // Shutdown ends the process
	}
}

func main() {
	Init()
	go ListenTermSignal()

	gin.SetMode(gin.ReleaseMode)   // run gin in release mode
	gin.DefaultWriter = io.Discard // silence gin's own output
	engine := gin.Default()

	// static assets are read from disk, so editing them only needs a page refresh, not a restart
	engine.Static("/js", "views/js")
	engine.Static("/img", "views/img")
	engine.LoadHTMLGlob("views/html/*.html")

	engine.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "lottery.html", nil)
	})
	engine.GET("/gifts", handler.GetAllGifts) // every prize, to fill in the wheel
	engine.GET("/lucky", handler.Lottery)     // the draw button
	engine.POST("/giveup", handler.GiveUp)
	engine.POST("/pay", handler.Pay)
	engine.GET("/result", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "pay.html", nil)
	})

	server = &http.Server{
		Addr:    "localhost:5678",
		Handler: engine,
	}
	slog.Info("lottery server started", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

// go run .
// Open http://localhost:5678/ in a browser. The project uses cookies, so the host has to be localhost.
