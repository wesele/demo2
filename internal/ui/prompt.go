package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm displays a confirmation prompt and returns true if the user presses
// Enter (with an empty line). Returns false on Ctrl+C (EOF/error).
func Confirm() bool {
	fmt.Print("\nRun this command? [Enter to confirm / Ctrl+C to cancel]: ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		// EOF (Ctrl+D) or Ctrl+C signal → do not execute.
		fmt.Println()
		return false
	}

	// Accept an empty Enter press or explicit "y"/"yes".
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "" || line == "y" || line == "yes"
}
