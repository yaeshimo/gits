package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// Subcmd for subcommand
type Subcmd struct {
	name  string
	limit time.Duration // 0 means no limit

	rwmux   *sync.RWMutex
	w, errw io.Writer
	r       io.Reader
}

// NewSubcmd create new Subcmd
func NewSubcmd(w io.Writer, errw io.Writer, r io.Reader, cmdName string, delay time.Duration) *Subcmd {
	return &Subcmd{
		name:  cmdName,
		limit: delay,

		rwmux: new(sync.RWMutex),
		w:     w,
		errw:  errw,
		r:     r,
	}
}

// Run running sub commands
func (sub *Subcmd) Run(premsg string, args []string) error {
	var cmd *exec.Cmd
	if sub.limit != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), sub.limit)
		defer cancel()
		cmd = exec.CommandContext(ctx, sub.name, args...)
	} else {
		cmd = exec.Command(sub.name, args...)
	}
	cmd.Stdout = sub.w
	cmd.Stderr = sub.errw
	cmd.Stdin = sub.r

	sub.rwmux.Lock()
	defer sub.rwmux.Unlock()
	fmt.Fprint(sub.w, premsg)
	return cmd.Run()
}

// WriteString TODO: really need?
func (sub *Subcmd) WriteString(s string) (int, error) {
	sub.rwmux.Lock()
	defer sub.rwmux.Unlock()
	return fmt.Fprintln(sub.w, s)
}

// WriteErrString write to *subcmd.err
func (sub *Subcmd) WriteErrString(s string) (int, error) {
	sub.rwmux.Lock()
	defer sub.rwmux.Unlock()
	return fmt.Fprintln(sub.errw, s)
}
