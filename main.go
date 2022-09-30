package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/muesli/termenv"

	tea "github.com/charmbracelet/bubbletea"
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

var (
	color      = termenv.EnvColorProfile().Color
	oneMines   = termenv.Style{}.Foreground(color("4")).Styled
	twoMines   = termenv.Style{}.Foreground(color("2")).Styled
	threeMines = termenv.Style{}.Foreground(color("1")).Styled
	fourMines  = termenv.Style{}.Foreground(color("5")).Styled
	fiveMines  = termenv.Style{}.Foreground(color("5")).Styled
	sixMines   = termenv.Style{}.Foreground(color("6")).Styled
	sevenMines = termenv.Style{}.Foreground(color("7")).Styled
	eightMines = termenv.Style{}.Foreground(color("8")).Styled
	nineMines  = termenv.Style{}.Foreground(color("9")).Styled
)

type Point [2]int

type model struct {
	field, mines [][]rune
	n, m         int
	curr         Point

	leftToOpen int
	state      int
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state == OVER {
		return m, tea.Quit
	}

	// current cell on field
	c := m.field[m.curr[0]][m.curr[1]]
	// current cell on mines
	mine := m.mines[m.curr[0]][m.curr[1]]

	if msg, ok := msg.(tea.KeyMsg); ok {
		if m.state == WIN {
			return m, tea.Quit
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		// TODO: I'd like to add WASD control here as well
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
			switch mine {
			case MINE:
				m.state = OVER
			case ZERO:
				var openCell func(r, c int)
				openCell = func(r, c int) {
					if m.field[r][c] == EMPTY {
						return
					}

					if m.mines[r][c] != ZERO {
						m.field[r][c] = m.mines[r][c]
						return
					}

					m.field[r][c] = EMPTY
					m.leftToOpen--

					dirs := [][]int{
						{-1, -1}, {-1, 0}, {-1, 1},
						{0, -1}, {0, 1},
						{1, -1}, {1, 0}, {1, 1},
					}
					for _, d := range dirs {
						newR, newC := r+d[0], c+d[1]
						if newR >= 0 && newR < m.n && newC >= 0 && newC < m.m {
							openCell(newR, newC)
						}
					}
				}
				openCell(m.curr[0], m.curr[1])
				if m.leftToOpen == 0 {
					m.state = WIN
					for r := 0; r < m.n; r++ {
						for c := 0; c < m.m; c++ {
							if m.field[r][c] == HIDE {
								m.field[r][c] = m.mines[r][c]
							}
						}
					}
				}

			default:
				m.field[m.curr[0]][m.curr[1]] = mine
			}

		case tea.KeyEnter:
			switch c {
			case HIDE:
				m.field[m.curr[0]][m.curr[1]] = FLAG
			case FLAG:
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

	switch m.state {
	case GAME:
		for r := 0; r < m.n; r++ {
			var line string
			for c := 0; c < m.m; c++ {
				lo, hi := " ", " "
				if m.curr[0] == r && m.curr[1] == c {
					lo, hi = "[", "]"
				}
				line += lo
				line += styled(m.field[r][c])
				line += hi
			}
			frame = append(frame, line)
		}
	case WIN:
		for r := 0; r < m.n; r++ {
			var line string
			for c := 0; c < m.m; c++ {
				lo, hi := " ", " "
				line += lo
				line += styled(m.field[r][c])
				line += hi
			}
			frame = append(frame, line)
		}
		frame = append(frame, "YOU WON")
	case OVER:
		for r := 0; r < m.n; r++ {
			var line string
			for c := 0; c < m.m; c++ {
				lo, hi := " ", " "
				if m.curr[0] == r && m.curr[1] == c {
					lo, hi = "[", "]"
					line += lo + string(BOOM) + hi
					continue
				}
				if m.mines[r][c] == MINE {
					line += lo + string(MINE) + hi
					continue
				}
				line += lo
				line += styled(m.field[r][c])
				line += hi

			}
			frame = append(frame, string(line))
		}
		frame = append(frame, "GAME OVER")
	}

	// TODO: remove debug print
	// var ms []string
	// for _, r := range m.mines {
	// var line []rune
	// for _, c := range r {
	// line = append(line, c)
	// }
	// ms = append(ms, string(line))
	// }
	// frame = append(frame, strings.Join(ms, "\n"))
	// frame = append(frame, fmt.Sprintln(m.leftToOpen))

	return strings.Join(frame, "\n")
}

func styled(r rune) string {
	s := string(r)
	switch r {
	case '1':
		return oneMines(s)
	case '2':
		return twoMines(s)
	case '3':
		return threeMines(s)
	case '4':
		return fourMines(s)
	case '5':
		return fiveMines(s)
	case '6':
		return sixMines(s)
	case '7':
		return sevenMines(s)
	case '8':
		return eightMines(s)
	case '9':
		return nineMines(s)

	}
	return s
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

	return model{
		field:      field,
		mines:      mines,
		leftToOpen: shouldOpen,
		n:          n,
		m:          m,
		curr:       Point{0, 0},
	}
}

func main() {
	p := tea.NewProgram(newModel(10, 10, 10))
	if err := p.Start(); err != nil {
		panic(err)
	}
}
