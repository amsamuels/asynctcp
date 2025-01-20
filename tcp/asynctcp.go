// Improved AsyncTCPServer Implementation in Go
package tcp

import (
	"errors"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	defaultMaxPacketSize = 1024 << 10 // 1 MB
	readChanSize         = 100
	writeChanSize        = 100
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "", log.Lshortfile)
}

// AsyncTCPServer manages the lifecycle of a TCP server.
type AsyncTCPServer struct {
	callback CallBack
	protocol Protocol

	tcpAddr  string
	listener *net.TCPListener

	readDeadline  time.Duration
	writeDeadline time.Duration

	exitChan chan struct{}

	bucket     *TCPConnBucket
	shutdownWg sync.WaitGroup
}

// NewAsyncTCPServer creates a new AsyncTCPServer instance.
func NewAsyncTCPServer(callback CallBack, protocol Protocol) *AsyncTCPServer {
	return &AsyncTCPServer{
		callback: callback,
		protocol: protocol,
		bucket:   NewTCPConnBucket(),
		exitChan: make(chan struct{}),
	}
}

// ListenAndServe starts the server and listens for incoming connections.
func (srv *AsyncTCPServer) Run(port string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		return err
	}

	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	srv.listener = ln
	logger.Println("Server started, listening on", srv.tcpAddr)
	go srv.serve()
	return nil
}

// serve handles incoming connections and manages server operations.
func (srv *AsyncTCPServer) serve() {
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Server panic: %v", r)
		}
		srv.listener.Close()
	}()

	// Periodically clean up closed connections.
	go func() {
		for {
			select {
			case <-srv.exitChan:
				return
			default:
				srv.bucket.removeClosedTCPConn()
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	for {
		select {
		case <-srv.exitChan:
			logger.Println("Server shutting down...")
			return
		default:
			conn, err := srv.listener.AcceptTCP()
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					logger.Println("Temporary accept error, retrying:", err)
					time.Sleep(50 * time.Millisecond)
					continue
				}
				logger.Println("Accept error:", err)
				return
			}

			srv.shutdownWg.Add(1)
			go srv.handleConnection(conn)
		}
	}
}

// handleConnection sets up a new TCP connection and manages its lifecycle.
func (srv *AsyncTCPServer) handleConnection(conn *net.TCPConn) {
	defer srv.shutdownWg.Done()

	tcpConn := srv.newTCPConn(conn, srv.callback, srv.protocol)
	tcpConn.setReadDeadline(srv.readDeadline)
	tcpConn.setWriteDeadline(srv.writeDeadline)
	srv.bucket.Put(tcpConn.GetRemoteAddr().String(), tcpConn)
}

func (srv *AsyncTCPServer) newTCPConn(conn *net.TCPConn, callback CallBack, protocol Protocol) *TCPConn {
	if callback == nil {
		callback = srv.callback
	}
	if protocol == nil {
		protocol = srv.protocol
	}

	c := NewTCPConn(conn, callback, protocol)
	c.Serve()
	return c
}

// Close gracefully shuts down the server and closes all connections.
func (srv *AsyncTCPServer) Close() {
	close(srv.exitChan)
	logger.Println("Waiting for active connections to close...")
	srv.shutdownWg.Wait()

	for _, conn := range srv.bucket.GetAll() {
		if !conn.IsClosed() {
			conn.Close()
		}
	}

	logger.Println("Server shutdown complete.")
}

// SetReadDeadline configures the read deadline for connections.
func (srv *AsyncTCPServer) SetReadDeadline(d time.Duration) {
	srv.readDeadline = d
}

// SetWriteDeadline configures the write deadline for connections.
func (srv *AsyncTCPServer) SetWriteDeadline(d time.Duration) {
	srv.writeDeadline = d
}
