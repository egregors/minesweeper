package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"

	g "github.com/egregors/minesweeper/pkg"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var (
	Color      = termenv.EnvColorProfile().Color
	RedStyle   = termenv.Style{}.Foreground(Color("1")).Styled
	GreenStyle = termenv.Style{}.Foreground(Color("2")).Styled

	P1Style = termenv.Style{}.Foreground(Color("11")).Styled
	P2Style = termenv.Style{}.Foreground(Color("13")).Styled

	// 	mainStyle       = termenv.Style{}.Foreground(color("11")).Styled
	// 	modelFieldStyle = termenv.Style{}.Foreground(color("39")).Styled
	// 	modelValStyle   = termenv.Style{}.Foreground(color("87")).Styled
)

type player struct {
	id       string
	addr     string
	conn     net.Conn
	isOnline bool

	cur g.Point
}

func (p player) String() string {
	online := GreenStyle("ON-LINE")
	offline := RedStyle("OFF-LINE")

	status := offline
	if p.isOnline {
		status = online
	}

	var id string
	if p.id == "P1" {
		id = P1Style(p.id)
	}

	if p.id == "P2" {
		id = P2Style(p.id)
	}

	return fmt.Sprintf("%s [%s]: %s => [%d:%d]", id, p.addr, status, p.cur[0], p.cur[1])
}

type players map[string]*player

func (ps players) getByID(id string) *player {
	for _, v := range ps {
		if v.id == id {
			return v
		}
	}
	return nil
}

func (ps players) add(player *player) {
	if p, ok := ps[player.addr]; ok {
		// reconnect
		p.isOnline = true
	} else {
		// new player
		n := len(ps) + 1
		ps[player.addr] = player
		ps[player.addr].id = fmt.Sprintf("P%d", n)
	}
}

func (ps players) disconnect(addr string) {
	for k, v := range ps {
		if v.addr == addr {
			ps[k].isOnline = false
			return
		}
	}
}

type Srv struct {
	game *g.Game
	ps   players
	ui   *tea.Program

	logger g.Logger
	dbg    bool
}

func NewServer(game *g.Game, dbg bool, logger g.Logger) *Srv {
	s := new(Srv)
	s.game = game
	s.ps = make(players)
	s.ui = tea.NewProgram(serverUIModel{
		Model: game.M,
		s:     s,
		dbg:   false,
	})

	s.dbg = dbg
	s.logger = logger

	return s
}

func (s *Srv) String() string {
	ls := []string{"\n"}
	for _, v := range s.ps {
		status := "ONLINE"
		if !v.isOnline {
			status = "OFFLINE"
		}
		ls = append(ls, fmt.Sprintf("%s:%s CURR: %s", v.addr, status, v.cur.String()))
	}
	return strings.Join(ls, "\n")
}

func (s *Srv) Run() error {
	// start WS server
	log.Print("Server started, waiting for connection from players...")

	go func() {
		http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, _, _, err := ws.UpgradeHTTP(r, w)
			if err != nil {
				// TODO: use own logger
				log.Println("Error starting socket server: " + err.Error())
				log.Printf("Error starting socket server: %s", err.Error())
			}

			addr := conn.RemoteAddr().String()
			log.Printf("Client %s connected", addr)

			go func() {
				defer func() { _ = conn.Close() }()
				for {
					msg, op, err := wsutil.ReadClientData(conn)
					if err != nil {
						log.Printf("Error receiving data: " + err.Error())
						log.Printf("Client %s disconnected", addr)
						s.ps.disconnect(addr)
						s.ui.Send(*s.ps[addr])
						return
					}

					switch op {
					case ws.OpText:
						// join \ rejoin player
						s.ps.add(&player{
							conn:     conn,
							addr:     addr,
							isOnline: true,
						})
						s.ui.Send(*s.ps[addr])

						// send game state to client
						if err := wsutil.WriteServerMessage(conn, ws.OpBinary, s.game.ToGob()); err != nil {
							log.Printf("Error sending data: %s", err.Error())
							log.Print("Client disconnected")
							s.ps.disconnect(addr)
							return
						}

					case ws.OpBinary:
						// TODO:
						//  - [ ] receive GUSS, FLAG, OPEN command
						//  - [ ] render new admin field after OPEN
						//  - [ ] turns (P1 -> P2 -> ...)
						//  - [ ] client commands and maybe refactor client code maybe

						var p g.Point
						p.FromGob(msg)
						s.ps[addr].cur = p
						log.Printf("GOT request: %s", p.String())
						s.ui.Send(*s.ps[addr])
					}
				}
			}()

		}))
	}()

	log.Print("UI started")
	return s.ui.Start()
}

