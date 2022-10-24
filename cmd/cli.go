package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/muesli/termenv"

	"reflect"
	"strings"
	"strconv"
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

func  LogsWidget(m LoggedModel) string {
	// TODO: add time marks and fancy colors

	// logsFrameSize should be less than 23 (visible ASCII colors 232-255)
	logsFrameSize := 10
	// logs := m.s.logger.GetLogs()
	logs := m.GetLogs()

	title := "LOGS:"
	var logLines []string

	limit := len(logs)
	if limit > logsFrameSize {
		limit = logsFrameSize
	}

	clrCode := 255
	s := func(c int, s string) string {
		clr := strconv.Itoa(c)
		return termenv.Style{}.Foreground(color(clr)).Styled(s)
	}

	for i := len(logs) - 1; i > len(logs)-limit; i-- {
		logLines = append(logLines, s(clrCode, logs[i]))
		clrCode -= 2
	}

	// TODO: extract it to utils
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
