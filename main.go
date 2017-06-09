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

// TODO: split to outside json file
var acceptFirstGitArgs = []string{
	"status",
	"version",
	"fetch",
	"grep",
	"ls-remote",
	"ls-files",
	"ls-tree",

	"diff",
	"add",
	"commit",
}

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
func gitWalker(git *subcmd, repoMap watchList, args []string) []error {
	// work on current directory
	if repoMap == nil {
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

	for key, repoInfo := range repoMap {
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
	opt := &option{}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	flags.BoolVar(&opt.version, "version", false, "")
	flags.BoolVar(&opt.template, "template", false, "output template json")
	flags.BoolVar(&opt.list, "list", false, "list accept git commands")

	flags.StringVar(&opt.git, "git", "git", "name of git command or fullpath")
	flags.StringVar(&opt.conf, "conf", "", "path to json format watchlist")
	flags.DurationVar(&opt.timeout, "timeout", time.Hour*12, "set timeout for running git")

	flags.Parse(args[1:])
	if flags.NArg() == 0 {
		switch {
		case opt.version:
			fmt.Fprintf(w, "version %s\n", version)
			return validExit
		case opt.template:
			if err := template(w); err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			return validExit
		case opt.list:
			fmt.Fprintln(w, "Accept git commands:")
			for _, s := range acceptFirstGitArgs {
				fmt.Fprintf(w, "\t%s\n", s)
			}
			return validExit
		default:
			// TODO: change default run is to status?
			fmt.Fprintln(w, "Accept git commands:")
			for _, s := range acceptFirstGitArgs {
				fmt.Fprintf(w, "\t%s\n", s)
			}
			//fmt.Fprintln(errw, "Not enough argument")
			return exitWithErr
		}
	}

	// TODO: split to another function
	//     : func(*watchList, fa string, check1, check2 []string) error
	//     : delete keis: fa == in(check2) && readonly == true
	//     : if err != nil { echo errmsg; return exitWithErr }
	accept := false
	subname := flags.Arg(0)
	for _, s := range acceptFirstGitArgs {
		if s == subname {
			accept = true
			break
		}
	}
	if !accept {
		fmt.Fprintf(errw, "invalid argument: %+v\n", flags.Args())
		return exitWithErr
	}

	// TODO: move position
	var repoMap watchList
	var err error
	if opt.conf != "" {
		repoMap, err = readWatchList(opt.conf)
		if err != nil {
			fmt.Fprintln(errw, err)
			return exitWithErr
		}
	}

	git := newSubcmd(w, errw, r, opt.git, opt.timeout)
	if errs := gitWalker(git, repoMap, flags.Args()); errs != nil {
		fmt.Fprintln(errw, errs)
		return exitWithErr
	}
	return validExit
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Stdin, os.Args))
}
