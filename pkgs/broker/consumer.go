package broker

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"golang.org/x/exp/slog"
)

type AfterConsumedFunc = func([]byte) error

type Consumer interface {
	Consume(topics string)
	Close() error
}

type consumer struct {
	ready             chan bool
	cg                sarama.ConsumerGroup
	afterConsumedFunc AfterConsumedFunc
}

func NewConsumer(addrs, groupID string, afterConsumedFunc AfterConsumedFunc) Consumer {
	slog.Info("broker: initializing a consumer group", slog.String("addrs", addrs), slog.String("groupID", groupID))
	cg, err := sarama.NewConsumerGroup(strings.Split(addrs, ","), groupID, sarama.NewConfig())
	if err != nil {
		panic(err)
	}
	slog.Info("broker: initialized a consumer group")
	return &consumer{ready: make(chan bool), cg: cg, afterConsumedFunc: afterConsumedFunc}
}

func (c *consumer) Consume(topic string) {
	slog.Info("broker: consuming", slog.String("topic", topic))
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.cg.Consume(ctx, []string{topic}, c); err != nil {
				slog.Error("consumer: failed to consume", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			c.ready = make(chan bool)
		}
	}()
	<-c.ready

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	keepRunning, consumptionIsPaused := true, false

	for keepRunning {
		select {
		case <-ctx.Done():
			slog.Warn("consumer: context cancelled after terminated")
			keepRunning = false
		case <-sigterm:
			slog.Warn("consumer: terminated via signal")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(c.cg, &consumptionIsPaused)
		}
	}

	cancel()
	wg.Wait()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			slog.Info("consumer: message claimed",
				slog.String("messageValue", string(message.Value)),
				slog.Time("timestamp", message.Timestamp),
				slog.String("topic", message.Topic),
			)
			if err := c.afterConsumedFunc(message.Value); err != nil {
				return err
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *consumer) Close() error {
	return c.cg.Close()
}

func toggleConsumptionFlow(cg sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		cg.ResumeAll()
		slog.Info("consumer: resuming consumption")
	} else {
		cg.PauseAll()
		slog.Info("consumer: pausing consumption")
	}

	*isPaused = !*isPaused
}
