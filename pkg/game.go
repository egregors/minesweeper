package game

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type Difficulty int

const (
	EASY Difficulty = iota
	NORMAL
	HARD
)

const (
	GAME = iota
	OVER
	WIN

	HIDE  = '~'
	MINE  = '*'
	FLAG  = '!'
	GESS  = '?'
	BOOM  = 'X'
	EMPTY = ' '

	ZERO = '0'
)

type Game struct {
	M          Model
	Difficulty Difficulty

	dbg bool
}

func (g Game) ToGob() []byte {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(g)
	if err != nil {
		log.Printf("can't convert to gob: %s", err.Error())
	}
	return buf.Bytes()
}

func (g Game) getModel() Model {
	return g.M
}

func NewGame(difficulty Difficulty, dbg bool) *Game {
	var m Model

	switch difficulty {
	case EASY:
		m = NewModel(9, 9, 10, dbg)
	case NORMAL:
		m = NewModel(16, 16, 40, dbg)
	case HARD:
		m = NewModel(16, 30, 99, dbg)
	}

	return &Game{
		M:          m,
		Difficulty: difficulty,
		dbg:        dbg,
	}
}

type Point [2]int

func (p *Point) String() string {
	return fmt.Sprintf("[%d:%d]", p[0], p[1])
}

func (p *Point) FromGob(from []byte) {
	FromGob(from, p)
}

func (p *Point) ToGob() []byte {
	return ToGob(p)
}

type Model struct {
	Field, Mines [][]rune
	N, M         int
	LeftToOpen   int
	State        int

	Dbg bool
}

func NewModel(n, m, minesCount int, dbg bool) Model {
	var field, mines [][]rune
	field = make([][]rune, n)
	mines = make([][]rune, n)

	for i := 0; i < n; i++ {
		field[i] = make([]rune, m)
		mines[i] = make([]rune, m)
		for j := 0; j < m; j++ {
			mines[i][j] = '0'
		}
		for j := 0; j < m; j++ {
			field[i][j] = HIDE
		}
	}

	// setup Mines
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for minesCount > 0 {
		r, c := rnd.Intn(n), rnd.Intn(m)
		if mines[r][c] != MINE {
			mines[r][c] = MINE
			minesCount--
		}
	}
	// count Mines
	dirs := [][]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for r := 0; r < n; r++ {
		for c := 0; c < m; c++ {
			if mines[r][c] == MINE {
				for _, d := range dirs {
					newR, newC := r+d[0], c+d[1]
					if newR >= 0 && newR < n && newC >= 0 && newC < m {
						if mines[newR][newC] != MINE {
							mines[newR][newC]++
						}
					}
				}
			}
		}
	}

	// count shouldOpen \ empty cells
	var shouldOpen int
	for _, r := range mines {
		for _, c := range r {
			if c == ZERO {
				shouldOpen++
			}
		}
	}

	return Model{
		Field:      field,
		Mines:      mines,
		LeftToOpen: shouldOpen,
		N:          n,
		M:          m,
		Dbg:        dbg,
	}
}
