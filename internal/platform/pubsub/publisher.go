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

// Publisher handles publishing messages to JetStream
type Publisher struct {
	natsConn *pool.NATSConnection
	js       nats.JetStreamContext
}

// PublishOptions contains options for publishing messages
type PublishOptions struct {
	Subject      string
	Headers      map[string]string
	MsgID        string
	ExpectStream string
	Timeout      time.Duration
}

// NewPublisher creates a new JetStream publisher
func NewPublisher() (*Publisher, error) {
	natsConn, err := pool.NewNATSConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS connection: %w", err)
	}

	return &Publisher{
		natsConn: natsConn,
		js:       natsConn.JetStream,
	}, nil
}

// Publish publishes a message to the specified subject
func (p *Publisher) Publish(ctx context.Context, subject string, data interface{}) error {
	opts := PublishOptions{
		Subject: subject,
		Timeout: 5 * time.Second,
	}
	return p.PublishWithOptions(ctx, opts, data)
}

// PublishWithOptions publishes a message with custom options
func (p *Publisher) PublishWithOptions(ctx context.Context, opts PublishOptions, data interface{}) error {
	if !p.natsConn.IsConnected() {
		return fmt.Errorf("NATS connection is not active")
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &nats.Msg{
		Subject: opts.Subject,
		Data:    payload,
	}

	if opts.Headers != nil {
		msg.Header = nats.Header{}
		for k, v := range opts.Headers {
			msg.Header.Add(k, v)
		}
	}

	pubOpts := []nats.PubOpt{}
	if opts.MsgID != "" {
		pubOpts = append(pubOpts, nats.MsgId(opts.MsgID))
	}
	if opts.ExpectStream != "" {
		pubOpts = append(pubOpts, nats.ExpectStream(opts.ExpectStream))
	}

	if opts.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
		pubOpts = append(pubOpts, nats.Context(ctx))
	}

	result, err := p.natsConn.ExecuteWithRetry(func() (interface{}, error) {
		return p.js.PublishMsg(msg, pubOpts...)
	})

	if err != nil {
		logrus.Errorf("Failed to publish message to %s: %v", opts.Subject, err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	ack, ok := result.(*nats.PubAck)
	if !ok {
		return fmt.Errorf("unexpected publish result type")
	}

	logrus.Debugf("Published message to %s (stream: %s, seq: %d)", opts.Subject, ack.Stream, ack.Sequence)
	return nil
}

// PublishAsync publishes a message asynchronously
func (p *Publisher) PublishAsync(subject string, data interface{}, handler func(*nats.PubAck, error)) error {
	if !p.natsConn.IsConnected() {
		return fmt.Errorf("NATS connection is not active")
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ackFuture, err := p.js.PublishAsync(subject, payload)
	if err != nil {
		return fmt.Errorf("failed to publish async: %w", err)
	}

	if handler != nil {
		go func() {
			select {
			case ack := <-ackFuture.Ok():
				logrus.Debugf("Async published to %s (stream: %s, seq: %d)", subject, ack.Stream, ack.Sequence)
				handler(ack, nil)
			case err := <-ackFuture.Err():
				logrus.Errorf("Async publish failed for %s: %v", subject, err)
				handler(nil, err)
			}
		}()
	}

	return nil
}

// CreateOrUpdateStream creates or updates a JetStream stream
func (p *Publisher) CreateOrUpdateStream(streamConfig *nats.StreamConfig) error {
	if streamConfig == nil {
		return fmt.Errorf("stream config cannot be nil")
	}

	_, err := p.natsConn.ExecuteWithRetry(func() (interface{}, error) {
		info, err := p.js.StreamInfo(streamConfig.Name)
		if err != nil {
			return p.js.AddStream(streamConfig)
		}

		if info != nil {
			return p.js.UpdateStream(streamConfig)
		}

		return p.js.AddStream(streamConfig)
	})

	if err != nil {
		logrus.Errorf("Failed to create/update stream %s: %v", streamConfig.Name, err)
		return fmt.Errorf("failed to create/update stream: %w", err)
	}

	logrus.Infof("Stream %s created/updated successfully", streamConfig.Name)
	return nil
}

// DeleteStream deletes a JetStream stream
func (p *Publisher) DeleteStream(streamName string) error {
	err := p.js.DeleteStream(streamName)
	if err != nil {
		logrus.Errorf("Failed to delete stream %s: %v", streamName, err)
		return fmt.Errorf("failed to delete stream: %w", err)
	}

	logrus.Infof("Stream %s deleted successfully", streamName)
	return nil
}

// StreamInfo returns information about a stream
func (p *Publisher) StreamInfo(streamName string) (*nats.StreamInfo, error) {
	info, err := p.js.StreamInfo(streamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}
	return info, nil
}

// Close closes the publisher
func (p *Publisher) Close() error {
	return nil
}
