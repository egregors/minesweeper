package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	game "github.com/egregors/minesweeper/pkg"
	"github.com/muesli/termenv"
	"reflect"
	"strings"
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

	mainStyle       = termenv.Style{}.Foreground(color("11")).Styled
	modelFieldStyle = termenv.Style{}.Foreground(color("39")).Styled
	modelValStyle   = termenv.Style{}.Foreground(color("87")).Styled
)

func RunCli(m game.Model) error {
	UI := tea.NewProgram(model{m})
	if err := UI.Start(); err != nil {
		return err
	}
	return nil
}

type model struct {
	game.Model
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.State == game.OVER {
		return m, tea.Quit
	}

	// current cell on Field
	c := m.Field[m.Curr[0]][m.Curr[1]]
	// current cell on Mines
	mine := m.Mines[m.Curr[0]][m.Curr[1]]

	if msg, ok := msg.(tea.KeyMsg); ok {
		if m.State == game.WIN {
			return m, tea.Quit
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		// TODO: I'd like to add WASD control here as well
		case tea.KeyUp:
			if m.Curr[0] > 0 {
				m.Curr[0]--
			}
		case tea.KeyDown:
			if m.Curr[0] < len(m.Field)-1 {
				m.Curr[0]++
			}
		case tea.KeyLeft:
			if m.Curr[1] > 0 {
				m.Curr[1]--
			}
		case tea.KeyRight:
			if m.Curr[1] < len(m.Field[0])-1 {
				m.Curr[1]++
			}

		case tea.KeySpace:
			switch mine {
			case game.MINE:
				m.State = game.OVER
			case game.ZERO:
				var openCell func(r, c int)
				openCell = func(r, c int) {
					if m.Field[r][c] == game.EMPTY {
						return
					}

					if m.Mines[r][c] != game.ZERO {
						m.Field[r][c] = m.Mines[r][c]
						return
					}

					m.Field[r][c] = game.EMPTY
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
				openCell(m.Curr[0], m.Curr[1])
				if m.LeftToOpen == 0 {
					m.State = game.WIN
					for r := 0; r < m.N; r++ {
						for c := 0; c < m.M; c++ {
							if m.Field[r][c] == game.HIDE {
								m.Field[r][c] = m.Mines[r][c]
							}
						}
					}
				}

			default:
				m.Field[m.Curr[0]][m.Curr[1]] = mine
			}

		case tea.KeyEnter:
			switch c {
			case game.HIDE:
				m.Field[m.Curr[0]][m.Curr[1]] = game.FLAG
			case game.FLAG:
				m.Field[m.Curr[0]][m.Curr[1]] = game.GESS
			case game.GESS:
				m.Field[m.Curr[0]][m.Curr[1]] = game.HIDE
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	// TODO: title should be relative to Field wight
	frame := []string{
		"     *** Minesweeper ***",
		"     ===================",
	}

	switch m.State {
	case game.GAME:
		for r := 0; r < m.N; r++ {
			var line string
			for c := 0; c < m.M; c++ {
				lo, hi := " ", " "
				if m.Curr[0] == r && m.Curr[1] == c {
					lo, hi = "[", "]"
				}
				line += lo
				line += styled(m.Field[r][c])
				line += hi
			}
			frame = append(frame, line)
		}
	case game.WIN:
		for r := 0; r < m.N; r++ {
			var line string
			for c := 0; c < m.M; c++ {
				lo, hi := " ", " "
				line += lo
				line += styled(m.Field[r][c])
				line += hi
			}
			frame = append(frame, line)
		}
		frame = append(frame, "YOU WON")
	case game.OVER:
		for r := 0; r < m.N; r++ {
			var line string
			for c := 0; c < m.M; c++ {
				lo, hi := " ", " "
				if m.Curr[0] == r && m.Curr[1] == c {
					lo, hi = "[", "]"
					line += lo + string(game.BOOM) + hi
					continue
				}
				if m.Mines[r][c] == game.MINE {
					line += lo + string(game.MINE) + hi
					continue
				}
				line += lo
				line += styled(m.Field[r][c])
				line += hi

			}
			frame = append(frame, string(line))
		}
		frame = append(frame, "GAME OVER")
	}

	if m.Dbg {
		frame = append(frame, DebugWidget(m))
	}

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

type DebugModel interface {
	tea.Model
}

func DebugWidget(m DebugModel) string {
	err := "ERR: "

	s := strings.Join([]string{
		mainStyle("= = = = = DEBUG = = = = ="),
		mainStyle("| ") + err,
		getModelFrame(m),
		mainStyle("= = = = = ----- = = = = ="),
	}, "\n")

	return s
}

func getModelFrame(m DebugModel) string {
	e := reflect.ValueOf(&m).Elem().Elem()
	n := e.NumField()
	res := make([]string, n)

	for i := 0; i < n; i++ {
		varName := e.Type().Field(i).Name
		varType := e.Type().Field(i).Type
		varValue := e.Field(i).Interface()

		res[i] = fmt.Sprintf(
			"%-40s %-20s %-30s",
			mainStyle("| ")+modelFieldStyle(varName),
			"["+varType.String()+"]",
			modelValStyle(fmt.Sprintf("%v", varValue)),
		)
	}

	return strings.Join(res, "\n")
}
