package cmd

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net/http"
	"strings"
)

type player struct {
	addr     string
	isOnline bool
}

type players map[string]*player

func (ps players) getAmount() int { return len(ps) }

func (ps players) add(player *player) {
	if p, ok := ps[player.addr]; ok {
		// reconnect
		p.isOnline = true
	} else {
		// new player
		ps[player.addr] = player
	}
}

func (ps players) disconnect(addr string) {
	for k, v := range ps {
		if v.addr == addr {
			ps[k].isOnline = false
			return
		}
	}
}

type Srv struct {
	ps players
}

func (s Srv) String() string {
	ls := []string{"\n"}
	for k, v := range s.ps {
		status := "ONLINE"
		if !v.isOnline {
			status = "OFFLINE"
		}
		ls = append(ls, fmt.Sprintf("%s:%s", k, v.addr, status))
	}
	return strings.Join(ls, "\n")
}

func NewServer() *Srv {
	return &Srv{
		ps: make(players),
	}
}

func (s Srv) Run() error {
	fmt.Println("Server started, waiting for connection from players...")
	return http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Println("Error starting socket server: " + err.Error())
		}

		addr := conn.RemoteAddr().String()
		log.Printf("Client %s connected", addr)

		go func() {
			defer func() { _ = conn.Close() }()
			for {
				// TODO: remove
				log.Println(s.String())

				// GET msg
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					log.Println("Error receiving data: " + err.Error())
					log.Printf("Client %s disconnected", addr)
					s.ps.disconnect(addr)
					return
				}

				switch op {
				case ws.OpText:
					// join \ rejoin player
					s.ps.add(&player{
						addr:     addr,
						isOnline: true,
					})

					if err := wsutil.WriteServerMessage(conn, ws.OpText, []byte("GAME DATA")); err != nil {
						log.Println("Error sending data: " + err.Error())
						log.Println("Client disconnected")
						s.ps.disconnect(addr)
						return
					}

				case ws.OpBinary:
					log.Println("Receive bin data....", msg)
				}
			}
		}()
	}))
}
