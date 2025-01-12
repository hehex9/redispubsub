package redispubsub_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	_ "github.com/hehex9/redispubsub"
	"gocloud.dev/pubsub"
)

func TestBasicUsage(t *testing.T) {
	ctx := context.Background()

	orig1 := &pubsub.Message{
		Body: []byte("Message #1"),
		// Metadata is optional and can be nil.
		Metadata: map[string]string{
			// These are examples of metadata.
			// There is nothing special about the key names.
			"language":   "en",
			"importance": "high",
		},
	}

	orig2 := &pubsub.Message{
		Body: []byte("Message #2"),
		// Metadata is optional and can be nil.
		Metadata: map[string]string{
			// These are examples of metadata.
			// There is nothing special about the key names.
			"language": "en",
		},
	}

	// send before consumer attach
	pubTest(ctx, orig1, t)
	time.Sleep(100 * time.Millisecond)
	pubTest(ctx, orig2, t)
	time.Sleep(100 * time.Millisecond)
	pubTest(ctx, orig1, t)
	time.Sleep(100 * time.Millisecond)

	// attach consumer and create group if needed
	subTest(ctx, orig1, t)
	time.Sleep(100 * time.Millisecond)
	subTest(ctx, orig2, t)
	time.Sleep(100 * time.Millisecond)
	subTest(ctx, orig1, t)
	time.Sleep(100 * time.Millisecond)

	res, err := redisCli.XPending(ctx, "topics/1", "group1").Result()
	if res.Count != 0 {
		t.Error(res.Count, err)
	}
}

func pubTest(ctx context.Context, orig *pubsub.Message, t *testing.T) {
	topic, err := pubsub.OpenTopic(ctx, "redis://topics/1")
	if err != nil {
		t.Errorf("could not open topic: %v", err)
		return
	}
	defer topic.Shutdown(ctx)

	err = topic.Send(ctx, orig)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(100 * time.Millisecond)
	err = topic.Send(ctx, orig)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(100 * time.Millisecond)
	err = topic.Send(ctx, orig)
	if err != nil {
		t.Error(err)
		return
	}
}

func subTest(ctx context.Context, orig *pubsub.Message, t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		subs, err := pubsub.OpenSubscription(ctx, "redis://group1?consumer=cons1&topic=topics/1")
		if err != nil {
			t.Error(err)
			return
		}
		defer subs.Shutdown(ctx)
		for i := 0; i < 4; i++ {
			msg, err := subs.Receive(ctx)
			if err != nil {
				// Errors from Receive indicate that Receive will no longer succeed.
				t.Errorf("Receiving message: %v", err)
				return
			}
			// Do work based on the message, for example:
			t.Logf("Got message %s: %q\n", msg.LoggableID, msg.Body)
			// Emulate not ack message 0
			if i > 0 {
				// Messages must always be acknowledged with Ack.
				msg.Ack()
				// wait for Ack asynchronous (see docs for Ack)
				time.Sleep(100 * time.Millisecond)
			}

			if !bytes.Equal(msg.Body, orig.Body) {
				t.Error("body not equal")
				return
			}
			for k, v := range msg.Metadata {
				if orig.Metadata[k] != v {
					t.Error("metadata not equal")
					return
				}
			}
		}
	}()
	<-done
}
