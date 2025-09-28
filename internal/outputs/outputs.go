package outputs

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

func PrintColoredMessage(textColor string, message string, args ...any) {
	PrintColoredMessageTo(os.Stdout, textColor, message, args...)
}

func PrintColoredMessageTo(w io.Writer, textColor string, message string, args ...any) {
	var selectedColor color.Attribute
	switch strings.ToLower(textColor) {
	case "green":
		selectedColor = color.FgGreen
	case "yellow":
		selectedColor = color.FgYellow
	case "red":
		selectedColor = color.FgRed
	case "blue":
		selectedColor = color.FgBlue
	case "cyan":
		selectedColor = color.FgCyan
	default:
		selectedColor = color.FgWhite
	}
	colorFunc := color.New(selectedColor).SprintFunc()
	fullMessage := fmt.Sprintf(message, args...)
	fmt.Fprintf(w, "%s\n", colorFunc(fullMessage))
}

func PrintTerminalWideLine(char string) {
	PrintTerminalWideLineTo(os.Stdout, char)
}

func PrintTerminalWideLineTo(w io.Writer, char string) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// fallback to default width
		width = 40
	}
	line := ""
	for i := 0; i < width; i++ {
		line += string(char)
	}
	fmt.Fprintln(w, line)
}
