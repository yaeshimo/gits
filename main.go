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
	version bool
	git     string
	conf    string
}

// recursive walker
// use: git --git-dir=/path/to/work/.git --work-tree=/path/to/work
func gitWalker(conf string, gitname string, args []string, timeout time.Duration) error {
	// TODO: be implement

	_, err := readWatchList(conf)
	if err != nil {
		return err
	}
	return nil
}

func run(w io.Writer, errw io.Writer, args []string) int {
	opt := &option{}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.BoolVar(&opt.version, "version", false, "")
	flags.StringVar(&opt.git, "git", "git", "name of git command or fullpath")
	flags.StringVar(&opt.conf, "conf", "", "")
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
	case "status", "version":
		// accept commands passing
	default:
		fmt.Fprintf(errw, "invalid argument: %+v\n", flags.Args())
		return exitWithErr
	}

	// TODO: fix for multitarget
	// gitWalker()
	git := newSubcmd(w, errw, opt.git, time.Minute)
	if err := git.run(flags.Args()); err != nil {
		fmt.Fprintln(errw, err)
		return exitWithErr
	}

	return validExit
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}
