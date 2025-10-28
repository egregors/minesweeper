package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	g "github.com/egregors/minesweeper/pkg"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"strings"
)

type clientState int

const (
	INIT clientState = iota
	GAME
	OVER
)

type Client struct {
	serverAddr string
	conn       net.Conn

	game *g.Game
	ui   *tea.Program

	state clientState

	logger g.Logger
	dbg    bool

	mu sync.Mutex
}

func NewClient(serverAddr string, logger g.Logger, dbg bool) *Client {
	c := new(Client)
	c.serverAddr = serverAddr

	c.dbg = dbg
	c.logger = logger

	return c
}

func (c *Client) updateGame(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	g.FromGob(data, &c.game)
}

func (c *Client) connect() error {
	if c.conn == nil {
		conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), c.serverAddr)
		if err != nil {
			return fmt.Errorf("can't conntect to the server: %e", err)
		}
		c.conn = conn
	}
	return nil
}

func (c *Client) pullServerEvents() {
	for {
		switch c.state {
		case GAME:
			msg, _, err := wsutil.ReadServerData(c.conn)
			if err != nil {
				log.Printf("Can't receive data: %s", err.Error())
				continue
			}
			c.updateGame(msg)
			log.Printf("Updated: %s", c.game)
			c.ui.Send(noop{})
		case OVER:
			// TODO: why I do not see it in log?
			log.Print("Game is over, stop pulling...")
			c.ui.Send(noop{})
			return
		}
	}
}

func (c *Client) Run() error {
	log.Println("Client started")
	// connection retry loop
	for {
		if err := c.connect(); err != nil {
			log.Printf("sleep before retry: %s", err.Error())
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		log.Println("connected!")
		// Ensure connection is closed when function exits
		defer func() {
			if c.conn != nil {
				_ = c.conn.Close()
			}
		}()

		// game loop
		for {
			switch c.state {
			case INIT:
				// hi server message
				if err := wsutil.WriteClientMessage(c.conn, ws.OpText, nil); err != nil {
					fmt.Println("Cannot send: " + err.Error())
					continue
				}

				// get game data
				msg, _, err := wsutil.ReadServerData(c.conn)
				if err != nil {
					fmt.Println("Cannot receive data: " + err.Error())
					continue
				}

				// start game
				// TODO: extract it somehow
				c.state = GAME
				c.updateGame(msg)
				c.ui = tea.NewProgram(clientUIModel{
					Model: c.game.M,
					Conn:  c.conn,
					Cur:   g.Point{},
					Dbg:   c.dbg,
					C:     c,
				})

				// pull game update form the server
				go c.pullServerEvents()

				log.Print("UI started")
				return c.ui.Start()
			}
		}
	}
}

type clientUIModel struct {
	*g.Model
	Cur  g.Point
	Conn net.Conn

	C *Client

	Dbg bool
}

func (m clientUIModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m clientUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// current cell on Field
	c := m.Field[m.Cur[0]][m.Cur[1]]

	switch msg := msg.(type) {
	case noop:
		// update UI
		return m, nil

	case tea.KeyMsg:
		// control
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

		case tea.KeySpace:
			if m.State == g.GAME {
				eT = g.OpenCell
			}

		case tea.KeyEnter:
			if m.State == g.GAME {
				switch c {
				case g.HIDE:
					m.Field[m.Cur[0]][m.Cur[1]] = g.FLAG
				case g.FLAG:
					m.Field[m.Cur[0]][m.Cur[1]] = g.GESS
				case g.GESS:
					m.Field[m.Cur[0]][m.Cur[1]] = g.HIDE
				}
			}

		default:
			// WASD controls
			switch msg.String() {
			case "w":
				if m.Cur[0] > 0 {
					m.Cur[0]--
					eT = g.CursorMove
				}
			case "s":
				if m.Cur[0] < len(m.Field)-1 {
					m.Cur[0]++
					eT = g.CursorMove
				}
			case "a":
				if m.Cur[1] > 0 {
					m.Cur[1]--
					eT = g.CursorMove
				}
			case "d":
				if m.Cur[1] < len(m.Field[0])-1 {
					m.Cur[1]++
					eT = g.CursorMove
				}
			}
		}
	}
	return m, nil
}

func (m clientUIModel) View() string {
	frame := []string{
		m.titleFrame(),
		m.fieldFrame(),
		m.statusFrame(),
		LogsWidget(m, 5),
	}

	if m.Dbg {
		frame = append(frame, DebugWidget(m))
	}

	return strings.Join(frame, "\n")
}

func (m clientUIModel) titleFrame() string {
	// Calculate field width (each cell is 3 characters: space + char + space)
	fieldWidth := m.M * 3
	title := "*** Minesweeper ***"
	separator := strings.Repeat("=", len(title))
	
	// Center the title based on field width
	titlePadding := (fieldWidth - len(title)) / 2
	if titlePadding < 0 {
		titlePadding = 0
	}
	paddedTitle := strings.Repeat(" ", titlePadding) + title
	paddedSeparator := strings.Repeat(" ", titlePadding) + separator
	
	return strings.Join([]string{paddedTitle, paddedSeparator}, "\n")
}

func (m clientUIModel) fieldFrame() string {
	var lines []string
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
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m clientUIModel) statusFrame() string {
	switch m.State {
	case g.WIN:
		return "YOU WON"
	case g.OVER:
		return "GAME OVER"
	default:
		return ""
	}
}

func (m clientUIModel) GetLogs() []string {
	return m.C.logger.GetLogs()
}
