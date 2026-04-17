package router

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
)

type App struct {
	service *gin.Engine
	addr    string
}

func NewApp(port int, router IRouter) *App {
	engine := gin.New()
	engine.Use(gin.Recovery())
	router.Register(engine)
	return &App{
		service: engine,
		addr:    ":" + strconv.Itoa(port),
	}
}

func (app *App) Run() {
	https := http.Server{
		Addr:    app.addr,
		Handler: app.service,
	}
	go func() {
		if err := https.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 主协程卡在这里，直到收到退出信号

	log.Println("Shutting Down Server...")
	if err := https.Shutdown(context.Background()); err != nil {
		log.Fatal("Shutdown err:", err)
	}
}
