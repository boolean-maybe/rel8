package view

import (
	"github.com/rivo/tview"
	"rel8/config"
)

// Header wraps a Flex with header-specific functionality
type Header struct {
	*tview.Flex
	leftHeader  *tview.TextView
	keys        *Keys
	rightHeader *tview.TextView
}

// NewHeader creates a new header with proper configuration
func NewHeader() *Header {
	leftHeader := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

	headerText := ` [` + Colors.HeaderLabel + `]Context: [` + Colors.HeaderValue + `]dev[-]
 [` + Colors.HeaderLabel + `]Cluster: [` + Colors.HeaderValue + `]arn:aws:eks:us-east-1:897436064625:cluster/dev[-]
 [` + Colors.HeaderLabel + `]Svc: [` + Colors.HeaderValue + `]arn:aws:eks:us-east-1:897436064625:cluster/dev[-]
 [` + Colors.HeaderLabel + `]K9s Rev: [` + Colors.HeaderValue + `]v0.40.10 [` + Colors.HeaderSecondary + `](v0.50.9)[-]
 [` + Colors.HeaderLabel + `]K8s Rev: [` + Colors.HeaderValue + `]v1.28.15-eks-6096722[-]
 [` + Colors.HeaderLabel + `]CPU: [` + Colors.HeaderHighlight + `]8%[-]
 [` + Colors.HeaderLabel + `]MEM: [` + Colors.HeaderHighlight + `]8%[-]`

	leftHeader.SetText(headerText)
	leftHeader.SetBackgroundColor(Colors.BackgroundDefault)

	rightHeader := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(tview.AlignRight)

	artText := config.GetArt()
	artText = tview.TranslateANSI(artText)
	rightHeader.SetText(artText)
	rightHeader.SetBackgroundColor(Colors.BackgroundDefault)

	keys := NewKeys()

	headerFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(leftHeader, 0, 1, false).
		AddItem(keys.Flex, 0, 1, false).
		AddItem(rightHeader, 0, 1, false).
		AddItem(nil, 1, 0, false)

	return &Header{
		Flex:        headerFlex,
		leftHeader:  leftHeader,
		keys:        keys,
		rightHeader: rightHeader,
	}
}

// UpdateContext updates the context information in the header
func (h *Header) UpdateContext(context, cluster, service, k9sRev, k8sRev, cpu, mem string) {
	headerText := ` [` + Colors.HeaderLabel + `]Context: [` + Colors.HeaderValue + `]` + context + `[-]
 [` + Colors.HeaderLabel + `]Cluster: [` + Colors.HeaderValue + `]` + cluster + `[-]
 [` + Colors.HeaderLabel + `]Svc: [` + Colors.HeaderValue + `]` + service + `[-]
 [` + Colors.HeaderLabel + `]K9s Rev: [` + Colors.HeaderValue + `]` + k9sRev + `[-]
 [` + Colors.HeaderLabel + `]K8s Rev: [` + Colors.HeaderValue + `]` + k8sRev + `[-]
 [` + Colors.HeaderLabel + `]CPU: [` + Colors.HeaderHighlight + `]` + cpu + `[-]
 [` + Colors.HeaderLabel + `]MEM: [` + Colors.HeaderHighlight + `]` + mem + `[-]`

	h.leftHeader.SetText(headerText)
}

// UpdateKeys updates the keys display in the middle section
func (h *Header) UpdateKeys(text string) {
	h.keys.UpdateKeys(text)
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
