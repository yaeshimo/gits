// git wrapper
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

const version = "0.0.0"

const (
	validExit = iota
	exitWithErr
)

type option struct {
	version   bool
	gitname   string
	watchlist string
}

// recursive walker
// use: git --git-dir=/path/to/work/.git --work-tree=/path/to/work
func gitWalker() {
}

func run(w io.Writer, errw io.Writer, args []string) int {
	opt := &option{}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.BoolVar(&opt.version, "version", false, "")
	flags.StringVar(&opt.gitname, "gitname", "git", "")
	flags.StringVar(&opt.watchlist, "watchlist", "", "")
	flags.Parse(args[1:])
	if flags.NArg() == 0 {
		if opt.version {
			fmt.Fprintf(w, "version %s\n", version)
			return validExit
		}
		// TODO: modify default run is to status
		fmt.Fprintln(errw, "not enough argument")
		return exitWithErr
	}
	switch flags.Arg(0) {
	// accept commands passing
	case "status", "version":
	default:
		fmt.Fprintf(errw, "invalid argument: %+v", flags.Args())
		return exitWithErr
	}
	// TODO: fix for multitarget
	git, err := newGit(w, errw, opt.gitname, time.Minute)
	if err != nil {
		fmt.Fprintln(errw, err)
		return exitWithErr
	}
	if err := git.run(flags.Args()); err != nil {
		fmt.Fprintln(errw, err)
		return exitWithErr
	}
	return validExit
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}
