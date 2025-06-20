package article

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/zmsocc/practice/webook/internal/repository"
	"github.com/zmsocc/practice/webook/pkg/logger"
	"github.com/zmsocc/practice/webook/pkg/saramax"
	"time"
)

type HistoryReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewHistoryReadEventConsumer(client sarama.Client, l logger.Logger,
	repo repository.InteractiveRepository) *HistoryReadEventConsumer {
	return &HistoryReadEventConsumer{
		client: client,
		l:      l,
		repo:   repo,
	}
}

func (k *HistoryReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", k.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{"read_article"},
			saramax.NewHandler[ReadEvent](k.l, k.Consume))
		if er != nil {
			k.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (k *HistoryReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return k.repo.AddRecord(ctx, t.Aid, t.Uid)
}
