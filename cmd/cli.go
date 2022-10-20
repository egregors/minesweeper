package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	g "github.com/egregors/minesweeper/pkg"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/muesli/termenv"
	"log"
	"net"
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

func RunCli(m *g.Model, conn net.Conn) error {
	UI := tea.NewProgram(model{m, g.Point{}, conn})
	if err := UI.Start(); err != nil {
		return err
	}
	return nil
}

type model struct {
	*g.Model
	// all this field should be exposed because of debug mode
	Cur  g.Point
	Conn net.Conn
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.State == g.OVER {
		return m, tea.Quit
	}

	// current cell on Field
	c := m.Field[m.Cur[0]][m.Cur[1]]
	// current cell on Mines
	// mine := m.Mines[m.Cur[0]][m.Cur[1]]

	if msg, ok := msg.(tea.KeyMsg); ok {
		if m.State == g.WIN {
			return m, tea.Quit
		}

		// each Update client state should send this state on server
		var eT g.EventType
		defer func(eT *g.EventType) {
			e := g.NewEvent(*eT, m.Cur)
			if err := wsutil.WriteClientMessage(m.Conn, ws.OpBinary, e.Bytes()); err != nil {
				log.Printf("can't sent cur to server")
			}
		}(&eT)

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		// TODO: I'd like to add WASD control here as well
		case tea.KeyUp:
			if m.Cur[0] > 0 {
				m.Cur[0]--
				eT = g.CursorMove
			}
		case tea.KeyDown:
			if m.Cur[0] < len(m.Field)-1 {
				m.Cur[0]++
				eT = g.CursorMove
			}
		case tea.KeyLeft:
			if m.Cur[1] > 0 {
				m.Cur[1]--
				eT = g.CursorMove
			}
		case tea.KeyRight:
			if m.Cur[1] < len(m.Field[0])-1 {
				m.Cur[1]++
				eT = g.CursorMove
			}

		// TODO: all open-cell logic should calculate server
		case tea.KeySpace:
			eT = g.OpenCell
			
		case tea.KeyEnter:
			switch c {
			case g.HIDE:
				m.Field[m.Cur[0]][m.Cur[1]] = g.FLAG
			case g.FLAG:
				m.Field[m.Cur[0]][m.Cur[1]] = g.GESS
			case g.GESS:
				m.Field[m.Cur[0]][m.Cur[1]] = g.HIDE
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
	case g.GAME:
		for r := 0; r < m.N; r++ {
			var line string
			for c := 0; c < m.M; c++ {
				lo, hi := " ", " "
				if m.Cur[0] == r && m.Cur[1] == c {
					lo, hi = "[", "]"
				}
				line += lo
				line += styled(m.Field[r][c])
				line += hi
			}
			frame = append(frame, line)
		}
	case g.WIN:
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
	case g.OVER:
		for r := 0; r < m.N; r++ {
			var line string
			for c := 0; c < m.M; c++ {
				lo, hi := " ", " "
				if m.Cur[0] == r && m.Cur[1] == c {
					lo, hi = "[", "]"
					line += lo + string(g.BOOM) + hi
					continue
				}
				if m.Mines[r][c] == g.MINE {
					line += lo + string(g.MINE) + hi
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
