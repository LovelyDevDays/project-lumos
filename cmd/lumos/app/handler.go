package app

import (
	"context"
	"log/slog"

	"github.com/devafterdark/project-lumos/cmd/lumos/app/chain"
	"github.com/devafterdark/project-lumos/cmd/lumos/app/chat"
	"github.com/devafterdark/project-lumos/pkg/slack/event"
)

type BotHandler struct {
	chatHandler chat.Handler
}

func NewBotHandler(appToken, botToken string) *BotHandler {
	return &BotHandler{
		chatHandler: BuildChatHandlerChain(appToken, botToken),
	}
}

func (b *BotHandler) HandleEventsAPI(ctx context.Context, payload *event.EventsAPIPayload) {
	if payload.Type != event.EventsAPITypeEventCallback {
		return
	}

	e := payload.OfEventCallback.Event
	switch e.Type {
	case event.EventTypeAssistantThreadStarted:
		// TODO: Implement thread started handling
	case event.EventTypeAssistantThreadContextChanged:
		// TODO: Implement thread context changed handling
	case event.EventTypeMessage:
		b.chatHandler.HandleChat(&chat.Chat{
			Channel:   e.OfMessage.Channel,
			Timestamp: e.OfMessage.Timestamp,
			Thread:    []string{e.OfMessage.Text},
		})
	default:
		slog.Warn("unknown event type", slog.String("type", string(e.Type)))
	}
}

func BuildChatHandlerChain(appToken, botToken string) chat.Handler {
	handler := chain.ResponseHandler()

	// 메시지 생성 핸들러 설정.
	handler = chain.WithResponseGeneration(handler)
	handler = chain.WithAssistantStatus(handler, "generating response...")

	// 패시지 검색 핸들러 설정.
	handler = chain.WithPassageRetrieval(handler)
	handler = chain.WithAssistantStatus(handler, "retrieving passages...")

	// 슬랙 클라이언트 초기화 핸들러 설정.
	handler = chain.WithSlackClientInit(handler, appToken, botToken)

	// 패닉 복구 핸들러 설정.
	handler = chain.WithPanicRecovery(handler)

	return handler
}
