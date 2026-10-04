package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"secure-alert-gateway/email"
)

const (
	listenAddr = ":8443"
	smtpAddr   = "127.0.0.1:1025"
	from       = "server@example.com"
	to         = "duty-engineer@internal.net"
)

func main() {
	cert, err := tls.LoadX509KeyPair(
		"certs/server.crt",
		"certs/server.key",
	)
	if err != nil {
		log.Fatalf("load server certificate: %v", err)
	}

	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("read CA certificate: %v", err)
	}

	caPool := x509.NewCertPool()

	if !caPool.AppendCertsFromPEM(caCert) {
		log.Fatal("failed to add CA certificate to pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS13,
	}

	listener, err := tls.Listen("tcp", listenAddr, tlsConfig)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	fmt.Printf(
		"=== Захищений шлюз критичних сповіщень (mTLS: %s, SMTP: %s) ===\n",
		listenAddr,
		smtpAddr,
	)

	fmt.Println("[СЕРВЕР] Пул CA завантажено: certs/ca.crt")
	fmt.Printf("[СЕРВЕР] Слухач mTLS відкрито на порту %s\n", listenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("[ПІДКЛЮЧЕННЯ] Нове mTLS-з'єднання від %s\n", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	message, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("[ПОВІДОМЛЕННЯ] Не вдалося прочитати повідомлення: %v", err)
		return
	}

	incident := strings.TrimSpace(message)

	if incident == "" {
		log.Println("[ПОВІДОМЛЕННЯ] Порожнє повідомлення")
		return
	}

	registeredAt := time.Now()

	fmt.Printf("[ПОВІДОМЛЕННЯ] Отримано інцидент: %s\n", incident)

	emailBody := fmt.Sprintf(
		"Критичне сповіщення\n\nЧас реєстрації: %s\nІнцидент: %s\n",
		registeredAt.Format("2006-01-02 15:04:05"),
		incident,
	)

	err = email.Send(
		smtpAddr,
		from,
		to,
		"CRITICAL ALERT",
		emailBody,
	)

	if err != nil {
		log.Printf("[SMTP] Не вдалося надіслати лист: %v", err)

		_, _ = fmt.Fprintln(
			conn,
			"[SERVER]: ALERT RECEIVED, EMAIL FAILED",
		)

		return
	}

	fmt.Printf(
		"[SMTP] Лист успішно відправлено на %s через %s\n",
		to,
		smtpAddr,
	)

	response := "[SERVER]: ALERT RECEIVED AND VERIFIED"

	if _, err := fmt.Fprintln(conn, response); err != nil {
		log.Printf("[ВІДПОВІДЬ] Не вдалося відправити відповідь: %v", err)
		return
	}

	fmt.Printf("[ВІДПОВІДЬ] %s\n", response)
}
