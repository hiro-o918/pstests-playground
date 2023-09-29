package main

import (
	"context"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestPublishMessage(t *testing.T) {
	ctx := context.Background()

	// Start a fake server running locally.
	srv := pstest.NewServer()
	defer srv.Close()

	// Connect to the server without using TLS.
	conn, err := grpc.Dial(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.Dial: %v", err)
	}
	defer conn.Close()

	client, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("pubsub.NewClient: %v", err)
	}
	topicID := "topic"
	if _, err := client.CreateTopic(ctx, topicID); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	subscriptionID := "testSubscription"
	_, err = client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
		Topic: client.Topic(topicID),
	})
	if err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}
	defer client.Subscription(subscriptionID).Delete(ctx)

	if err := PublishMessage(ctx, topicID, client); err != nil {
		t.Fatalf("PublishMessage: %v", err)
	}

	// Use a channel to pass the received messages.
	msgCh := make(chan *pubsub.Message, 3) // assuming 3 messages as in your original PublishMessage function

	go func() {
		err := client.Subscription(subscriptionID).Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			msgCh <- msg
			msg.Ack()
		})
		if err != nil {
			log.Printf("receive: %v", err)
		}
	}()

	// Verify the received messages
	for i := 0; i < 3; i++ { // 3 messages
		select {
		case msg := <-msgCh:
			if got, want := string(msg.Data), "Hello World"; got != want {
				t.Errorf("msg.Data = %q; want %q", got, want)
			}
		case <-time.After(10 * time.Second): // timeout after waiting for 10 seconds
			t.Errorf("Did not receive message in time")
		}
	}

	defer client.Close()
}
