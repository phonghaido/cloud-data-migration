package main

import (
	"crypto/subtle"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/phonghaido/cloud-data-migration/handlers"
	"github.com/phonghaido/cloud-data-migration/internal/config"
	"github.com/phonghaido/cloud-data-migration/pkg/http_error"
	"github.com/sirupsen/logrus"
)

func main() {
	systemConfig, err := config.GetSystemConfig()
	if err != nil {
		logrus.Fatalf("Error while parsing system config: %v", err)
	}

	awsClientConfig := config.GetAWSConfig()
	awsClient := handlers.NewAWSClient(awsClientConfig)

	redisConfig := config.GetRedisConfig()
	redisClient := handlers.NewRedisClient(redisConfig)
	defer redisClient.RedisClient.Close()

	pubsubClientConfig := config.GetPubSubConfig()
	pubsubClient, err := handlers.NewPubSucClient(pubsubClientConfig, systemConfig)
	if err != nil {
		logrus.Fatalf("Error while creating PubSub client: %v", err)
	}
	defer pubsubClient.PubSubClient.Close()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BasicAuth(func(username, password string, ctx echo.Context) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(username), []byte(systemConfig.AdminUsername)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(systemConfig.AdminPassword)) == 1 {
			return true, nil
		}
		return false, nil
	}))

	subHTTPHandler := handlers.NewSubHTTPHandler(awsClient, pubsubClient, redisClient)

	cacheGroup := e.Group("/cache")
	cacheGroup.GET("", http_error.ErrorWrapper(subHTTPHandler.HandlerFindAll))
	cacheGroup.GET("/type", http_error.ErrorWrapper(subHTTPHandler.HandlerFindRecordByType))
	cacheGroup.GET("/name", http_error.ErrorWrapper(subHTTPHandler.HandlerFindRecordByName))
	cacheGroup.DELETE("", http_error.ErrorWrapper(subHTTPHandler.HandlerDeleteAll))
	cacheGroup.DELETE("/type", http_error.ErrorWrapper(subHTTPHandler.HandlerDeleteRecordByType))
	cacheGroup.DELETE("/name", http_error.ErrorWrapper(subHTTPHandler.HandlerDeleteRecordByName))

	go func() {
		for {
			err := awsClient.PublishS3Keys(redisClient, pubsubClient)
			if err != nil {
				logrus.Fatalf("Error: %v", err)
			}
			time.Sleep(10 * time.Minute)
		}
	}()

	e.Logger.Fatal(e.Start(":8080"))
}
