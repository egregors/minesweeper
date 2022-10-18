package cmd

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	g "github.com/egregors/minesweeper/pkg"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net"
	"time"
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

	// TODO: when should i close the Conn?
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
				g.FromGob(msg, &c.game)

				// run ui
				err = RunCli(&c.game.M, c.conn)
				if err != nil {
					// TODO: don't panic, bro
					panic(err)
				}
				return nil

			case GAME:
				msg, _, err := wsutil.ReadServerData(c.conn)
				if err != nil {
					fmt.Println("Cannot receive data: " + err.Error())
					continue
				}
				log.Println("got update ", string(msg))
			}
 
		}
	}
}
