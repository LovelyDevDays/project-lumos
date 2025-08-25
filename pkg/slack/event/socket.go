package event

import (
	"encoding/json"
)

type SocketEventType string

const (
	SocketEventTypeHello      SocketEventType = "hello"
	SocketEventTypeDisconnect SocketEventType = "disconnect"
	SocketEventTypeEventsAPI  SocketEventType = "events_api"
)

type SocketEvent struct {
	Type SocketEventType `json:"type"`

	OfHello      *Hello      `json:"-"`
	OfDisconnect *Disconnect `json:"-"`
	OfEventsAPI  *EventsAPI  `json:"-"`

	Raw []byte `json:"-"`
}

func (s *SocketEvent) UnmarshalJSON(data []byte) error {
	type alias SocketEvent

	raw := &alias{}
	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}

	s.Type = raw.Type
	switch raw.Type {
	case SocketEventTypeHello:
		s.OfHello = &Hello{}
		if err := json.Unmarshal(data, s.OfHello); err != nil {
			return err
		}
	case SocketEventTypeDisconnect:
		s.OfDisconnect = &Disconnect{}
		if err := json.Unmarshal(data, s.OfDisconnect); err != nil {
			return err
		}
	case SocketEventTypeEventsAPI:
		s.OfEventsAPI = &EventsAPI{}
		if err := json.Unmarshal(data, s.OfEventsAPI); err != nil {
			return err
		}
	}
	s.Raw = data

	return nil
}

type Hello struct {
	ConnectionCount int `json:"num_connections"`
	ConnectionInfo  struct {
		AppID string `json:"app_id"`
	} `json:"connection_info"`
}

type Disconnect struct {
	Reason string `json:"reason"`
}
