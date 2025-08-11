package config

import (
	"fmt"
	"math"
	"strings"
)

const art = "██████╗ ███████╗██╗      █████╗ \n██╔══██╗██╔════╝██║     ██╔══██╗\n██████╔╝█████╗  ██║     ╚█████╔╝\n██╔══██╗██╔══╝  ██║     ██╔══██╗\n██║  ██║███████╗███████╗╚█████╔╝\n╚═╝  ╚═╝╚══════╝╚══════╝ ╚════╝ "

func PrintArt() {
	seed := 29
	// tweak these if you want to match your terminal vibe
	freq := 0.08  // lower = smoother rainbow
	spread := 2.0 // controls horizontal change rate

	fmt.Print(lolcatish(art, seed, freq, spread))
	fmt.Print("\x1b[0m") // reset
}

func GetArt() string {
	seed := 29
	// tweak these if you want to match your terminal vibe
	freq := 0.02   // lower = smoother rainbow
	spread := 50.0 // controls horizontal change rate

	return "\n" + lolcatish(art, seed, freq, spread) + "\x1b[0m" + "\n\n"
}

// lolcatish applies a deterministic 256-color rainbow similar to: | lolcat -S <seed>
func lolcatish(s string, seed int, freq, spread float64) string {
	var b strings.Builder
	phase := float64(seed%360) * (math.Pi / 180.0) // seed -> phase (radians)

	for _, line := range strings.Split(s, "\n") {
		for x, r := range line {
			// hue varies smoothly across x
			h := 0.5 + 0.5*math.Sin(freq*float64(x)/spread+phase) // [0..1]
			r8, g8, b8 := hsvToRGB(h, 1, 1)
			ansi := rgbToANSI256(r8, g8, b8)
			b.WriteString(fmt.Sprintf("\x1b[38;5;%dm%s", ansi, string(r)))
		}
		b.WriteString("\x1b[0m\n") // reset each line so copy/paste looks right
	}
	return b.String()
}

// HSV(0..1) -> RGB(0..255)
func hsvToRGB(h, s, v float64) (int, int, int) {
	if s == 0 {
		c := int(v * 255)
		return c, c, c
	}
	h = math.Mod(h*6, 6)
	i := math.Floor(h)
	f := h - i
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	var r, g, bl float64
	switch int(i) {
	case 0:
		r, g, bl = v, t, p
	case 1:
		r, g, bl = q, v, p
	case 2:
		r, g, bl = p, v, t
	case 3:
		r, g, bl = p, q, v
	case 4:
		r, g, bl = t, p, v
	default:
		r, g, bl = v, p, q
	}
	return int(r * 255), int(g * 255), int(bl * 255)
}

// Map truecolor to ANSI 256-color cube (16..231)
func rgbToANSI256(r, g, b int) int {
	rc := int(math.Round(float64(r) / 255 * 5))
	gc := int(math.Round(float64(g) / 255 * 5))
	bc := int(math.Round(float64(b) / 255 * 5))
	return 16 + 36*rc + 6*gc + bc
}
