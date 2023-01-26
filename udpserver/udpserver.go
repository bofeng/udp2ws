package udpserver

import (
	"log"
	"net"

	"github.com/gofiber/websocket/v2"
)

type UDPDataType string

const (
	UDPDataText   UDPDataType = "text"
	UDPDataBinary UDPDataType = "binary"
	bufSize       int         = 1024 * 1024
)

type WSConn struct {
	ID      string
	Conn    *websocket.Conn
	ErrChan chan error
}

type UPDServer struct {
	listenAddr    string
	udpAddr       *net.UDPAddr
	dataType      UDPDataType
	allWSConn     map[string]WSConn
	addWSConnChan chan WSConn
	delWSConnChan chan string
}

func NewUDPServer(listenAddr string, dataType UDPDataType) *UPDServer {
	udpAddr, err := net.ResolveUDPAddr("udp4", listenAddr)
	if err != nil {
		log.Fatalln("Invalid UDP listen address:", listenAddr, "error:", err)
	}

	if err != nil {
		log.Fatalln(err)
	}
	return &UPDServer{
		listenAddr:    listenAddr,
		udpAddr:       udpAddr,
		dataType:      dataType,
		allWSConn:     make(map[string]WSConn),
		addWSConnChan: make(chan WSConn, 1),
		delWSConnChan: make(chan string, 1),
	}
}

func (s *UPDServer) Run() {
	udpConn, err := net.ListenUDP("udp", s.udpAddr)
	if err != nil {
		log.Fatalln("Failed to start UDP server, error:", err)
	}
	defer udpConn.Close()

	wsMsgType := websocket.TextMessage
	if s.dataType == UDPDataBinary {
		wsMsgType = websocket.BinaryMessage
	}

	udpDataChan := make(chan []byte, 1)

	go func() {
		buf := make([]byte, bufSize)
		for {
			n, err := udpConn.Read(buf)
			if err != nil {
				log.Fatalln("Failed to read UDP data:", err)
			}
			// buf will be used for next read, so here we
			// send a data copy to the channel to consume
			bufCopy := make([]byte, n)
			copy(bufCopy, buf)
			udpDataChan <- bufCopy
		}
	}()

	for {
		select {
		case data := <-udpDataChan:
			for id, conn := range s.allWSConn {
				err = conn.Conn.WriteMessage(wsMsgType, data)
				if err != nil {
					delete(s.allWSConn, id)
					conn.ErrChan <- err
				}
			}
		case conn := <-s.addWSConnChan:
			s.allWSConn[conn.ID] = conn
		case connID := <-s.delWSConnChan:
			delete(s.allWSConn, connID)
		}
	}
}

func (s *UPDServer) AddWSConn(conn WSConn) {
	s.addWSConnChan <- conn
}

func (s *UPDServer) DelWSConn(connID string) {
	s.delWSConnChan <- connID
}
