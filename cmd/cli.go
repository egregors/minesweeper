package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"

	g "github.com/egregors/minesweeper/pkg"
	"github.com/muesli/termenv"

	"reflect"
	"strconv"
	"strings"
)

// noop is a No Operation event just to update UI
type noop struct{}

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

type LoggedModel interface {
	GetLogs() []string
}

func LogsWidget(m LoggedModel, tail int) string {
	// tail should be less than 23 (visible ASCII colors 232-255)
	if tail == 0 {
		tail = 10
	}
	logs := m.GetLogs()

	title := termenv.Style{}.Bold().Styled("LOGS:")
	var logLines []string

	limit := len(logs)
	if limit > tail {
		limit = tail
	}

	// Use gradient of colors from dark to light (232-255)
	clrCode := 255
	clrStep := 23 / tail
	if clrStep < 1 {
		clrStep = 1
	}

	for i := len(logs) - 1; i > len(logs)-limit; i-- {
		log := logs[i]
		
		// Parse and colorize timestamp if present
		// Log format is typically: "2006/01/02 15:04:05 message"
		coloredLog := log
		if len(log) > 19 && log[4] == '/' && log[7] == '/' && log[13] == ':' && log[16] == ':' {
			timestamp := log[:19]
			message := log[19:]
			
			// Style timestamp in cyan/blue
			styledTimestamp := termenv.Style{}.Foreground(color("39")).Styled(timestamp)
			
			// Gradient style for message
			clr := strconv.Itoa(clrCode)
			styledMessage := termenv.Style{}.Foreground(color(clr)).Styled(message)
			
			coloredLog = styledTimestamp + styledMessage
		} else {
			// No timestamp, just apply gradient color
			clr := strconv.Itoa(clrCode)
			coloredLog = termenv.Style{}.Foreground(color(clr)).Styled(log)
		}
		
		logLines = append(logLines, coloredLog)
		clrCode -= clrStep
		if clrCode < 232 {
			clrCode = 232
		}
	}

	g.ReverseStrings(logLines)

	return strings.Join([]string{
		title,
		strings.Join(logLines, ""),
	}, "\n")
}
