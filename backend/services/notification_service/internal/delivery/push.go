package delivery

import (
	"context"
	"log"
)

type FCMMock struct {}

func (f *FCMMock) Send(ctx context.Context, userID string, title, body string) error {
	log.Printf("Mocked push provider call with %s %s %s", userID, title, body)
	return nil
}