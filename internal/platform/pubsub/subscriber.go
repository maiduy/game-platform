package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pool "game-platform/internal/platform/pools"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// MessageHandler is a function that processes received messages
type MessageHandler func(ctx context.Context, msg *Message) error

// Message represents a received message
type Message struct {
	Subject  string
	Data     []byte
	Headers  map[string]string
	Metadata *nats.MsgMetadata
	msg      *nats.Msg
}

// Ack acknowledges the message
func (m *Message) Ack() error {
	if m.msg == nil {
		return fmt.Errorf("cannot ack: original message is nil")
	}
	return m.msg.Ack()
}

// Nak negatively acknowledges the message with a delay
func (m *Message) Nak(delay time.Duration) error {
	if m.msg == nil {
		return fmt.Errorf("cannot nak: original message is nil")
	}
	return m.msg.NakWithDelay(delay)
}

// Term terminates processing of the message
func (m *Message) Term() error {
	if m.msg == nil {
		return fmt.Errorf("cannot term: original message is nil")
	}
	return m.msg.Term()
}

// InProgress indicates that work is still in progress
func (m *Message) InProgress() error {
	if m.msg == nil {
		return fmt.Errorf("cannot send in-progress: original message is nil")
	}
	return m.msg.InProgress()
}

// Unmarshal unmarshals the message data into the provided interface
func (m *Message) Unmarshal(v interface{}) error {
	return json.Unmarshal(m.Data, v)
}

// Subscriber handles consuming messages from JetStream
type Subscriber struct {
	natsConn     *pool.NATSConnection
	js           nats.JetStreamContext
	subscription *nats.Subscription
}

// SubscribeOptions contains options for subscribing to messages
type SubscribeOptions struct {
	Stream        string
	Consumer      string
	Subject       string
	Durable       string
	AckWait       time.Duration
	MaxDeliver    int
	FilterSubject string
	DeliverPolicy nats.DeliverPolicy
	StartSequence uint64
	StartTime     time.Time
	ReplayPolicy  nats.ReplayPolicy
	MaxAckPending int
	MaxWaiting    int
	ManualAck     bool
	AckExplicit   bool
}

// NewSubscriber creates a new JetStream subscriber
func NewSubscriber() (*Subscriber, error) {
	natsConn, err := pool.NewNATSConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS connection: %w", err)
	}

	return &Subscriber{
		natsConn: natsConn,
		js:       natsConn.JetStream,
	}, nil
}

// Subscribe subscribes to a subject with the given handler
func (s *Subscriber) Subscribe(subject string, handler MessageHandler) error {
	opts := SubscribeOptions{
		Subject:   subject,
		ManualAck: true,
	}
	return s.SubscribeWithOptions(opts, handler)
}

// SubscribeWithOptions subscribes with custom options
func (s *Subscriber) SubscribeWithOptions(opts SubscribeOptions, handler MessageHandler) error {
	if !s.natsConn.IsConnected() {
		return fmt.Errorf("NATS connection is not active")
	}

	if handler == nil {
		return fmt.Errorf("message handler cannot be nil")
	}

	consumerConfig := buildConsumerConfig(opts)
	subOpts := buildSubscribeOptions(opts, consumerConfig)

	sub, err := s.js.Subscribe(opts.Subject, func(msg *nats.Msg) {
		ctx := context.Background()
		message := s.buildMessage(msg)

		if err := handler(ctx, message); err != nil {
			logrus.Errorf("Handler error for message on %s: %v", msg.Subject, err)
			if opts.ManualAck {
				msg.Nak()
			}
			return
		}

		if opts.ManualAck || opts.AckExplicit {
			if err := msg.Ack(); err != nil {
				logrus.Errorf("Failed to ack message: %v", err)
			}
		}
	}, subOpts...)

	if err != nil {
		logrus.Errorf("Failed to subscribe to %s: %v", opts.Subject, err)
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	s.subscription = sub
	logrus.Infof("Subscribed to %s (consumer: %s)", opts.Subject, opts.Consumer)
	return nil
}

// QueueSubscribe creates a queue subscription for load balancing
func (s *Subscriber) QueueSubscribe(subject, queue string, handler MessageHandler) error {
	if !s.natsConn.IsConnected() {
		return fmt.Errorf("NATS connection is not active")
	}

	if handler == nil {
		return fmt.Errorf("message handler cannot be nil")
	}

	sub, err := s.js.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		ctx := context.Background()
		message := s.buildMessage(msg)

		if err := handler(ctx, message); err != nil {
			logrus.Errorf("Handler error for message on %s: %v", msg.Subject, err)
			msg.Nak()
			return
		}

		if err := msg.Ack(); err != nil {
			logrus.Errorf("Failed to ack message: %v", err)
		}
	}, nats.ManualAck())

	if err != nil {
		logrus.Errorf("Failed to queue subscribe to %s: %v", subject, err)
		return fmt.Errorf("failed to queue subscribe: %w", err)
	}

	s.subscription = sub
	logrus.Infof("Queue subscribed to %s (queue: %s)", subject, queue)
	return nil
}

