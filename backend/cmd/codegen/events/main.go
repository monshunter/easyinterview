package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	eventsPath := flag.String("events", "../shared/events.yaml", "path to shared/events.yaml")
	jobsPath := flag.String("jobs", "../shared/jobs.yaml", "path to shared/jobs.yaml")
	repoRoot := flag.String("repo-root", "..", "repository root")
	verbose := flag.Bool("v", false, "print rendered files")
	flag.Parse()

	if err := Run(*eventsPath, *jobsPath, *repoRoot, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "codegen-events: %v\n", err)
		os.Exit(1)
	}
}
