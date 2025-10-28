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
	Server bool `short:"s" long:"server" description:"Run as server"`
	Client bool `short:"c" long:"client" description:"Run as client"`
	Addr   string `short:"a" long:"addr" default:"127.0.0.1:8080" description:"Server address (for client mode) or bind address (for server mode)"`
	Dbg    bool `long:"debug" env:"DEBUG" description:"Enable debug mode"`
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

	// If neither server nor client is specified, default to client mode
	if !opts.Server && !opts.Client {
		opts.Client = true
	}

	if opts.Server {
		game := g.NewGame(g.EASY, true)
		if err := cmd.NewServer(game, logger, opts.Dbg).Run(); err != nil {
			panic(err)
		}
		return
	}

	if opts.Client {
		serverAddr := "ws://" + opts.Addr
		if err := cmd.NewClient(serverAddr, logger, opts.Dbg).Run(); err != nil {
			panic(err)
		}
	}
}
