package main

import (
	"strings"

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
	field [][]rune
	curr  Point
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
	return strings.Join(frame, "\n")
}

func newModel(n, m int) model {
	var field [][]rune
	field = make([][]rune, n)
	for i := 0; i < n; i++ {
		field[i] = make([]rune, m)
		for j := 0; j < m; j++ {
			field[i][j] = HIDE
		}
	}
	return model{
		field: field,
		curr:  Point{0, 0},
	}
}

func main() {
	p := tea.NewProgram(newModel(10, 10))
	if err := p.Start(); err != nil {
		panic(err)
	}
}
