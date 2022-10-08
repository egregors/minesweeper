package cmd

import (
	"context"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"math/rand"
	"net"
	"time"
)

type Client struct {
	serverAddr string
	conn       net.Conn

	gameData string

	//UI tea.Program
}

func NewClient() *Client {
	return &Client{
		// TODO: take it from outside
		serverAddr: "ws://127.0.0.1:8080/",
	}
}

func (c *Client) connect() error {
	if c.conn == nil {
		conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), c.serverAddr)
		if err != nil {
			return fmt.Errorf("can't conntect to the server: %e", err)
		}
		c.conn = conn
	}
	return nil

	// TODO: when should i close the conn?
}

func (c *Client) Run() error {
	log.Println("Client started")

	// connection retry loop
	for {
		if err := c.connect(); err != nil {
			log.Printf("sleep before retry: %s", err.Error())
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		log.Println("connected!")

		// game loop
		for {
			log.Printf("STATE: %s", c.gameData)

			// need to INIT game
			// TODO: make if through fsm
			if c.gameData == "" {
				// hi server messsage
				if err := wsutil.WriteClientMessage(c.conn, ws.OpText, nil); err != nil {
					fmt.Println("Cannot send: " + err.Error())
					continue
				}

				// get game data
				msg, _, err := wsutil.ReadServerData(c.conn)
				if err != nil {
					fmt.Println("Cannot receive data: " + err.Error())
					continue
				}
				gameData := string(msg)
				c.gameData = gameData

				continue
			}

			// waiting for update
			msg, _, err := wsutil.ReadServerData(c.conn)
			if err != nil {
				fmt.Println("Cannot receive data: " + err.Error())
				continue
			}
			log.Println("got update ", string(msg))
		}
	}
}

func genClientID() string {
	// FIXME: it should be uniq UUID, not just a random int
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("id_%d", rnd.Intn(100))
}
