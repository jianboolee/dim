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

type wsMessage struct {
	Type WsMessageType `json:"type"`
}

type WSManager struct {
	Clients           map[string]*Client
	Broadcast         chan []byte
	Register          chan *Client
	Unregister        chan *Client
	redisClient       *redis.Client
	mutex             sync.RWMutex
	stop              chan struct{}
	messageService    *MessageService
	sessionService    *SessionService
	heartbeatInterval time.Duration // 新增: 心跳间隔
	heartbeatTimeout  time.Duration // 新增: 超时检测
}

func NewWSManager(redisClient *redis.Client, sessionService *SessionService) *WSManager {
	return &WSManager{
		Clients:           make(map[string]*Client),
		Broadcast:         make(chan []byte),
		Register:          make(chan *Client),
		Unregister:        make(chan *Client),
		redisClient:       redisClient,
		stop:              make(chan struct{}),
		sessionService:    sessionService,
		heartbeatInterval: 30 * time.Second,
		heartbeatTimeout:  10 * time.Second,
	}
}

// SetMessageService 设置消息服务
func (m *WSManager) SetMessageService(messageService *MessageService) {
	m.messageService = messageService
}

func (m *WSManager) Run() {
	m.startRedisSubscriber()

	// 启动定时器，定期发送 ping 消息
	ticker := time.NewTicker(m.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case client := <-m.Register:
			m.mutex.Lock()
			m.Clients[client.UserID] = client
			m.mutex.Unlock()
			log.Printf("Client connected: %s", client.UserID)

			// 更新用户在线状态
			if err := m.sessionService.UpdateUserStatus(context.Background(), client.UserID, true); err != nil {
				log.Printf("Failed to update user status: %v", err)
			}

		case client := <-m.Unregister:
			m.mutex.Lock()
			if _, ok := m.Clients[client.UserID]; ok {
				delete(m.Clients, client.UserID)
				close(client.Send)
			}
			m.mutex.Unlock()
			log.Printf("Client disconnected: %s", client.UserID)

			// 更新用户离线状态
			if err := m.sessionService.UpdateUserStatus(context.Background(), client.UserID, false); err != nil {
				log.Printf("Failed to update user status: %v", err)
			}

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

		case <-ticker.C:
			// 每个心跳间隔发送 ping 消息
			m.mutex.RLock()
			for _, client := range m.Clients {
				if err := client.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Printf("Failed to send ping to %s: %v", client.UserID, err)
					client.Conn.Close()
					delete(m.Clients, client.UserID)
				}
			}
			m.mutex.RUnlock()
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
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			return
		}
	}

	// 通道已关闭，发送关闭消息
	c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

// ReadPump 处理从客户端读取消息
func (c *Client) ReadPump(manager *WSManager) {
	defer func() {
		manager.Unregister <- c
		c.Conn.Close()
	}()

	// 设置读取超时，防止阻塞太久
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 设置PongHandler，接收到Pong时重置读取超时
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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
			if err := c.Conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
				log.Printf("Failed to send pong: %v", err)
				break // 发送失败时退出循环
			}
			// 重置读取超时并继续等待下一条消息
			c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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
