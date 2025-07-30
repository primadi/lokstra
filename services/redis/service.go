package redis

import (
	"context"

	"github.com/primadi/lokstra/serviceapi"
	redisgo "github.com/redis/go-redis/v9"
)

type Service struct {
	client *redisgo.Client
}

// Client implements serviceapi.Redis.
func (s *Service) Client() *redisgo.Client {
	return s.client
}

var _ serviceapi.Redis = (*Service)(nil)

func New(client *redisgo.Client) *Service {
	return &Service{client: client}
}

func NewDsn(dsn string) *Service {
	opt, err := redisgo.ParseURL(dsn)
	if err != nil {
		return nil
	}
	client := redisgo.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil
	}
	return &Service{client: client}
}
