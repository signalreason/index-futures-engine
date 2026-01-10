package main

import (
	"fmt"
	"os"

	"trading-algo-generator/internal/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