// PullSubscribe creates a pull-based subscription
func (s *Subscriber) PullSubscribe(opts SubscribeOptions) error {
	if !s.natsConn.IsConnected() {
		return fmt.Errorf("NATS connection is not active")
	}

	consumerConfig := buildConsumerConfig(opts)
	subOpts := buildPullSubscribeOptions(opts, consumerConfig)

	sub, err := s.js.PullSubscribe(opts.Subject, opts.Durable, subOpts...)
	if err != nil {
		logrus.Errorf("Failed to pull subscribe to %s: %v", opts.Subject, err)
		return fmt.Errorf("failed to pull subscribe: %w", err)
	}

	s.subscription = sub
	logrus.Infof("Pull subscribed to %s (consumer: %s)", opts.Subject, opts.Consumer)
	return nil
}

// Fetch fetches messages from a pull subscription
func (s *Subscriber) Fetch(batchSize int, timeout time.Duration) ([]*Message, error) {
	if s.subscription == nil {
		return nil, fmt.Errorf("no active subscription")
	}

	msgs, err := s.subscription.Fetch(batchSize, nats.MaxWait(timeout))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	messages := make([]*Message, len(msgs))
	for i, msg := range msgs {
		messages[i] = s.buildMessage(msg)
	}

	return messages, nil
}

// CreateOrUpdateConsumer creates or updates a consumer
func (s *Subscriber) CreateOrUpdateConsumer(streamName string, consumerConfig *nats.ConsumerConfig) error {
	if consumerConfig == nil {
		return fmt.Errorf("consumer config cannot be nil")
	}

	_, err := s.natsConn.ExecuteWithRetry(func() (interface{}, error) {
		info, err := s.js.ConsumerInfo(streamName, consumerConfig.Durable)
		if err != nil {
			return s.js.AddConsumer(streamName, consumerConfig)
		}

		if info != nil {
			return s.js.UpdateConsumer(streamName, consumerConfig)
		}

		return s.js.AddConsumer(streamName, consumerConfig)
	})

	if err != nil {
		logrus.Errorf("Failed to create/update consumer %s: %v", consumerConfig.Durable, err)
		return fmt.Errorf("failed to create/update consumer: %w", err)
	}

	logrus.Infof("Consumer %s created/updated successfully", consumerConfig.Durable)
	return nil
}

// Unsubscribe unsubscribes from the current subscription
func (s *Subscriber) Unsubscribe() error {
	if s.subscription != nil {
		err := s.subscription.Unsubscribe()
		if err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
		s.subscription = nil
		logrus.Info("Unsubscribed successfully")
	}
	return nil
}

// Drain drains the subscription
func (s *Subscriber) Drain() error {
	if s.subscription != nil {
		err := s.subscription.Drain()
		if err != nil {
			return fmt.Errorf("failed to drain subscription: %w", err)
		}
		logrus.Info("Subscription drained successfully")
	}
	return nil
}

// Close closes the subscriber
func (s *Subscriber) Close() error {
	return s.Unsubscribe()
}

func (s *Subscriber) buildMessage(msg *nats.Msg) *Message {
	message := &Message{
		Subject: msg.Subject,
		Data:    msg.Data,
		Headers: make(map[string]string),
		msg:     msg,
	}

	if msg.Header != nil {
		for key := range msg.Header {
			message.Headers[key] = msg.Header.Get(key)
		}
	}

	if metadata, err := msg.Metadata(); err == nil {
		message.Metadata = metadata
	}

	return message
}

func buildConsumerConfig(opts SubscribeOptions) *nats.ConsumerConfig {
	config := &nats.ConsumerConfig{
		Durable:       opts.Durable,
		DeliverPolicy: opts.DeliverPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		ReplayPolicy:  opts.ReplayPolicy,
	}

	if opts.AckWait > 0 {
		config.AckWait = opts.AckWait
	}

	if opts.MaxDeliver > 0 {
		config.MaxDeliver = opts.MaxDeliver
	}

	if opts.FilterSubject != "" {
		config.FilterSubject = opts.FilterSubject
	}

	if opts.StartSequence > 0 {
		config.OptStartSeq = opts.StartSequence
	}

	if !opts.StartTime.IsZero() {
		config.OptStartTime = &opts.StartTime
	}

	if opts.MaxAckPending > 0 {
		config.MaxAckPending = opts.MaxAckPending
	}

	if opts.MaxWaiting > 0 {
		config.MaxWaiting = opts.MaxWaiting
	}

	return config
}

func buildSubscribeOptions(opts SubscribeOptions, config *nats.ConsumerConfig) []nats.SubOpt {
	subOpts := []nats.SubOpt{}

	if opts.Stream != "" {
		subOpts = append(subOpts, nats.BindStream(opts.Stream))
	}

	if opts.Consumer != "" {
		subOpts = append(subOpts, nats.Bind(opts.Stream, opts.Consumer))
	} else if opts.Durable != "" {
		subOpts = append(subOpts, nats.Durable(opts.Durable))
	}

	if opts.ManualAck || opts.AckExplicit {
		subOpts = append(subOpts, nats.ManualAck())
	}

	return subOpts
}

func buildPullSubscribeOptions(opts SubscribeOptions, config *nats.ConsumerConfig) []nats.SubOpt {
	subOpts := []nats.SubOpt{}

	if opts.Stream != "" {
		subOpts = append(subOpts, nats.BindStream(opts.Stream))
	}

	return subOpts
}
