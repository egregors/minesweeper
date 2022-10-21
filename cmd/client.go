package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	g "github.com/egregors/minesweeper/pkg"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type clientState int

const (
	INIT clientState = iota
	GAME
)

type Client struct {
	serverAddr string
	conn       net.Conn

	game *g.Game
	ui   tea.Program

	state clientState

	mu sync.Mutex
}

func NewClient() *Client {
	return &Client{
		// TODO: take it from outside
		serverAddr: "ws://127.0.0.1:8080/",
	}
}

func (c *Client) updateGame(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	g.FromGob(data, &c.game)
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

	// TODO: when should i close the Conn?
}

func (c *Client) pullServerEvents() {
	for {
		switch c.state {
		case GAME:
			msg, _, err := wsutil.ReadServerData(c.conn)
			if err != nil {
				fmt.Printf("Can't receive data: %s", err.Error())
				continue
			}
			// TODO: mu?
			// TODO: sent noop to update UI
			c.updateGame(msg)
		}
	}
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
			// TODO: client should wait messages from the Server in separated goroutine
			switch c.state {
			case INIT:
				// hi server message
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

				// start game
				c.state = GAME

				// pull game update form the server
				go c.pullServerEvents()

				// run ui
				c.updateGame(msg)
				err = RunCli(&c.game.M, c.conn)
				if err != nil {
					// TODO: don't panic, bro
					panic(err)
				}
				return nil
			}

		}
	}
}
