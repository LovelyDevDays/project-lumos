package event

import (
	"encoding/json"
)

type EventsAPI struct {
	EnvelopeID             string            `json:"envelope_id"`
	Payload                *EventsAPIPayload `json:"payload,omitempty"`
	AcceptsResponsePayload bool              `json:"accepts_response_payload"`
	RetryAttempt           int               `json:"retry_attempt"`
	RetryReason            string            `json:"retry_reason"`
}

type EventsAPIType string

const (
	EventsAPITypeEventCallback EventsAPIType = "event_callback"
)

type EventsAPIPayload struct {
	Type EventsAPIType `json:"type"`

	OfEventCallback *EventCallback `json:"-"`
}

func (p *EventsAPIPayload) UnmarshalJSON(data []byte) error {
	type alias EventsAPIPayload

	raw := &alias{}
	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}

	p.Type = raw.Type
	switch raw.Type {
	case EventsAPITypeEventCallback:
		p.OfEventCallback = &EventCallback{}
		if err := json.Unmarshal(data, p.OfEventCallback); err != nil {
			return err
		}
	}

	return nil
}

type EventCallback struct {
	EventID string `json:"event_id"`
	Event   Event  `json:"event"`
}
