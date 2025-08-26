package bot

import (
	"context"

	"github.com/devafterdark/project-lumos/pkg/slack/event"
)

type Handler interface {
	HandleEventsAPI(ctx context.Context, payload *event.EventsAPIPayload)
}
