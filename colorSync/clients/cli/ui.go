package main 

import (
	"fmt"
	"strings"
	
	"github.com/fatih/color"
)

// UI handles terminal display
type UI struct {
	// Color functions for terminal output
	red   *color.Color
	blue   *color.Color
	green  *color.Color
	yellow *color.Color
	cyan   *color.Color
	bold   *color.Color
}

// constructor to create color objects we can call .Println() on
func newUI() *UI {
	return &UI{
		red:    color.New(color.FgRed),
		blue:   color.New(color.FgBlue),
		green:  color.New(color.FgGreen),
		yellow: color.New(color.FgYellow),
		cyan:   color.New(color.FgCyan),
		bold:   color.New(color.Bold),
	}
}

// showWelcome displays the welcome message
func (ui *UI) showWelcome() {
	ui.clear()
	ui.cyan.Println("╔════════════════════════════════════════╗")
	ui.cyan.Println("║                                        ║")
	ui.cyan.Println("║         🎨 COLOR SYNC GAME 🎨         ║")
	ui.cyan.Println("║                                        ║")
	ui.cyan.Println("║      Real-time Stroop Test Game       ║")
	ui.cyan.Println("║                                        ║")
	ui.cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()
}