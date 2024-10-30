package repository

import (
	"github.com/go-redis/redis/v8"
	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

type Repository struct {
	db          *gorm.DB
	MinioClient *minio.Client
	redisClient *redis.Client
}

func New(dsn string, m *minio.Client) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Настройки Redis
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	return &Repository{
		db:          db,
		MinioClient: m,
		redisClient: redisClient,
	}, nil
}
