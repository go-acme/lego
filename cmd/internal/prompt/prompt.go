package prompt

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/go-acme/lego/v5/log"
)

func Confirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(message + " Y/n")

		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Could not read from the console", log.ErrorAttr(err))
		}

		text = strings.Trim(text, "\r\n")
		switch strings.ToUpper(text) {
		case "", "Y":
			return true
		case "N":
			return false
		default:
			log.Warn("Your input was invalid. Please answer with one of Y/y, N/n or by pressing enter.")
		}
	}
}

func Choose(message string, options []string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(message + " " + strings.Join(options, ", "))

		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Could not read from the console", log.ErrorAttr(err))
		}

		text = strings.ToUpper(strings.Trim(text, "\r\n"))

		if slices.Contains(options, text) {
			return text
		}

		log.Warnf(log.LazySprintf(
			"Your input was invalid. Please answer with one of %s or by pressing enter.",
			strings.Join(options, ", "),
		))
	}
}
