package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StatResponse struct {
	Mongo Stat `json:"mongo"`
	Redis Stat `json:"redis"`
}

type Stat struct {
	IP     string `json:"ip" bson:"ip"`
	Visits int64  `json:"visits"`
}

func main() {
	api := fiber.New()

	iCtx, iCancel := context.WithTimeout(context.Background(), 2*time.Second)
	db, dErr := connectMongo(iCtx, os.Getenv("MONGO_CONNECT"), os.Getenv("MONGO_DB_NAME"))
	iCancel()
	if dErr != nil {
		logrus.WithError(dErr).Fatal("connect mongo")
	}

	rd := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	api.Use(
		cors.New(cors.Config{}),
		requestid.New(requestid.Config{}),
		logger.New(logger.Config{
			Format:     "${time} ${method} ${path} - ${status} - ${latency}\n",
			TimeFormat: "2006-01-02 15:04:05.000000",
			Output:     os.Stdout,
		}),
		recover.New(),
		limiter.New(limiter.Config{
			Expiration: 30 * time.Second,
			Max:        100,
		}),
	)

	api.Get("/health", func(c *fiber.Ctx) error {
		if db == nil {
			return fmt.Errorf("NOT OK")
		}

		return c.SendString("OK")
	})

	api.Get("/stats", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		sr := StatResponse{}
		var err error

		// Incr counters in mongo
		sr.Mongo, err = incMongoStats(db, ctx, Stat{
			IP:     c.IP(),
			Visits: 1,
		})
		if err != nil {
			return fmt.Errorf("inc mongo stats: %w", err)
		}

		// Incr counters in redis
		sr.Redis.IP = c.IP()
		sr.Redis.Visits, err = incRedisStats(rd, ctx, c.IP())
		if err != nil {
			return fmt.Errorf("inc redis stats: %w", err)
		}

		return c.JSON(sr)
	})

	err := api.Listen(":" + os.Getenv("API_PORT"))
	if err != nil {
		logrus.WithError(err).Fatal("api listen")
	}
}

func connectMongo(ctx context.Context, mongoConnect, dbName string) (*mongo.Database, error) {
	mongoCli, err := mongo.Connect(ctx, options.Client().ApplyURI(
		fmt.Sprintf("mongodb://%s", mongoConnect),
	))
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	err = mongoCli.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return mongoCli.Database(dbName), nil
}

func incMongoStats(db *mongo.Database, ctx context.Context, s Stat) (Stat, error) {
	filter := bson.D{
		{"ip", s.IP},
	}

	update := bson.D{{"$inc", bson.D{
		{"visits", 1},
	}}}

	after := options.After

	err := db.Collection("stats").FindOneAndUpdate(ctx, filter, update, &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		_, iErr := db.Collection("stats").InsertOne(ctx, s)
		if iErr != nil {
			return Stat{}, fmt.Errorf("insert one: %w", iErr)
		}

		return s, nil
	}
	if err != nil {
		return Stat{}, fmt.Errorf("find one and update: %w", err)
	}

	return s, nil
}

func incRedisStats(rd *redis.Client, ctx context.Context, ip string) (int64, error) {
	visits, err := rd.Incr(ctx, ip).Result()
	if err != nil {
		return 0, fmt.Errorf("incr: %w", err)
	}

	return visits, nil
}
