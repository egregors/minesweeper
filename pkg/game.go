package game

import (
	"fmt"
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

func (g Game) OpenCell(p Point) {
	m := g.M
	mine := m.Mines[p[0]][p[1]]
	switch mine {
	case MINE:
		g.M.State = OVER
	case ZERO:
		var openCell func(r, c int)
		openCell = func(r, c int) {
			if m.Field[r][c] == EMPTY {
				return
			}

			if m.Mines[r][c] != ZERO {
				m.Field[r][c] = m.Mines[r][c]
				return
			}

			m.Field[r][c] = EMPTY
			m.LeftToOpen--

			dirs := [][]int{
				{-1, -1}, {-1, 0}, {-1, 1},
				{0, -1}, {0, 1},
				{1, -1}, {1, 0}, {1, 1},
			}
			for _, d := range dirs {
				newR, newC := r+d[0], c+d[1]
				if newR >= 0 && newR < m.N && newC >= 0 && newC < m.M {
					openCell(newR, newC)
				}
			}
		}
		openCell(p[0], p[1])
		if m.LeftToOpen == 0 {
			m.State = WIN
			for r := 0; r < m.N; r++ {
				for c := 0; c < m.M; c++ {
					if m.Field[r][c] == HIDE {
						m.Field[r][c] = m.Mines[r][c]
					}
				}
			}
		}
	default:
		m.Field[p[0]][p[1]] = mine
	}

	// TODO: send new field state to ALL clients

}

func (g Game) Bytes() []byte {
	return ToGob(g)
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
