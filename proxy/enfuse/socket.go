package enfuse

import (
	"fmt"
	"github.com/enfabrica/enkit/lib/logger"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net"
)

var _ io.ReadWriter = &SocketShim{}

// SocketShim is a simple wrapper that implements io.ReadWriter that writes and reads the full buffer while translating
// payloads. If read from, it will strip the Prefix from the payload if it is present. If it is written to, it will
// automatically append the Prefix.
// in the future this could be nicely re-connectable with a buffer window.
type SocketShim struct {
	WebConn   *WebsocketLocker
	Prefix    []byte
	PrefixLen int
}

func (s SocketShim) Read(p []byte) (n int, err error) {
	_, data, err := s.WebConn.ReadMessage()
	if err != nil {
		return 0, err
	}
	copy(p, data[s.PrefixLen:])
	return 0, io.EOF
}

func (s SocketShim) Write(p []byte) (n int, err error) {
	// websockets always write the full buffer
	realPayload := append(s.Prefix, p...)
	return len(p), s.WebConn.WriteMessage(websocket.BinaryMessage, realPayload)
}

func NewSocketShim(start PayloadAppendStrategy, Conn *websocket.Conn) (*SocketShim, error) {
	pLen, f := start()
	pre, err := f()
	if err != nil {
		return nil, err
	}
	return &SocketShim{PrefixLen: pLen, Prefix: pre, WebConn: NewWebsocketLock(Conn)}, nil
}

func NewWebsocketTCPShim(strat PayloadAppendStrategy, lis net.Listener, web *websocket.Conn) *WebsocketTCPShim {
	l, _ := strat()
	wShim := &WebsocketTCPShim{
		clientMap: map[string]net.Conn{},
		prefixLen: l,
		netConn:   lis,
		webConn:   NewWebsocketLock(web),
		l:         logger.DefaultLogger{Printer: log.Printf},
	}
	go wShim.handleWrites()
	return wShim
}

// WebsocketTCPShim manages a duplex connection from a single websocket.Conn to a single net.Listener
// the id of the duplexed connection is determined by the initial PayloadAppendStrategy.
type WebsocketTCPShim struct {
	webConn *WebsocketLocker
	netConn net.Listener

	clientMap map[string]net.Conn
	prefixLen int

	l logger.Logger
}

// handleWrites will handle all writes from webConn to netConn. It reads the content of the payload based on
func (w *WebsocketTCPShim) handleWrites() {
	for {
		_, data, err := w.webConn.ReadMessage()
		if err != nil {
			w.l.Warnf("Error reading from websocket %v", err)
			continue
		}
		uid := data[:w.prefixLen]
		remainingData := data[w.prefixLen:]
		clientConn := w.clientMap[string(uid)]
		if clientConn == nil {
			clientConn, err = net.Dial(w.netConn.Addr().Network(), w.netConn.Addr().String())
			if err != nil {
				w.l.Warnf("Error dialing to existing listener %v", err)
				continue
			}
			w.clientMap[string(uid)] = clientConn
			go w.handleReadToWebsocket(clientConn, uid)
		}
		_, err = clientConn.Write(remainingData)
		if err != nil {
			w.l.Warnf("Error writing to client %v", err)
			continue
		}
	}
}

// handleRead will read from the net.Conn and write the full payload to webConn
func (w *WebsocketTCPShim) handleReadToWebsocket(c net.Conn, uid []byte) {
	go func() {
		if _, err := io.Copy(SocketShim{WebConn: w.webConn, Prefix: uid, PrefixLen: len(uid)}, c); err != nil {
			fmt.Println(err, "err copying for client")
		}
	}()
}

//TODO(adam): decide whether or not it's even necessary to have websocketTCPShim be a struct.
//entirely possible for it to just be a function
func (w *WebsocketTCPShim) Close() error  {
	return nil
}