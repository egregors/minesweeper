package main

import (
	"fmt"
	"github.com/egregors/minesweeper/pkg"
	"github.com/jessevdk/go-flags"
	"os"
)

type Opts struct {
	Dbg bool `long:"dbg" env:"DEBUG" description:"Debug mode"`
}

func main() {
	var opts Opts
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("cli error: %v", err)
		}
		os.Exit(2)
	}

	g := game.NewGame(game.EASY, opts.Dbg)
	if err := g.Run(); err != nil {
		panic(err)
	}
}
