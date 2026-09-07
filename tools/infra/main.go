package main

import (
	"fmt"

	"github.com/fatih/color"
)

// Preserve the upstream attribution without presenting it as this deployment.
func printLinks() {
	blue := color.New(color.FgBlue).SprintFunc()
	fmt.Printf("Upstream: %s\n", blue("https://github.com/openimsdk/open-im-server"))
}

func main() {
	yellow := color.New(color.FgYellow)
	blue := color.New(color.FgBlue, color.Bold)

	yellow.Println("Please use the release branch or tag for production environments!")

	blue.Println("Server")
	printLinks()
}
