package main

import (
	"context"
	"encoding/json"

	"github.com/shinerio/skillflow/core/notify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func forwardEvents(ctx context.Context, hub *notify.Hub) {
	ch := hub.Subscribe()
	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(evt.Payload)
			runtime.EventsEmit(ctx, string(evt.Type), string(data))
		case <-ctx.Done():
			return
		}
	}
}
