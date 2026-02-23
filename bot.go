package emi_transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	core "github.com/aK1r4z/emi-core"
	milky_types "github.com/aK1r4z/emi-core/types"
)

var ErrEventRegistryAlreadyExists = errors.New("event registry already exists")

type Bot struct {
	logger core.Logger

	*HttpClient
	eventSource core.EventSource

	eventRegistries map[milky_types.EventType]milky_types.Event
	eventHandlers   map[milky_types.EventType][]core.EventHandler

	closeChan chan any // 通道有可能为空！
}

func NewBot(logger core.Logger, httpClient *HttpClient, eventSource core.EventSource) *Bot {
	return &Bot{
		logger: logger,

		HttpClient:  httpClient,
		eventSource: eventSource,

		eventRegistries: map[milky_types.EventType]milky_types.Event{},
		eventHandlers:   map[milky_types.EventType][]core.EventHandler{},

		closeChan: nil,
	}
}

func (b *Bot) Open(ctx context.Context) error {
	eventChan, err := b.eventSource.Open()
	if err != nil {
		return err
	}

	b.closeChan = make(chan any)
	go b.handleEvent(eventChan)

	return nil
}

func (b *Bot) Close() error {
	return b.eventSource.Close()
}

func (b *Bot) Wait() {
	if b.closeChan == nil {
		return
	}
	<-b.closeChan
}

func (b *Bot) handleEvent(eventChan chan milky_types.RawEvent) {
	for rawEvent := range eventChan {

		// 获取对应的事件注册注册
		eventRegistry, ok := b.eventRegistries[rawEvent.Type]
		if !ok {
			b.logger.Warnf("Unknown event type: %s", rawEvent.Type)
			continue
		}

		// 获取对应的事件处理器
		handlers, ok := b.eventHandlers[eventRegistry.Type()]
		if !ok {
			b.logger.Tracef("No handler registered for event: %s", eventRegistry.Type())
			continue
		}

		// 把原始事件数据解码为实际对象
		event := eventRegistry.New()
		decoder := json.NewDecoder(bytes.NewReader(rawEvent.Data))
		if err := decoder.Decode(event); err != nil {
			b.logger.Errorf("Failed to decode event data: %v", err)
			continue
		}

		// [TODO] 异步处理事件
		for _, handler := range handlers {
			handler.Handle(b, event, rawEvent)
		}
	}

	go func() { b.closeChan <- struct{}{} }()
}

func (b *Bot) SetEventRegistry(eventType milky_types.EventType, event milky_types.Event) {
	b.eventRegistries[eventType] = event
}

func (b *Bot) Handle(handler core.EventHandler) {
	b.eventHandlers[handler.Type()] = append(b.eventHandlers[handler.Type()], handler)
}

func (b *Bot) Logger() core.Logger {
	return b.logger
}
