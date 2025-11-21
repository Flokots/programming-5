package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// UI handles terminal display
type UI struct {
	// Color functions
	red    *color.Color
	blue   *color.Color
	green  *color.Color
	yellow *color.Color
	cyan   *color.Color
	bold   *color.Color
}

// newUI creates a new UI instance
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

// ==========================================
// WELCOME SCREEN
// ==========================================

func (ui *UI) showWelcome() {
	ui.clear()
	ui.cyan.Println("╔════════════════════════════════════════╗")
	ui.cyan.Println("║                                        ║")
	ui.cyan.Print("║         ")
	ui.bold.Print("🎨 COLOR SYNC GAME 🎨")
	ui.cyan.Println("         ║")
	ui.cyan.Println("║                                        ║")
	ui.cyan.Println("║      Real-time Stroop Test Game       ║")
	ui.cyan.Println("║                                        ║")
	ui.cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()
}

// ==========================================
// GAME START
// ==========================================

func (ui *UI) showGameStart(maxRounds int) {
	ui.clear()
	ui.bold.Println("🎮 GAME STARTING!")
	fmt.Println()
	ui.cyan.Printf("  First to win %d rounds wins!\n", maxRounds)
	ui.cyan.Println("  Click the COLOR of the text (ignore the word)")
	fmt.Println()
	ui.yellow.Println("  Controls: r=red  b=blue  g=green  y=yellow")
	fmt.Println()
	ui.green.Println("  Get ready...")
	fmt.Println()
}

// ==========================================
// ROUND DISPLAY
// ==========================================

func (ui *UI) showRound(round int, word, textColor string) {
	fmt.Println(strings.Repeat("─", 50))
	ui.bold.Printf("ROUND %d\n", round)
	fmt.Println()

	// Display the Stroop test
	ui.cyan.Print("What COLOR is this text? → ")

	// Print the word in the specified color
	switch textColor {
	case "red":
		ui.red.Println(word)
	case "blue":
		ui.blue.Println(word)
	case "green":
		ui.green.Println(word)
	case "yellow":
		ui.yellow.Println(word)
	default:
		fmt.Println(word)
	}

	fmt.Println()
	ui.yellow.Print("Your answer [r/b/g/y]: ")
}

// ==========================================
// ROUND RESULT
// ==========================================

func (ui *UI) showRoundResult(round int, iWon, isDraw bool, latency int64, myScore, opponentScore int) {
	fmt.Println()

	if isDraw {
		ui.yellow.Println("⏱️  Time's up! No one answered correctly.")
	} else if iWon {
		ui.green.Printf("✅ You won this round! (%dms)\n", latency)
	} else {
		ui.red.Println("❌ Opponent won this round!")
	}

	fmt.Println()
	ui.bold.Printf("Score: YOU %d - %d OPPONENT\n", myScore, opponentScore)
	fmt.Println()
}

// ==========================================
// GAME OVER
// ==========================================

func (ui *UI) showGameOver(iWon, isDraw bool, wins, losses int, totalLatency, avgLatency int64) {
	ui.clear()
	fmt.Println()
	ui.bold.Println("🏁 GAME OVER!")
	fmt.Println()
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println()

	// Show result
	if isDraw {
		ui.yellow.Println("  🤝 It's a DRAW!")
	} else if iWon {
		ui.green.Println("  🎉 YOU WON! 🎉")
	} else {
		ui.red.Println("  😞 You Lost")
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))

	// Show stats
	ui.cyan.Println("\n📊 Your Stats:")
	fmt.Printf("   Rounds Won:      %d\n", wins)
	fmt.Printf("   Rounds Lost:     %d\n", losses)
	fmt.Printf("   Total Latency:   %dms\n", totalLatency)
	if wins > 0 {
		fmt.Printf("   Average Latency: %dms\n", avgLatency)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println()
}

// ==========================================
// UTILITY FUNCTIONS
// ==========================================

func (ui *UI) showInfo(message string) {
	ui.cyan.Println(message)
}

func (ui *UI) showSuccess(message string) {
	ui.green.Println(message)
}

func (ui *UI) showError(message string) {
	ui.red.Println(message)
}

func (ui *UI) clear() {
	fmt.Print("\033[H\033[2J") // ANSI escape code to clear screen
}
