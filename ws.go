package emi_transport

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	core "github.com/aK1r4z/emi-core"
	milky_types "github.com/aK1r4z/emi-core/types"
	"github.com/gorilla/websocket"
)

var ErrAlreadyConnected = errors.New("already connected")

type WebsocketEventSource struct {
	sync.RWMutex

	logger core.Logger

	wsGateway   string
	accessToken string

	wsConn *websocket.Conn

	eventChan chan milky_types.RawEvent
	closeChan chan any
}

func NewWebsocketEventSource(logger core.Logger, wsGateway string, accessToken string) *WebsocketEventSource {
	return &WebsocketEventSource{
		logger: logger,

		wsGateway:   wsGateway,
		accessToken: accessToken,

		wsConn: nil,

		eventChan: nil,
		closeChan: nil,
	}
}

func (w *WebsocketEventSource) Wait() {
	<-w.closeChan
}

// 开启
func (w *WebsocketEventSource) Open() (chan milky_types.RawEvent, error) {
	w.Lock()
	defer w.Unlock()

	if w.wsConn != nil {
		return nil, ErrAlreadyConnected
	}

	dialer := websocket.DefaultDialer

	header := http.Header{}
	if w.accessToken != "" {
		header.Add("Authorization", "Bearer "+w.accessToken)
	}

	wsConn, _, err := dialer.Dial(w.wsGateway, header)
	if err != nil {
		return nil, err
	}

	w.wsConn = wsConn
	w.eventChan = make(chan milky_types.RawEvent)
	w.closeChan = make(chan any)

	go w.receive(wsConn, w.eventChan, w.closeChan)

	return w.eventChan, nil
}

// 关闭
func (w *WebsocketEventSource) Close() error {
	w.Lock()
	defer w.Unlock()

	if w.wsConn == nil {
		return nil
	}

	err := w.wsConn.Close()
	if err != nil {
		return err
	}

	w.wsConn = nil
	close(w.eventChan)
	close(w.closeChan)

	return nil
}

func (w *WebsocketEventSource) receive(
	wsConn *websocket.Conn,
	eventChan chan milky_types.RawEvent,
	closeChan chan any,
) {
	for {
		messageType, message, err := wsConn.ReadMessage()

		// 在读取消息过程中出现错误
		if err != nil {

			w.RLock()
			ws := w.wsConn
			w.RUnlock()

			// 如果当前连接已经关闭，停止接收消息
			if wsConn != ws {
				return
			}

			// [TODO] 如果连接仍在运行中，上报错误信息，然后尝试重连
			w.logger.Errorf("Error when reading message: %v", err)

			err := w.Close()
			if err != nil {
				w.logger.Errorf("Failed to close websocket connection: %v", err)
				// [TODO] 错误处理
			}

			return // [TODO] 重连
		}

		// 读取消息
		messageBytes := message

		// 如果消息是压缩的，使用 zlib 解压
		if messageType == websocket.BinaryMessage {
			zlib, err := zlib.NewReader(bytes.NewReader(message))
			if err != nil {
				w.logger.Errorf("Failed to decompress message: %v", err)
				// [TODO] 错误处理
			}

			messageBytes, err = io.ReadAll(zlib)
			if err != nil {
				w.logger.Errorf("Failed to read decompressed message: %v", err)
				// [TODO] 错误处理
			}

			err = zlib.Close()
			if err != nil {
				w.logger.Errorf("Failed to close zlib reader: %v", err)
				// [TODO] 错误处理
			}
		}

		// 把事件解码为结构体
		rawEvent := milky_types.RawEvent{}
		if err = json.Unmarshal(messageBytes, &rawEvent); err != nil {
			w.logger.Errorf("Failed to decode message: %v", err)
			// [TODO] 错误处理
		}
		w.logger.Debugf("Received event: {event_type: %s, self_id: %d, time: %d, data: %s}", rawEvent.Type, rawEvent.SelfID, rawEvent.Time, rawEvent.Data)

		// 发送事件
		eventChan <- rawEvent

		// 查看通道是否关闭
		select {
		case <-closeChan:
			return
		default:
		}
	}
}
