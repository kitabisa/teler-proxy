package main

import (
	"github.com/charmbracelet/log"
	"github.com/kitabisa/teler-proxy/internal/runner"
)

func main() {
	opt := runner.ParseOptions()

	if err := opt.Validate(); err != nil {
		log.Fatal("Cannot validate options", "msg", err)
	}

	if err := runner.New(opt); err != nil {
		log.Fatal("Something went wrong", "err", err)
	}
}
