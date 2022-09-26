package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type Point [2]int

type model struct {
	field [][]rune
	curr  Point
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyUp:
			m.curr[0]--
		case tea.KeyDown:
			m.curr[0]++
		case tea.KeyLeft:
			m.curr[1]--
		case tea.KeyRight:
			m.curr[1]++
		}
	}
	return m, nil
}

func (m model) View() string {
	var frame []string
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

func main() {
	var field [][]rune
	field = make([][]rune, 10)
	for i := 0; i < 10; i++ {
		field[i] = make([]rune, 10)
		for j := 0; j < 10; j++ {
			field[i][j] = '~'
		}
	}
	m := model{
		field: field,
		curr:  Point{0, 0},
	}
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		panic(err)
	}
}
