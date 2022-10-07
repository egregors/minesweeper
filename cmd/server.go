package cmd

import (
	"net"

	game "github.com/egregors/minesweeper/pkg"
)

type Srv struct {
	g    game.Game
	conn net.Conn
}

func NewServer() *Srv {
	return nil
}
