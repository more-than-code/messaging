package repository

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	RedisUri      string `envconfig:"REDIS_URI"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`
}

type Repository struct {
	redisClient *redis.Client
}

type VerificationInfo struct {
	Code        string
	Attempt     int
	LastAttempt time.Time
}

func NewRepository() (*Repository, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisUri,
		Password: cfg.RedisPassword,
		DB:       0, // use default DB
	})

	return &Repository{redisClient: redisClient}, nil
}

func (r *Repository) GetVerificationInfo(ctx context.Context, phoneOrEmail string) (*VerificationInfo, error) {
	str, err := r.redisClient.Get(ctx, strings.ToLower(phoneOrEmail)).Result()

	var found *VerificationInfo

	if err != nil {
		return nil, nil
	}

	if str != "" {
		found = &VerificationInfo{}
		err = json.Unmarshal([]byte(str), found)
		if err != nil {
			return nil, err
		}
	}

	return found, nil
}

func (r *Repository) SetVerificationInfo(ctx context.Context, phoneOrEmail string, info *VerificationInfo) error {
	data, err := json.Marshal(info)
	str := string(data)
	if err != nil {
		return err
	}
	return r.redisClient.Set(ctx, strings.ToLower(phoneOrEmail), str, time.Minute*5).Err()
}

func (r *Repository) DeleteVerificationInfo(ctx context.Context, phoneOrEmail string) error {
	return r.redisClient.Del(ctx, phoneOrEmail).Err()
}
