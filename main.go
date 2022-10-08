package main

import (
	"fmt"
	"github.com/egregors/minesweeper/cmd"
	"github.com/jessevdk/go-flags"
	"os"
)

type Opts struct {
	Dbg bool `long:"dbg" env:"DEBUG" description:"Debug mode"`
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

	// TODO: single player?
	//g := game.NewGame(game.EASY, opts.Dbg)
	//if err := cmd.RunCli(g.GetModel()); err != nil {
	//	panic(err)
	//}

	if opts.Srv {
		if err := cmd.NewServer().Run(); err != nil {
			panic(err)
		}
		return
	}

	if err := cmd.NewClient().Run(); err != nil {
		panic(err)
	}
}
