package main

import (
	"context"
	"io"
	"os/exec"
	"time"
)

type subcmd struct {
	name  string
	limit time.Duration
	w     io.Writer
	errw  io.Writer
}

func (sub *subcmd) run(args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), sub.limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, sub.name, args...)
	cmd.Stdout = sub.w
	cmd.Stderr = sub.errw
	return cmd.Run()
}

func newSubcmd(w io.Writer, errw io.Writer, cmdName string, delay time.Duration) *subcmd {
	return &subcmd{
		name:  cmdName,
		limit: delay,
		w:     w,
		errw:  errw,
	}
}
