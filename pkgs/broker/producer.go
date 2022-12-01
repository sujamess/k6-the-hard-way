package broker

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"golang.org/x/exp/slog"
)

const (
	CreateOrderTopic = "order.create"
	UpdateCartTopic  = "cart.update"
)

type Producer interface {
	CreateTopic(topic string, options ...func(*createTopic)) error
	Publish(topic string, msg []byte)
	Close() error
}

type producer struct {
	producer sarama.SyncProducer
	broker   *sarama.Broker
}

func NewProducer(addrs string) Producer {
	slog.Info("broker: initializing a producer", slog.String("host", addrs))
	p, err := sarama.NewSyncProducer(strings.Split(addrs, ","), nil)
	if err != nil {
		panic(err)
	}
	slog.Info("broker: initailized a producer")
	return &producer{producer: p, broker: sarama.NewBroker(addrs)}
}

type createTopic struct {
	partitions *int32
}

func WithPartitions(p int32) func(*createTopic) {
	return func(ct *createTopic) {
		ct.partitions = &p
	}
}

func (p *producer) CreateTopic(topic string, options ...func(*createTopic)) error {
	ct := &createTopic{}
	for _, o := range options {
		o(ct)
	}
	partitions := int32(1)
	if ct.partitions != nil && *ct.partitions != 0 {
		partitions = *ct.partitions
	}
	res, err := p.broker.CreateTopics(&sarama.CreateTopicsRequest{
		TopicDetails: map[string]*sarama.TopicDetail{
			topic: {
				NumPartitions:     partitions,
				ReplicationFactor: int16(1),
				ConfigEntries:     make(map[string]*string),
			},
		},
	})
	if err != nil {
		return err
	}

	for _, v := range res.TopicErrors {
		if v.Err.Error() != "" {
			return v.Err
		}
	}

	return nil
}

func (p *producer) Publish(topic string, msg []byte) {
	partition, offset, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(string(msg)),
	})
	if err != nil {
		slog.Error(fmt.Sprintf("failed to send a message to topic '%s'", topic), err)
	}
	slog.Info("message sent", slog.Int("partition", int(partition)), slog.Int64("offset", offset))
}

func (p *producer) Close() error {
	err := p.producer.Close()
	if err != nil {
		return err
	}
	err = p.broker.Close()
	if err != nil {
		return err
	}
	return nil
}
