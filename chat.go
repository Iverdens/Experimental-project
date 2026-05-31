package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const maxClients = 10

type Client struct {
	Name string
	Conn net.Conn
}

type ClientManager struct {
	clients  map[net.Conn]*Client
	mu       sync.RWMutex
	messages []string
}

// В chat.go
func (m *ClientManager) KickAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for conn := range m.clients {
		conn.Write([]byte("\n[SECURITY ALERT] CHANNEL COMPROMISED! ZEROIZATION TRIGGERED. DISCONNECTING...\n"))
		conn.Close() // Закрываем сокет
	}
	m.clients = make(map[net.Conn]*Client) // Очищаем список
	m.messages = []string{}                // Стираем историю
}

func (cm *ClientManager) ClearHistory() {
	cm.messages = []string{} // Очищаем слайс с сообщениями
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:  make(map[net.Conn]*Client),
		messages: []string{},
	}
}

func (m *ClientManager) AddClient(conn net.Conn, name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.clients) >= maxClients {
		return false
	}

	m.clients[conn] = &Client{
		Name: name,
		Conn: conn,
	}

	return true
}

func (m *ClientManager) RemoveClient(conn net.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, conn)
}

func (m *ClientManager) Broadcast(sender net.Conn, message string, senderName string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for conn := range m.clients {
		if conn != sender { // Не отправляем сообщение отправителю
			// Отправляем только сообщение
			_, err := conn.Write([]byte(message))
			if err != nil {
				m.RemoveClient(conn)
			}
			// Отправляем приглашение
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			prompt := fmt.Sprintf("[%s][%s]: ", timestamp, m.clients[conn].Name)
			conn.Write([]byte(prompt))
		}
	}
}

func (m *ClientManager) StoreMessage(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *ClientManager) SendHistory(conn net.Conn) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, msg := range m.messages {
		conn.Write([]byte(msg))
	}
}

/* фикс гонки бродкаста на всякий случай
func (m *ClientManager) Broadcast(sender net.Conn, message string, senderName string) {
	m.mu.RLock()

	var conns []net.Conn
	for conn := range m.clients {
		if conn != sender {
			conns = append(conns, conn)
		}
	}

	m.mu.RUnlock()

	for _, conn := range conns {
		conn.Write([]byte(message))

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		prompt := fmt.Sprintf("[%s][%s]: ", timestamp, senderName)
		conn.Write([]byte(prompt))
	}
}*/
