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
	version      bool
	template     bool
	list         bool
	showConfPath bool

	// TODO: consider add flags
	// logfile string // specify output logfile

	watch   string /// add watch to conf
	unwatch string /// delte watch to conf

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
			premsg := fmt.Sprintf("\n[%s]\n", key)
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
	flags.BoolVar(&opt.list, "list", false, "list of accept first argument and repository")
	flags.BoolVar(&opt.showConfPath, "conf-path", false, "show default conf path")

	flags.StringVar(&opt.watch, "watch", "", "add watching repository to conf")
	flags.StringVar(&opt.unwatch, "unwatch", "", "remove watching repository in conf")

	// setting
	flags.StringVar(&opt.git, "git", "git", "command name of git or full path")
	flags.StringVar(&opt.conf, "conf", defConfPath, "path to json format watchlist")
	flags.DurationVar(&opt.timeout, "timeout", time.Minute*30, "set timeout for running git")
	flags.Parse(args[1:])

	if opt.showConfPath {
		fmt.Fprintln(w, defConfPath)
		return validExit
	}

	wl := &watchList{}
	var err error
	if opt.conf != "" {
		wl, err = readWatchList(opt.conf)
		if err != nil {
			fmt.Fprintln(errw, err)
			return exitWithErr
		}
	}

	// be graceful
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
			fmt.Fprintf(w, "conf:[%s]\n%s\n", opt.conf, wl)
			return validExit
		case opt.watch != "":
			fullpath, key, err := keyAbs(opt.watch)
			if err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			if err := wl.watch(fullpath, key); err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			if err := wl.writeFile(opt.conf); err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			fmt.Fprintf(w, "conf:[%s]\n%s\n", opt.conf, wl)
			fmt.Fprintf(w, "appended [%s] in [%s]\n", key, opt.conf)
			return validExit
		case opt.unwatch != "":
			_, key, err := keyAbs(opt.unwatch)
			if err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			if err := wl.unwatch(key); err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			if err := wl.writeFile(opt.conf); err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			fmt.Fprintf(w, "conf:[%s]\n%s\n", opt.conf, wl)
			fmt.Fprintf(w, "removed [%s] in [%s]\n", key, opt.conf)
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
		fmt.Fprintln(errw, "---------- found error ----------")
		fmt.Fprintln(errw, errs)
		return exitWithErr
	}
	return validExit
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Stdin, os.Args))
}
