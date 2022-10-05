package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/muesli/termenv"

	"fmt"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jessevdk/go-flags"
	"os"
)

type Difficulty int

const (
	EASY Difficulty = iota
	NORMAL
	HARD
)

type Game struct {
	UI         tea.Program
	difficulty Difficulty
}

func NewGame(difficulty Difficulty, dbg bool) *Game {
	var m model

	switch difficulty {
	case EASY:
		m = newModel(9, 9, 10, dbg)
	case NORMAL:
		m = newModel(16, 16, 40, dbg)
	case HARD:
		m = newModel(16, 30, 99, dbg)
	}

	return &Game{
		UI: *tea.NewProgram(m),
	}
}

func (g *Game) Run() error {
	return g.UI.Start()
}

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

	mainStyle       = termenv.Style{}.Foreground(color("11")).Styled
	modelFieldStyle = termenv.Style{}.Foreground(color("39")).Styled
	modelValStyle   = termenv.Style{}.Foreground(color("87")).Styled
)

type Point [2]int

type model struct {
	Field, Mines [][]rune
	N, M         int
	Curr         Point

	LeftToOpen int
	State      int

	Dbg bool
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.State == OVER {
		return m, tea.Quit
	}

	// current cell on Field
	c := m.Field[m.Curr[0]][m.Curr[1]]
	// current cell on Mines
	mine := m.Mines[m.Curr[0]][m.Curr[1]]

	if msg, ok := msg.(tea.KeyMsg); ok {
		if m.State == WIN {
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
			case MINE:
				m.State = OVER
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
				openCell(m.Curr[0], m.Curr[1])
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
				m.Field[m.Curr[0]][m.Curr[1]] = mine
			}

		case tea.KeyEnter:
			switch c {
			case HIDE:
				m.Field[m.Curr[0]][m.Curr[1]] = FLAG
			case FLAG:
				m.Field[m.Curr[0]][m.Curr[1]] = GESS
			case GESS:
				m.Field[m.Curr[0]][m.Curr[1]] = HIDE
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
	case GAME:
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
	case WIN:
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
	case OVER:
		for r := 0; r < m.N; r++ {
			var line string
			for c := 0; c < m.M; c++ {
				lo, hi := " ", " "
				if m.Curr[0] == r && m.Curr[1] == c {
					lo, hi = "[", "]"
					line += lo + string(BOOM) + hi
					continue
				}
				if m.Mines[r][c] == MINE {
					line += lo + string(MINE) + hi
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

func newModel(n, m, minesCount int, dbg bool) model {
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

	return model{
		Field:      field,
		Mines:      mines,
		LeftToOpen: shouldOpen,
		N:          n,
		M:          m,
		Curr:       Point{0, 0},
		Dbg:        dbg,
	}
}

// DebugModel is model supports debugging widget
type DebugModel interface {
	tea.Model
}

// DebugWidget returns some useful for debugging information
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

type Opts struct {
	Dbg bool `long:"dbg" env:"DEBUG" description:"Debug mode"`
}

func main() {
	// TODO:
	// - [x] add debug mode
	// - [x] get arguments from command line
	// - [x] add Game abstraction
	// - [ ] tutorial
	// - [ ] fancy title and game over \ won message
	// - [x] difficulty level
	//  * ease	- 9x9 board containing 10 mines
	//  * normal	- 16x16 board with 40 mines
	//  * hard	- 30x16 board with 99 mines
	// - [ ] endless circle game (start the new one, when player won or lose)

	var opts Opts
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("cli error: %v", err)
		}
		os.Exit(2)
	}

	g := NewGame(EASY, opts.Dbg)
	if err := g.Run(); err != nil {
		panic(err)
	}
}
