package main

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	HIDE = '~'
	MINE = '*'
	FLUG = '!'
	GESS = '?'
	BOOM = 'X'
)

type Point [2]int

type model struct {
	field, mines [][]rune
	curr         Point
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyUp:
			if m.curr[0] > 0 {
				m.curr[0]--
			}
		case tea.KeyDown:
			if m.curr[0] < len(m.field)-1 {
				m.curr[0]++
			}
		case tea.KeyLeft:
			if m.curr[1] > 0 {
				m.curr[1]--
			}
		case tea.KeyRight:
			if m.curr[1] < len(m.field[0])-1 {
				m.curr[1]++
			}

		case tea.KeySpace:
			c := m.field[m.curr[0]][m.curr[1]]
			switch c {
			case HIDE:
				m.field[m.curr[0]][m.curr[1]] = FLUG
			case FLUG:
				m.field[m.curr[0]][m.curr[1]] = GESS
			case GESS:
				m.field[m.curr[0]][m.curr[1]] = HIDE
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	// TODO: title should be relative to field wight
	frame := []string{
		"     *** Minesweeper ***",
		"     ===================",
	}
	for r := 0; r < 10; r++ {
		var line []rune
		for c := 0; c < 10; c++ {
			lo, hi := ' ', ' '
			if m.curr[0] == r && m.curr[1] == c {
				lo, hi = '[', ']'
			}
			line = append(line, lo, m.field[r][c], hi)
		}
		frame = append(frame, string(line))
	}

	// TODO: DEBUG remove it!
	// i want to see mines and numbers
	mines := []string{"mines"}
	for r := 0; r < 10; r++ {
		var line []rune
		for c := 0; c < 10; c++ {
			ch := m.mines[r][c]
			line = append(line, ' ', ch, ' ')
		}
		mines = append(mines, string(line))
	}

	return strings.Join(frame, "\n") + "\n\n" + strings.Join(mines, "\n")
}

func newModel(n, m, minesCount int) model {
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

	// setup mines
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for minesCount > 0 {
		r, c := rnd.Intn(n), rnd.Intn(m)
		if mines[r][c] != MINE {
			mines[r][c] = MINE
			minesCount--
		}
	}
	// count mines
	dirs := [][]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for r := 0; r < 10; r++ {
		for c := 0; c < 10; c++ {
			if mines[r][c] == MINE {
				for _, d := range dirs {
					newR, newC := r+d[0], c+d[1]
					if newR >= 0 && newR < 10 && newC >= 0 && newC < 10 {
						if mines[newR][newC] != MINE {
							mines[newR][newC]++
						}
					}
				}
			}
		}
	}

	return model{
		field: field,
		mines: mines,
		curr:  Point{0, 0},
	}
}

func main() {
	p := tea.NewProgram(newModel(10, 10, 10))
	if err := p.Start(); err != nil {
		panic(err)
	}
}