type serverUIModel struct {
	g.Model
	s   *Srv
	dbg bool
}

func (m serverUIModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m serverUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("msg: %T", msg)

	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case player:
		//_ := msg.(player)
		//m.s.log(fmt.Sprintf("%v", p))
		//m.ps[p.addr].cur = p.cur

		return m, nil
	default:
		log.Print("WRONG TYPE")
		return m, nil
	}
}

func (m serverUIModel) View() string {
	frames := []string{
		m.titleFrame(),
		m.fieldFrame(),
		m.playersFrame(),
		m.logsFrame(),
	}
	return strings.Join(frames, "\n")
}

func (m serverUIModel) titleFrame() string {
	return strings.Join([]string{
		"     *** Minesweeper ***",
		"     ===================",
	}, "\n")
}

func (m serverUIModel) fieldFrame() string {
	var frames []string

	var p1Cur, p2Cur *g.Point
	if p1 := m.s.ps.getByID("P1"); p1 != nil {
		p1Cur = &p1.cur
	}
	if p2 := m.s.ps.getByID("P2"); p2 != nil {
		p2Cur = &p2.cur
	}

	for r := 0; r < m.N; r++ {
		var line string
		for c := 0; c < m.M; c++ {
			// player cursors marks
			lo, hi := " ", " "
			{
				if p1Cur != nil {
					if r == p1Cur[0] && c == p1Cur[1] {
						lo = P1Style("[")
						hi = P1Style("]")
					}
				}
				if p2Cur != nil {
					if r == p2Cur[0] && c == p2Cur[1] {
						lo = P2Style("[")
						hi = P2Style("]")
					}
				}
				if p1Cur != nil && p2Cur != nil {
					if p1Cur[0] == p2Cur[0] && p1Cur[1] == p2Cur[1] {
						if r == p1Cur[0] && c == p1Cur[1] {
							lo = P1Style("[")
						}
					}
				}
			}

			// mines
			cell := m.Field[r][c]
			if m.Mines[r][c] == g.MINE {
				cell = g.MINE
			}

			line += lo
			line += string(cell)
			line += hi
		}
		frames = append(frames, line)
	}
	return strings.Join(frames, "\n")
}

func (m serverUIModel) playersFrame() string {
	var ps []string
	for _, v := range m.s.ps {
		ps = append(ps, v.String())
	}
	return strings.Join(ps, "\n")
}

func (m serverUIModel) logsFrame() string {
	// TODO: add time marks and fancy colors

	// logsFrameSize should be less than 23 (visible ASCII colors 232-255)
	logsFrameSize := 10
	logs := m.s.logger.GetLogs()

	title := "LOGS:"
	var logLines []string

	limit := len(logs)
	if limit > logsFrameSize {
		limit = logsFrameSize
	}

	// TODO: extract it maybe?
	clrCode := 255
	s := func(c int, s string) string {
		clr := strconv.Itoa(c)
		return termenv.Style{}.Foreground(color(clr)).Styled(s)
	}

	for i := len(logs) - 1; i > len(logs)-limit; i-- {
		logLines = append(logLines, s(clrCode, logs[i]))
		clrCode -= 2
	}

	rev := func(xs []string) {
		for i := 0; i < len(xs)/2; i++ {
			xs[i], xs[len(xs)-1-i] = xs[len(xs)-1-i], xs[i]
		}
	}

	rev(logLines)

	return strings.Join([]string{
		title,
		strings.Join(logLines, ""),
	}, "\n")
}
