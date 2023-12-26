package main

import (
	"debug/elf"
	"flag"
	"log/slog"
	"net/url"
	"os"
)

const (
	envToken = "FELF_TOKEN"
	envUrl   = "FELF_URL"
)

func main() { os.Exit(cli()) }

func cli() int {
	onlyMeasure := flag.Bool("only-measure", false, "Stop after performing measurements.")
	dryRun := flag.Bool("dry-run", false, "Don't push data to the server. Log the payload to stdout.")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		slog.Error("only a single positional argument is supported", "args", len(args))
		return 2
	}

	file, err := elf.Open(args[0])
	if err != nil {
		slog.Error(err.Error())
		return 74
	}

	measurements := newSize(file)
	slog.Info("analysis done", "size", measurements)

	if *onlyMeasure {
		return 0
	}

	payload, err := newPayload()
	if err != nil {
		slog.Error(err.Error())
		return 3
	}
	payload.Size = measurements
	if *dryRun {
		slog.Info("dry mode selected", "payload", *payload)
		return 0
	}

	token := os.Getenv(envToken)
	apiUrl := os.Getenv(envUrl)
	if len(token) == 0 {
		slog.Error("API token missing")
		return 4
	}
	if _, err := url.Parse(apiUrl); err != nil {
		slog.Error(err.Error())
		return 5
	}

	response, err := pushPayload(token, apiUrl, payload)
	if err != nil {
		slog.Error(err.Error())
		return 1
	}
	if response.StatusCode != 200 {
		slog.Error("unexpected status code", "code", response.StatusCode)
		return 1
	}

	slog.Info("successful push", "url", apiUrl)
	return 0
}
