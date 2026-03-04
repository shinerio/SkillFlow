package notify_test

import (
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/notify"
	"github.com/stretchr/testify/assert"
)

func TestHubPublishSubscribe(t *testing.T) {
	hub := notify.NewHub()
	ch := hub.Subscribe()
	defer hub.Unsubscribe(ch)

	hub.Publish(notify.Event{Type: notify.EventBackupStarted, Payload: nil})

	select {
	case evt := <-ch:
		assert.Equal(t, notify.EventBackupStarted, evt.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected event, got timeout")
	}
}

func TestHubMultipleSubscribers(t *testing.T) {
	hub := notify.NewHub()
	ch1 := hub.Subscribe()
	ch2 := hub.Subscribe()
	defer hub.Unsubscribe(ch1)
	defer hub.Unsubscribe(ch2)

	hub.Publish(notify.Event{Type: notify.EventSyncCompleted})

	for _, ch := range []<-chan notify.Event{ch1, ch2} {
		select {
		case evt := <-ch:
			assert.Equal(t, notify.EventSyncCompleted, evt.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("subscriber did not receive event")
		}
	}
}
