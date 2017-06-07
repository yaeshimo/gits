package main

import (
	"context"
	"io"
	"os/exec"
	"time"
)

type subcmdGit struct {
	fpath string
	limit time.Duration
	w     io.Writer
	errw  io.Writer
}

func (git *subcmdGit) run(args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), git.limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, git.fpath, args...)
	cmd.Stdout = git.w
	cmd.Stderr = git.errw
	return cmd.Run()
}

func newGit(w io.Writer, errw io.Writer, cmdName string, delay time.Duration) (*subcmdGit, error) {
	fpath, err := exec.LookPath(cmdName)
	if err != nil {
		return nil, err
	}
	git := &subcmdGit{
		fpath: fpath,
		limit: delay,
		w:     w,
		errw:  errw,
	}
	return git, nil
}
