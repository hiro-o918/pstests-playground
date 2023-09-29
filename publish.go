package main

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
)

func PublishMessage(ctx context.Context, topicID string, client *pubsub.Client) error {
	topic := client.Topic(topicID)
	defer topic.Stop()
	messages := []*pubsub.Message{
		{Data: []byte("Hello World")},
		{Data: []byte("Hello World")},
		{Data: []byte("Hello World")},
	}
	wg := sync.WaitGroup{}
	for _, m := range messages {
		result := topic.Publish(ctx, m)
		wg.Add(1)
		go func(rslt *pubsub.PublishResult) {
			defer wg.Done()
			// The Get method blocks until a server-generated ID or
			// an error is returned for the published message.
			id, err := rslt.Get(ctx)
			if err != nil {
				fmt.Printf("Failed to publish: %v", err)
				return
			}
			fmt.Printf("Published a message; msg ID: %v\n", id)
		}(result)
	}
	wg.Wait()
	return nil
}
