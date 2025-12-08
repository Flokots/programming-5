package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func main() {
	// Parse command-line flags
	username := flag.String("username", "", "Your username (optional - will prompt if not provided)")
	flag.Parse()

	// If username not provided via flag, prompt the user for it
	var finalUsername string
	if *username == "" {
		finalUsername = promptForUsername()
	} else {
		finalUsername = *username
	}

	// Create and run client
	client := newClient(finalUsername)

	if err := client.Run(); err != nil {
		log.Fatalf("Client error: %v", err)
	}
}

// promptForUsername asks the usr to enter their username via stdin
func promptForUsername() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("Enter your username: ")

	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// Validate
	if username == "" {
		fmt.Println("Username cannot be empty. Please try again.")
		os.Exit(1)
	}
	return username
}

// Function for password input (hidden)
func promptForPassword() string {
	// Read password without echoing to terminal
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	fmt.Println() // New line after password input

	password := strings.TrimSpace(string(passwordBytes))

	// Validate password
	if len(password) < 6 {
		fmt.Println("Password must be at least 6 characters")
		os.Exit(1)
	}

	return password
}
