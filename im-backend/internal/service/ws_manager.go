package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
}

type WsMessageType string

const (
	WsMessageTypePing WsMessageType = "ping"
	WsMessageTypePong WsMessageType = "pong"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
)

type wsMessage struct {
	Type WsMessageType `json:"type"`
}

type WSManager struct {
	Clients        map[string]*Client
	Broadcast      chan []byte
	Register       chan *Client
	Unregister     chan *Client
	redisClient    *redis.Client
	mutex          sync.RWMutex
	stop           chan struct{}
	messageService *MessageService
	sessionService *SessionService
}

func NewWSManager(redisClient *redis.Client, sessionService *SessionService) *WSManager {
	return &WSManager{
		Clients:        make(map[string]*Client),
		Broadcast:      make(chan []byte),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		redisClient:    redisClient,
		stop:           make(chan struct{}),
		sessionService: sessionService,
	}
}

// SetMessageService 设置消息服务
func (m *WSManager) SetMessageService(messageService *MessageService) {
	m.messageService = messageService
}

func (m *WSManager) Run() {
	m.startRedisSubscriber()

	for {
		select {
		case client := <-m.Register:
			m.mutex.Lock()
			if oldClient, ok := m.Clients[client.UserID]; ok && oldClient != client {
				delete(m.Clients, client.UserID)
				close(oldClient.Send)
				_ = oldClient.Conn.Close()
			}
			m.Clients[client.UserID] = client
			m.mutex.Unlock()
			log.Printf("Client connected: %s", client.UserID)

			// 更新用户在线状态
			if err := m.sessionService.UpdateUserStatus(context.Background(), client.UserID, true); err != nil {
				log.Printf("Failed to update user status: %v", err)
			}

		case client := <-m.Unregister:
			m.mutex.Lock()
			currentClient, ok := m.Clients[client.UserID]
			if ok && currentClient == client {
				delete(m.Clients, client.UserID)
				close(client.Send)
				log.Printf("Client disconnected: %s", client.UserID)

				// 更新用户离线状态
				if err := m.sessionService.UpdateUserStatus(context.Background(), client.UserID, false); err != nil {
					log.Printf("Failed to update user status: %v", err)
				}
			}
			m.mutex.Unlock()

		case message := <-m.Broadcast:
			var msg struct {
				ToID    string `json:"to_id"`
				FromID  string `json:"from_id"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			// 处理消息存储与转发...
			// 这部分代码和之前没有变化

		case <-m.stop:
			log.Println("WebSocket manager is shutting down...")
			return
		}
	}
}

// TryDeliver 尝试向本进程已连接的客户端投递消息
func (m *WSManager) TryDeliver(userID string, message []byte) bool {
	m.mutex.RLock()
	client, ok := m.Clients[userID]
	m.mutex.RUnlock()

	if !ok {
		return false
	}

	select {
	case client.Send <- message:
		return true
	default:
		log.Printf("Client send buffer full, dropping direct delivery for %s", userID)
		return false
	}
}

func (m *WSManager) startRedisSubscriber() {
	if m.redisClient == nil {
		return
	}

	pubsub := m.redisClient.Subscribe(context.Background(), wsPushChannel)
	log.Printf("Redis subscriber listening on channel: %s", wsPushChannel)
	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			var event WSPushEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("Failed to unmarshal ws push event: %v", err)
				continue
			}
			if m.TryDeliver(event.ReceiverID, []byte(event.Message)) {
				log.Printf("Delivered message to %s via redis push", event.ReceiverID)
			} else {
				log.Printf("No connected client for %s (redis push)", event.ReceiverID)
			}
		}
	}()
}

// SendMessage 发送消息给指定用户
func (m *WSManager) SendMessage(userID string, message []byte) error {
	m.mutex.RLock()
	client, ok := m.Clients[userID]
	m.mutex.RUnlock()

	if ok {
		client.Send <- message
	}
	return nil
}

// Shutdown 关闭 WebSocket 管理器
func (m *WSManager) Shutdown() {
	log.Println("Starting WebSocket manager shutdown...")

	// 发送停止信号
	close(m.stop)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 关闭所有客户端连接
	for _, client := range m.Clients {
		log.Printf("Closing connection for client: %s", client.UserID)
		close(client.Send)
		client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		client.Conn.Close()
	}

	// 清空客户端映射
	m.Clients = make(map[string]*Client)

	// 关闭所有通道
	close(m.Register)
	close(m.Unregister)
	close(m.Broadcast)

	log.Println("WebSocket manager shutdown completed")
}

// WritePump 处理向客户端写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				_ = w.Close()
				return
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 处理从客户端读取消息
func (c *Client) ReadPump(manager *WSManager) {
	defer func() {
		manager.Unregister <- c
		c.Conn.Close()
	}()

	// 设置读取超时，防止阻塞太久
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	// 设置PongHandler，接收到Pong时重置读取超时
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break // 连接错误时退出循环
		}

		// 解析消息内容
		var msg wsMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// 处理心跳消息
		if msg.Type == WsMessageTypePing {
			response := struct {
				Type WsMessageType `json:"type"`
			}{
				Type: WsMessageTypePong,
			}
			responseBytes, err := json.Marshal(response)
			if err != nil {
				log.Printf("Failed to marshal pong response: %v", err)
				continue
			}
			select {
			case c.Send <- responseBytes:
			default:
				log.Printf("Client send buffer full, dropping pong for %s", c.UserID)
			}
			// 重置读取超时并继续等待下一条消息
			_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
			continue
		}

		dbMessage, err := manager.messageService.ProcessWebSocketMessage(
			context.Background(),
			c.UserID,
			message,
		)
		if err != nil {
			log.Printf("Failed to process message: %v", err)
			continue
		}

		outMessage, err := json.Marshal(dbMessage)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			continue
		}

		// 推送给接收方（本进程或经 Redis 由 WS 进程投递）
		if err := manager.messageService.pushToUser(context.Background(), dbMessage.ReceiverID, outMessage); err != nil {
			log.Printf("Failed to push message to receiver: %v", err)
		}

		// 回显给发送方，便于多端同步
		if dbMessage.SenderID != dbMessage.ReceiverID {
			if err := manager.messageService.pushToUser(context.Background(), dbMessage.SenderID, outMessage); err != nil {
				log.Printf("Failed to push message to sender: %v", err)
			}
		}
	}
}
