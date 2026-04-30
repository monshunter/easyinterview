package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	eventsPath := flag.String("events", "../shared/events.yaml", "path to shared/events.yaml")
	jobsPath := flag.String("jobs", "../shared/jobs.yaml", "path to shared/jobs.yaml")
	conventionsPath := flag.String("conventions", "../shared/conventions.yaml", "path to shared/conventions.yaml")
	repoRoot := flag.String("repo-root", "..", "repository root")
	verbose := flag.Bool("v", false, "print rendered files")
	flag.Parse()

	if err := RunWithConventions(*eventsPath, *jobsPath, *conventionsPath, *repoRoot, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "codegen-events: %v\n", err)
		os.Exit(1)
	}
}
