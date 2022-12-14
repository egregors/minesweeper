package main

import (
	"fmt"
	"log"
	"os"

	"github.com/egregors/minesweeper/cmd"
	g "github.com/egregors/minesweeper/pkg"
	"github.com/jessevdk/go-flags"
)

type Opts struct {
	Dbg bool `long:"dbg" env:"DEBUG" description:"Debug mode"`
	// TODO: rework CLI
	Srv bool `short:"d"`
}

func main() {
	var opts Opts
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		//nolint:errorlint // it's
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("cli error: %v", err)
		}
		os.Exit(2)
	}

	// setup logger
	logger := g.NewLogger()
	log.SetOutput(logger)

	if opts.Srv {
		game := g.NewGame(g.EASY, true)
		if err := cmd.NewServer(game, logger, opts.Dbg).Run(); err != nil {
			panic(err)
		}
		return
	}

	if err := cmd.NewClient("ws://127.0.0.1:8080", logger, opts.Dbg).Run(); err != nil {
		panic(err)
	}
}
