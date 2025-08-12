package view

import "github.com/gdamore/tcell/v2"

// AppColors holds all colors used throughout the view package
type AppColors struct {
	// Background colors
	BackgroundDefault tcell.Color

	// Border colors
	BorderDefault tcell.Color

	// Text colors - tcell colors
	TextWhite        tcell.Color
	TextLightSkyBlue tcell.Color
	TextBlack        tcell.Color
	TextAqua         tcell.Color

	// Text colors - tview color tags
	KeyColor        string // For key bindings
	TextDefault     string // Default text color
	HeaderLabel     string // For header labels (Context:, CPU:, etc.)
	HeaderValue     string // For header values
	HeaderHighlight string // For highlighted header values
	HeaderSecondary string // For secondary header text

	// UI selection/row highlight
	SelectionBandBg tcell.Color // background for full-width selection band
}

// DefaultColors returns the default color scheme
func DefaultColors() *AppColors {
	return &AppColors{
		// Background colors
		BackgroundDefault: tcell.ColorBlack,

		// Border colors
		BorderDefault: tcell.ColorLightSkyBlue,

		// Text colors - tcell colors
		TextWhite:        tcell.ColorWhite,
		TextLightSkyBlue: tcell.ColorLightSkyBlue,
		TextBlack:        tcell.ColorBlack,
		TextAqua:         tcell.ColorAqua,

		// Text colors - tview color tags
		KeyColor:        "#00BFFF", // Bright blue for key bindings
		TextDefault:     "white",   // Default white text
		HeaderLabel:     "orange",  // Orange for labels like "Context:"
		HeaderValue:     "aqua",    // Aqua for values like "dev"
		HeaderHighlight: "lime",    // Lime for highlighted values like CPU/MEM percentages
		HeaderSecondary: "silver",  // Silver for secondary text

		// UI selection/row highlight
		SelectionBandBg: tcell.ColorAqua,
	}
}

// Global color instance - can be replaced for theming
var Colors = DefaultColors()

// SetColors allows changing the global color scheme
func SetColors(colors *AppColors) {
	Colors = colors
}
