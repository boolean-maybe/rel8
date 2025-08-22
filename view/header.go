package view

import (
	"rel8/config"
	"rel8/model"
	"strings"

	"github.com/rivo/tview"
)

func NewHeader(serverInfo *model.HeaderInfo) *tview.Flex {
	serverInfoHeader := createServerInfoHeader(serverInfo)
	keyHelpHeader := NewKeys()
	logoHeader := createLogoHeader()

	headerFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(serverInfoHeader, 0, 1, false).
		AddItem(keyHelpHeader.Flex, 0, 1, false).
		AddItem(logoHeader, 0, 1, false).
		AddItem(nil, 1, 0, false)

	return headerFlex
}

func createServerInfoHeader(serverInfo *model.HeaderInfo) *tview.TextView {
	serverInfoHeader := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

	serverInfoHeader.SetBackgroundColor(Colors.BackgroundDefault)

	// build header text dynamically from all key/value pairs in *h.serverInfo, preserving formatting
	info := *serverInfo

	// collect and sort keyHelpHeader for consistent order
	var keys []string
	for k := range info {
		keys = append(keys, k)
	}
	//sort.Strings(keyHelpHeader)

	// build header lines from all key/value pairs, all lines indented equally
	var headerLines []string
	for _, k := range keys {
		v := info[k]
		headerLines = append(headerLines, " ["+Colors.HeaderLabel+"]"+k+": ["+Colors.HeaderValue+"]"+v+"[-]")
	}

	// join lines with newline, all lines have same indent (including the first line)
	headerText := " " + strings.Join(headerLines, "\n ")

	serverInfoHeader.SetText(headerText)
	return serverInfoHeader
}

func createLogoHeader() *tview.TextView {
	rightHeader := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(tview.AlignRight)

	artText := config.GetArt()
	artText = tview.TranslateANSI(artText)
	rightHeader.SetText(artText)
	rightHeader.SetBackgroundColor(Colors.BackgroundDefault)
	return rightHeader
}
