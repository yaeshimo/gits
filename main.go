// git wrapper
// Quick Usage:
//   `gits -template > watchlist.json`
// edit gits.json, add your repository
//   `gits -conf=watchlist.json status`
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const version = "0.0.0"

const (
	validExit = iota
	exitWithErr
)

type option struct {
	version  bool
	template bool
	list     bool
	// TODO: consider add flags
	// watch string // add current git repository to watchlist.json
	// unwatch string // remove current git repository on watchlist.json
	// logfile string // specify output logfile

	git     string
	conf    string
	timeout time.Duration
}

// repository walker
// use: git --git-dir=/path/to/work/.git --work-tree=/path/to/work
//    : consider RWMutex write buffer?
func gitWalker(git *subcmd, wl *watchList, args []string) []error {
	// work on current directory
	// TODO: really need?
	if wl.Map == nil {
		msg := fmt.Sprintf("not found git repositories:\n\twork on current directory\n")
		git.WriteErrString(msg)
		if err := git.run("", args); err != nil {
			return []error{err}
		}
		return nil
	}

	var (
		errs []error
		mux  = new(sync.Mutex)
		wg   = new(sync.WaitGroup)
	)

	for key, repoInfo := range wl.Map {
		argsWithRepo := append(
			[]string{
				"--git-dir=" + repoInfo.Gitdir,
				"--work-tree=" + repoInfo.Workdir},
			args...,
		)
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			premsg := fmt.Sprintf("\n%s:\n", key)
			if err := git.run(premsg, argsWithRepo); err != nil {
				mux.Lock()
				errs = append(errs, fmt.Errorf("[%s]:%+v", key, err))
				mux.Unlock()
			}
		}(key)
	}
	wg.Wait()
	return errs
}

// TODO: be graceful
func run(w io.Writer, errw io.Writer, r io.Reader, args []string) int {
	opt := option{}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.SetOutput(errw)

	// one shot
	flags.BoolVar(&opt.version, "version", false, "")
	flags.BoolVar(&opt.template, "template", false, "output template json")
	flags.BoolVar(&opt.list, "list", false, "list accept git commands")

	// setting
	flags.StringVar(&opt.git, "git", "git", "name of git command or fullpath")
	flags.StringVar(&opt.conf, "conf", defConfPath, "path to json format watchlist")
	flags.DurationVar(&opt.timeout, "timeout", time.Hour*12, "set timeout for running git")
	flags.Parse(args[1:])

	wl := &watchList{}
	var err error
	if opt.conf != "" {
		wl, err = readWatchList(opt.conf)
		if err != nil {
			fmt.Fprintln(errw, err)
			return exitWithErr
		}
	}

	if flags.NArg() == 0 {
		switch {
		case opt.version:
			fmt.Fprintf(w, "version %s\n", version)
			return validExit
		case opt.template:
			if err := template(w); err != nil {
				// unreachable?
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			return validExit
		case opt.list:
			fmt.Fprintln(w, "Accept git commands:")
			if wl.Restriction == nil || len(wl.Restriction) == 0 {
				fmt.Fprintln(w, "\t[Allow all git commmands]")
				return validExit
			}
			for _, s := range wl.Restriction {
				fmt.Fprintf(w, "\t[%s]\n", s)
			}
			return validExit
		default:
			flags.Usage()
			return exitWithErr
		}
	}

	if !wl.isAllow(flags.Arg(0)) {
		fmt.Fprintf(errw, "invalid argument: %+v\n", flags.Args())
		return exitWithErr
	}

	git := newSubcmd(w, errw, r, opt.git, opt.timeout)
	if errs := gitWalker(git, wl, flags.Args()); errs != nil {
		fmt.Fprintln(errw, errs)
		return exitWithErr
	}
	return validExit
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Stdin, os.Args))
}
