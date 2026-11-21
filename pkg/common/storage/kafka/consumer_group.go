package kafka

import (
	"context"
	"errors"

	"server/v3/pkg/common/log"
	"github.com/IBM/sarama"
)

type MConsumerGroup struct {
	sarama.ConsumerGroup
	groupID string
	topics  []string
}

func (mc *MConsumerGroup) GetContextFromMsg(cMsg *sarama.ConsumerMessage) context.Context {
	return GetContextWithMQHeader(cMsg.Headers)
}

func (mc *MConsumerGroup) RegisterHandleAndConsumer(ctx context.Context, handler sarama.ConsumerGroupHandler) {
	for {
		err := mc.ConsumerGroup.Consume(ctx, mc.topics, handler)
		if errors.Is(err, sarama.ErrClosedConsumerGroup) {
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		if err != nil {
			log.ZWarn(ctx, "consume err", err, "topic", mc.topics, "groupID", mc.groupID)
		}
	}
}

func (mc *MConsumerGroup) Close() error {
	return mc.ConsumerGroup.Close()
}
