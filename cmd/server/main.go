package main

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"subscriptions-go/api"
	"subscriptions-go/config"
	"subscriptions-go/db"
	"subscriptions-go/model"
	"subscriptions-go/repository"
	"subscriptions-go/service"

	_ "subscriptions-go/docs"
)

// @title Subscriptions API
// @version 1.0
// @description API для управления подписками
// @host localhost:8000
// @BasePath /
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logrus.New()
	level, _ := logrus.ParseLevel(cfg.LogLevel)
	log.SetLevel(level)

	gormDB, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := gormDB.AutoMigrate(&model.Subscription{}); err != nil {
		log.Fatal("auto migrate failed:", err)
	}

	repo := repository.NewSubscriptionRepo(gormDB)
	svc := service.NewSubscriptionService(repo)
	handler := api.NewHandler(svc, log)

	r := gin.Default()

	r.POST("/subscriptions", handler.Create)
	r.GET("/subscriptions", handler.List)
	r.GET("/subscriptions/:id", handler.Get)
	r.GET("/subscriptions/summary", handler.Summary)
	r.PUT("/subscriptions/:id", handler.Update)
	r.DELETE("/subscriptions/:id", handler.Delete)
	// TODO: add GET /subscriptions, GET/PUT/DELETE /subscriptions/:id (implement in handler)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	port := strconv.Itoa(cfg.AppPort)
	addr := fmt.Sprintf("%s:%s", cfg.AppHost, port)
	log.Infof("Starting server at %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
