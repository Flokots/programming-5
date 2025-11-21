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
	magenta *color.Color
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
		magenta: color.New(color.FgMagenta),
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

// showGameStart displays the game start information
func (ui *UI) showGameStart(maxRounds int) {
	ui.clear()
	ui.bold.Println("🎮 GAME STARTING!")
	fmt.Println()
	ui.cyan.Printf("  You will play %d rounds\n", maxRounds)
	ui.cyan.Println(" Match the COLOR of the text (not the word!)")
	fmt.Println()
	ui.yellow.Println("  🏆 Winner Determination:")
	ui.yellow.Println("   1. Most rounds won")
	ui.yellow.Println("   2. If tied: Lowest total latency wins")
	ui.yellow.Println("   3. If still tied: It's a draw!")
	fmt.Println()
	ui.magenta.Println("  Controls: r=red  b=blue  g=green  y=yellow")
	fmt.Println()
	ui.cyan.Println("  Get ready...")
}

// showRound displays the Stroop test for the round
func (ui *UI) showRound(round int, word string, textColor string) {
	fmt.Println(strings.Repeat("─", 50))
	ui.bold.Printf("ROUND %d\n", round)
	fmt.Println()

	ui.cyan.Printf("What COLOR is this text? → ")

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
		fmt.Println(word) // Fallback to default color
	}

	fmt.Println()
	ui.yellow.Print("Your answer [r/b/g/y]: ")
}

// showRoundResult displays the result of a round
func (ui *UI) showRoundResult(round int, winner string, myUserID string, latency int64) {
	fmt.Println()

	if winner == "timeout" {
		ui.yellow.Println("⏱️  Time's up! No one answered in time.")
	} else if winner == myUserID {
		ui.green.Printf("✅ You won this round! (%dms)\n", latency)
	} else {
		ui.red.Println("❌ Opponent won this round!")
	}
}

// showGameOver displays the final game results
func (ui *UI) showGameOver(winner string, myUserID string, wins, opponentWins int, totalLatency, avgLatency int64) {
	ui.clear()
	fmt.Println()
	ui.bold.Println("🏁 GAME OVER!")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	// Determine result from winner
	if winner == "draw" {
		ui.yellow.Println("  🤝 It's a DRAW!")
	} else if winner == myUserID {
	ui.green.Println("  🎉 YOU WON! 🎉")
	} else {
		ui.red.Println("  😞 YOU LOST. 😞")
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))

	// Show stats
	ui.cyan.Println("\n📊 Your Stats:")
	fmt.Printf("  Rounds Won: %d\n", wins)
	fmt.Printf(". Rounds Lost: %d\n", opponentWins)
	fmt.Printf("  Total Latency: %dms\n", totalLatency)

	if wins > 0 {
		fmt.Printf("  Average Latency: %dms\n", avgLatency)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
}

// showInfo displays an info message in cyan
func (ui *UI) showInfo(message string) {
	ui.cyan.Println(message)
}

// showSuccess displays a success message in green
func (ui *UI) showSuccess(message string) {
	ui.green.Println(message)
}

// showError displays an error message in red
func (ui *UI) showError(message string) {
	ui.red.Println(message)
}

// clear clears the terminal screen
func (ui *UI) clear() {
	fmt.Print("\033[H\033[2J")
}
