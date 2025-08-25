package bot

import (
	"context"

	"github.com/devafterdark/project-lumos/pkg/slack/event"
)

type EventHandler interface {
	HandleEventsAPI(ctx context.Context, payload *event.EventsAPIPayload)
}
