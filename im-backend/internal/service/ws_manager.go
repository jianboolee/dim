package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	UserID    string
	Conn      *websocket.Conn
	Send      chan []byte
	closeOnce sync.Once
}

type WsMessageType string

const (
	WsMessageTypePing WsMessageType = "ping"
	WsMessageTypePong WsMessageType = "pong"
)

const defaultWriteWait = 10 * time.Second
const defaultPongWait = 75 * time.Second
const defaultPingPeriod = 30 * time.Second

type WSManagerOptions struct {
	PingPeriod time.Duration
	PongWait   time.Duration
	WriteWait  time.Duration
}

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
	done           chan struct{}
	stopOnce       sync.Once
	redisCancel    context.CancelFunc
	messageService *MessageService
	sessionService *SessionService
	pingPeriod     time.Duration
	pongWait       time.Duration
	writeWait      time.Duration
}

func NewWSManager(redisClient *redis.Client, sessionService *SessionService, options WSManagerOptions) *WSManager {
	if options.PingPeriod <= 0 {
		options.PingPeriod = defaultPingPeriod
	}
	if options.PongWait <= 0 {
		options.PongWait = defaultPongWait
	}
	if options.WriteWait <= 0 {
		options.WriteWait = defaultWriteWait
	}
	if options.PongWait <= options.PingPeriod {
		options.PongWait = options.PingPeriod + 30*time.Second
	}

	return &WSManager{
		Clients:        make(map[string]*Client),
		Broadcast:      make(chan []byte),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		redisClient:    redisClient,
		stop:           make(chan struct{}),
		done:           make(chan struct{}),
		sessionService: sessionService,
		pingPeriod:     options.PingPeriod,
		pongWait:       options.PongWait,
		writeWait:      options.WriteWait,
	}
}

// SetMessageService 设置消息服务
func (m *WSManager) SetMessageService(messageService *MessageService) {
	m.messageService = messageService
}

func (m *WSManager) Run() {
	m.startRedisSubscriber()
	defer close(m.done)

	for {
		select {
		case client := <-m.Register:
			m.mutex.Lock()
			if oldClient, ok := m.Clients[client.UserID]; ok && oldClient != client {
				delete(m.Clients, client.UserID)
				oldClient.closeSend()
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
			m.unregister(client)

		case message := <-m.Broadcast:
			var msg struct {
				ConversationID string `json:"conversation_id"`
				Content        string `json:"content"`
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

func (m *WSManager) unregister(client *Client) {
	m.mutex.Lock()
	currentClient, ok := m.Clients[client.UserID]
	if ok && currentClient == client {
		delete(m.Clients, client.UserID)
		client.closeSend()
		log.Printf("Client disconnected: %s", client.UserID)
	}
	m.mutex.Unlock()

	if ok && currentClient == client {
		// 更新用户离线状态
		if err := m.sessionService.UpdateUserStatus(context.Background(), client.UserID, false); err != nil {
			log.Printf("Failed to update user status: %v", err)
		}
	}
}

func (m *WSManager) unregisterAsync(client *Client) {
	select {
	case m.Unregister <- client:
	case <-m.done:
	case <-m.stop:
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

	if client.enqueue(message) {
		return true
	}

	log.Printf("Client send buffer full or closed, dropping direct delivery for %s", userID)
	return false
}

func (m *WSManager) startRedisSubscriber() {
	if m.redisClient == nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.redisCancel = cancel

	pubsub := m.redisClient.Subscribe(ctx, wsPushChannel)
	log.Printf("Redis subscriber listening on channel: %s", wsPushChannel)
	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}

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
			case <-m.stop:
				return
			}
		}
	}()
}

// SendMessage 发送消息给指定用户
func (m *WSManager) SendMessage(userID string, message []byte) error {
	m.mutex.RLock()
	client, ok := m.Clients[userID]
	m.mutex.RUnlock()

	if !ok {
		return nil
	}
	if !client.enqueue(message) {
		return fmt.Errorf("client send buffer full or closed for %s", userID)
	}
	return nil
}

// Shutdown 关闭 WebSocket 管理器
func (m *WSManager) Shutdown() {
	log.Println("Starting WebSocket manager shutdown...")

	m.stopOnce.Do(func() {
		close(m.stop)
		if m.redisCancel != nil {
			m.redisCancel()
		}
	})

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 关闭所有客户端连接
	for _, client := range m.Clients {
		log.Printf("Closing connection for client: %s", client.UserID)
		client.closeSend()
		client.Conn.Close()
	}

	// 清空客户端映射
	m.Clients = make(map[string]*Client)

	log.Println("WebSocket manager shutdown completed")
}

// WritePump 处理向客户端写入消息
func (c *Client) WritePump(manager *WSManager) {
	ticker := time.NewTicker(manager.pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(manager.writeWait))
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
			_ = c.Conn.SetWriteDeadline(time.Now().Add(manager.writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-manager.stop:
			return
		}
	}
}

// ReadPump 处理从客户端读取消息
func (c *Client) ReadPump(manager *WSManager) {
	defer func() {
		manager.unregisterAsync(c)
		c.Conn.Close()
	}()

	// 设置读取超时，防止阻塞太久
	_ = c.Conn.SetReadDeadline(time.Now().Add(manager.pongWait))

	// 设置PongHandler，接收到Pong时重置读取超时
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(manager.pongWait))
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
			if !c.enqueue(responseBytes) {
				log.Printf("Client send buffer full, dropping pong for %s", c.UserID)
			}
			// 重置读取超时并继续等待下一条消息
			_ = c.Conn.SetReadDeadline(time.Now().Add(manager.pongWait))
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

		_ = dbMessage
	}
}

func (c *Client) closeSend() {
	c.closeOnce.Do(func() {
		close(c.Send)
	})
}

func (c *Client) enqueue(message []byte) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()

	select {
	case c.Send <- message:
		return true
	default:
		return false
	}
}
