package main

import (
	"flag"
	"fmt"
	"fossinator/config"
	"fossinator/fs"
	"fossinator/processor"
)

func init() {
	config.Load("config.yaml")
}

func main() {
	dirFlag := flag.String("dir", "", "Directory to process")
	fmtFlag := flag.Bool("fmt", false, "run 'go fmt' step")
	tidyFlag := flag.Bool("tidy", false, "run 'go mod tidy' step")
	flag.Parse()

	var dir string
	if len(*dirFlag) == 0 {
		dir = "."
	} else {
		dir = *dirFlag
	}

	if _, err := fs.FindGoModFile(dir); err != nil {
		fmt.Printf("Directory '%s' is not a go module", dir)
		return
	}

	fmt.Println("Directory to process: ", dir)

	if err := processor.UpdateImports(dir); err != nil {
		fmt.Println("Error during update imports:", err)
	}

	if err := processor.UpdateGoMod(dir); err != nil {
		fmt.Println("Error during update go.mod:", err)
	}

	if err := processor.AddConfigLoaderConfiguration(dir); err != nil {
		fmt.Println("Error during AddConfigLoaderConfiguration:", err)
	}

	if *fmtFlag {
		processor.RunGoCommand(dir, "fmt", "./...")
	}

	if *tidyFlag {
		processor.RunGoCommand(dir, "mod", "tidy")
	}
}
