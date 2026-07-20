package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gabrielnakaema/project-chat/internal/architecture"
)

func main() {
	root := flag.String("root", ".", "backend module root")
	flag.Parse()

	violations, err := architecture.CheckImportBoundaries(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(violations) == 0 {
		fmt.Println("import boundaries: ok")
		return
	}

	for _, violation := range violations {
		fmt.Fprintln(os.Stderr, violation.Error())
	}
	os.Exit(1)
}
