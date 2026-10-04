package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"strings"
)

const serverAddr = "127.0.0.1:8443"

func main() {
	fmt.Println("Агент диспетчеризації (mTLS Client)")

	clientCert, err := tls.LoadX509KeyPair(
		"certs/client.crt",
		"certs/client.key",
	)
	if err != nil {
		log.Fatalf("load client certificate: %v", err)
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
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS13,
		ServerName:   "localhost",
	}

	fmt.Println("[КЛІЄНТ] Завантажено сертифікат: certs/client.crt")
	fmt.Printf("[КЛІЄНТ] Підключення до шлюзу %s...\n", serverAddr)

	conn, err := tls.Dial("tcp", serverAddr, tlsConfig)
	if err != nil {
		log.Fatalf("TLS connection failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("[КЛІЄНТ] mTLS-рукостискання успішне. Використано TLS 1.3")

	incident := "CRITICAL_ERROR: Database connection timeout on node-04"

	fmt.Printf("[КЛІЄНТ] Відправка алерту: %s\n", incident)

	if _, err := fmt.Fprintln(conn, incident); err != nil {
		log.Fatalf("send incident: %v", err)
	}

	reader := bufio.NewReader(conn)

	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("read response: %v", err)
	}

	fmt.Printf("[КЛІЄНТ] Відповідь від сервера: %s\n", strings.TrimSpace(response))
	fmt.Println("[КЛІЄНТ] Сесію успішно завершено.")
}
