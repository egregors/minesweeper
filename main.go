package main

import (
	"fmt"
	"github.com/egregors/minesweeper/cmd"
	g "github.com/egregors/minesweeper/pkg"
	"github.com/jessevdk/go-flags"
	"os"
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

	if opts.Srv {
		game := g.NewGame(g.EASY, true)
		if err := cmd.NewServer(game, opts.Dbg).Run(); err != nil {
			panic(err)
		}
		return
	}

	if err := cmd.NewClient().Run(); err != nil {
		panic(err)
	}
}
