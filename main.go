package main

import (
	"flag"
	"log"
	"strconv"
	"time"
	"udp2ws/udpserver"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

var udpServer *udpserver.UPDServer

func main() {
	wsAddrPtr := flag.String("wsaddr", ":6080", "websocket listen address")
	udpAddrPtr := flag.String("udpaddr", ":6081", "udp server listen address")
	dataTypePtr := flag.String(
		"data",
		"text",
		"udp data type: text or binary",
	)
	flag.Parse()

	if wsAddrPtr == nil || *wsAddrPtr == "" {
		log.Fatalln("Missing websocket address parameter. Use -h to help")
	}
	if udpAddrPtr == nil || *udpAddrPtr == "" {
		log.Fatalln("Missing udp server address parameter. Use -h to help")
	}

	dataType := udpserver.UDPDataType(*dataTypePtr)
	if dataType != udpserver.UDPDataText &&
		dataType != udpserver.UDPDataBinary {
		log.Fatalln("Unsupported value for data parameter. Use -h to help")
	}

	log.Println("* Websocket listen on:", *wsAddrPtr)
	log.Println("* UDP server listen on:", *udpAddrPtr)
	log.Println("* UDP server data type:", *dataTypePtr)

	udpServer = udpserver.NewUDPServer(*udpAddrPtr, dataType)
	go udpServer.Run()

	app := fiber.New(fiber.Config{
		Immutable: true,
	})

	app.Use(logger.New())
	app.Get(
		"/",
		wsCheckMiddleware(*udpAddrPtr, *dataTypePtr),
		websocket.New(wsHandler),
	)
	app.Listen(*wsAddrPtr)
}

func wsCheckMiddleware(backendURL string, dataType string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}
		return c.Next()
	}
}

func wsHandler(c *websocket.Conn) {
	clientID := strconv.FormatUint(uint64(time.Now().UnixMicro()), 36)
	defer func() {
		udpServer.DelWSConn(clientID)
		c.Close()
		log.Println("=\\= client", clientID, "disconnected")
	}()

	log.Println("==> client", clientID, "connected")

	errChan := make(chan error, 1)
	wsConn := udpserver.WSConn{
		ID:      clientID,
		Conn:    c,
		ErrChan: errChan,
	}
	udpServer.AddWSConn(wsConn)

	go func() {
		for {
			_, _, err := c.ReadMessage()
			// if client closed, send message to error channel
			if err != nil {
				errChan <- err
			}
		}
	}()

	err := <-errChan
	if websocket.IsUnexpectedCloseError(
		err,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived) {
		log.Println("Read data error:", err)
	}
}
