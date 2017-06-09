package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

type subcmd struct {
	name  string
	limit time.Duration

	rwmux *sync.RWMutex
	w     io.Writer
	errw  io.Writer
	r     io.Reader
}

func (sub *subcmd) run(premsg string, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), sub.limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, sub.name, args...)
	cmd.Stdout = sub.w
	cmd.Stderr = sub.errw
	cmd.Stdin = sub.r

	sub.rwmux.Lock()
	defer sub.rwmux.Unlock()
	fmt.Fprint(sub.w, premsg)
	return cmd.Run()
}

func newSubcmd(w io.Writer, errw io.Writer, r io.Reader, cmdName string, delay time.Duration) *subcmd {
	return &subcmd{
		name:  cmdName,
		limit: delay,

		rwmux: new(sync.RWMutex),
		w:     w,
		errw:  errw,
		r:     r,
	}
}

// TODO: really need?
func (sub *subcmd) WriteString(s string) (int, error) {
	sub.rwmux.Lock()
	defer sub.rwmux.Unlock()
	return fmt.Fprintln(sub.w, s)
}
func (sub *subcmd) WriteErrString(s string) (int, error) {
	sub.rwmux.Lock()
	defer sub.rwmux.Unlock()
	return fmt.Fprintln(sub.errw, s)
}
