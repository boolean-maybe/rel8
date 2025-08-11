package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"rel8/config"
)

// Header wraps a Flex with header-specific functionality
type Header struct {
	*tview.Flex
	leftHeader  *tview.TextView
	rightHeader *tview.TextView
}

// NewHeader creates a new header with proper configuration
func NewHeader() *Header {
	leftHeader := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

	headerText := ` [orange]Context: [aqua]dev[-]
 [orange]Cluster: [aqua]arn:aws:eks:us-east-1:897436064625:cluster/dev[-]
 [orange]Svc: [aqua]arn:aws:eks:us-east-1:897436064625:cluster/dev[-]
 [orange]K9s Rev: [aqua]v0.40.10 [silver](v0.50.9)[-]
 [orange]K8s Rev: [aqua]v1.28.15-eks-6096722[-]
 [orange]CPU: [lime]8%[-]
 [orange]MEM: [lime]8%[-]`

	leftHeader.SetText(headerText)
	leftHeader.SetBackgroundColor(tcell.ColorBlack)

	rightHeader := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(tview.AlignRight)

	artText := config.GetArt()
	artText = tview.TranslateANSI(artText)
	rightHeader.SetText(artText)
	rightHeader.SetBackgroundColor(tcell.ColorBlack)

	headerFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(leftHeader, 0, 1, false).
		AddItem(rightHeader, 0, 1, false).
		AddItem(nil, 1, 0, false)

	return &Header{
		Flex:        headerFlex,
		leftHeader:  leftHeader,
		rightHeader: rightHeader,
	}
}

// UpdateContext updates the context information in the header
func (h *Header) UpdateContext(context, cluster, service, k9sRev, k8sRev, cpu, mem string) {
	headerText := ` [orange]Context: [aqua]` + context + `[-]
 [orange]Cluster: [aqua]` + cluster + `[-]
 [orange]Svc: [aqua]` + service + `[-]
 [orange]K9s Rev: [aqua]` + k9sRev + `[-]
 [orange]K8s Rev: [aqua]` + k8sRev + `[-]
 [orange]CPU: [lime]` + cpu + `[-]
 [orange]MEM: [lime]` + mem + `[-]`

	h.leftHeader.SetText(headerText)
}

// UpdateArt updates the art on the right side of the header
func (h *Header) UpdateArt() {
	artText := config.GetArt()
	artText = tview.TranslateANSI(artText)
	h.rightHeader.SetText(artText)
}

// WrapHeader wraps header with same padding as other components
func WrapHeader(header *Header) *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 0, false). // Left padding
		AddItem(header.Flex, 0, 1, false).
		AddItem(nil, 0, 0, false) // Right padding
}
