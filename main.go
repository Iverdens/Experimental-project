package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

var clientManager = NewClientManager()

func main() {
	printLocalIP()
	args := os.Args[1:]

	if len(args) > 1 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	port := PORT
	if len(args) == 1 {
		port = args[0]
		if _, err := strconv.Atoi(port); err != nil {
			fmt.Println("[USAGE]: ./TCPChat $port")
			return
		}
	}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Listening on the port :" + port)

	// ЗАПУСКАЕМ ФОНОВЫЙ МОНИТОРИНГ
	go func() {
		for {
			if !performQuantumAuth() {
				fmt.Println("\n[!!!] ALERT: Zeroizing and disconnecting all users!")
				clientManager.KickAll()
				// Можно добавить небольшую паузу, чтобы не спамить в консоль
				time.Sleep(2 * time.Second)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	if !performQuantumAuth() {
		conn.Write([]byte("[SECURITY ALERT] Quantum channel intercepted! Zeroization triggered. Access denied.\n"))
		conn.Close()
		return
	}

	defer conn.Close()

	conn.Write([]byte(WelcomMessage))

	scanner := bufio.NewScanner(conn)
	var name string

	if scanner.Scan() {
		if !performQuantumAuth() {
			conn.Write([]byte("\n[SECURITY ALERT] CHANNEL COMPROMISED! ZEROIZATION TRIGGERED. DISCONNECTING...\n"))
			return
		}

		name = scanner.Text()
		if name == "" {
			conn.Write([]byte("Name cannot be empty.\n"))
			return
		}
	} else {
		return
	}

	if !clientManager.AddClient(conn, name) {
		conn.Write([]byte("Chat is full. Maximum number of users reached.\n"))
		return
	}

	defer clientManager.RemoveClient(conn)

	fmt.Printf("Client connected: %s\n", name)
	defer fmt.Printf("Client disconnected: %s\n", name)

	// Отправляем историю сообщений новому клиенту
	clientManager.SendHistory(conn)

	// Сообщаем другим о подключении клиента
	joinMsg := fmt.Sprintf("%s has joined our chat...\n", name)
	clientManager.StoreMessage(joinMsg)
	clientManager.Broadcast(conn, "\n"+joinMsg, name) // Не отправляем новому клиенту

	// Отправляем приглашение для ввода
	sendPrompt(conn, name)

	for scanner.Scan() {
		if !performQuantumAuth() {
			clientManager.ClearHistory()
			conn.Write([]byte("\n[SECURITY ALERT] SYSTEM COMPROMISED! DATA ZEROIZED!\n"))
			return // Выход из цикла и закрытие соединения
		}

		msg := scanner.Text()
		if msg == "" {
			// При пустом вводе отправляем пустую строку и приглашение
			conn.Write([]byte("\n"))
			sendPrompt(conn, name)
			continue
		}

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		formatted := fmt.Sprintf("[%s][%s]: %s\n", timestamp, name, msg)

		clientManager.StoreMessage(formatted)
		clientManager.Broadcast(conn, "\n"+formatted, name) // Не отправляем отправителю

		// Отправляем пустую строку и приглашение после обработки сообщения
		// conn.Write([]byte("\n"))
		sendPrompt(conn, name)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from %s: %v\n", name, err)
	}

	leftMsg := fmt.Sprintf("%s has left our chat...\n", name)
	clientManager.StoreMessage(leftMsg)
	clientManager.Broadcast(conn, "\n"+leftMsg, name)
}

// sendPrompt отправляет клиенту приглашение для ввода
func sendPrompt(conn net.Conn, name string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prompt := fmt.Sprintf("[%s][%s]: ", timestamp, name)
	conn.Write([]byte(prompt))
}

func printLocalIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error getting network interfaces:", err)
		return
	}
	fmt.Println("Your local IP addresses:")
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println("- ", ipnet.IP.String())
			}
		}
	}
}

func performQuantumAuth() bool {
	data, err := os.ReadFile("security_status.txt")
	if err != nil {
		return false // Файла нет -> доступ запрещен
	}

	return string(data) == "1"
}
