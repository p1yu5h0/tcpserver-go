package main

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

var requestCount int
var mu sync.Mutex

var maxThreads = 10
var threadPool = make(chan struct{}, maxThreads)

func do(conn net.Conn) {
	mu.Lock()
	requestCount++
	currentRequest := requestCount
	mu.Unlock()

	var buf []byte = make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Request %d: processing the request\n", currentRequest)
	time.Sleep(8 * time.Second)

	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nhello, piyush!\r\n"))
	log.Printf("Request %d: processing done\n", currentRequest)
	conn.Close()

	<-threadPool
}

func main() {
	listener, err := net.Listen("tcp", ":1729")
	if err != nil {
		log.Fatal(err)
	}

	backlog := 100
	listener = limitListener(listener, backlog)

	for {
		log.Println("waiting for client to get connected")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("client get connected")

		threadPool <- struct{}{}
		go do(conn)
	}
}

func limitListener(l net.Listener, backlog int) net.Listener {
	switch v := l.(type) {
	case *net.TCPListener:
		return &keepAliveListener{TCPListener: v}
	default:
		log.Fatalf("unsupported listener: %T", l)
	}
	return l
}

type keepAliveListener struct {
	*net.TCPListener
}

func (kl *keepAliveListener) Accept() (net.Conn, error) {
	conn, err := kl.TCPListener.Accept()
	if err != nil {
		return nil, err
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Println("Error: Connection is not a TCP connection")
		conn.Close()
		return nil, errors.New("non-TCP connection")
	}

	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(30 * time.Second)

	return conn, nil
}