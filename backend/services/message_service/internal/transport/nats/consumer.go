package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Consumer struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func NewConsumer(nc *nats.Conn, js jetstream.JetStream) *Consumer {
	return &Consumer{
		nc: nc,
		js: js,
	}
}

// TODO: Write message butcher on create and delete
