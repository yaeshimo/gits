// git wrapper
// Quick Usage:
//   `gits -template > /path/a/watchlist.json`
// edit gits.json, append your repository
// after append
//   `gits -conf-dir=/path/a -conf=watchlist.json status`
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	showConfDirs bool

	// TODO: add flags?
	// logfile string // specify output logfile
	// execute string // -e [command]

	watch   string /// add watch to conf
	unwatch string /// delte watch to conf

	git     string
	key     string
	conf    string
	confdir string
	timeout time.Duration
}

// repository walker
// use: git --git-dir=/path/to/work/.git --work-tree=/path/to/work
func gitWalker(git *subcmd, wl *watchList, args []string) []error {
	// work on current directory
	// need it?
	if wl.Map == nil || len(wl.Map) == 0 {
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
		argsWithRepo := append([]string{
			"--git-dir=" + repoInfo.Gitdir,
			"--work-tree=" + repoInfo.Workdir},
			args...)
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

func run(w io.Writer, errw io.Writer, r io.Reader, args []string) int {
	opt := option{}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.SetOutput(errw)

	// one shot
	flags.BoolVar(&opt.version, "version", false, "")
	flags.BoolVar(&opt.template, "template", false, "output the template of watchlist")
	flags.BoolVar(&opt.list, "list", false, "list of accept first argument and repository")
	flags.BoolVar(&opt.showConfPath, "conf-path", false, "show default conf path")
	flags.BoolVar(&opt.showConfDirs, "candidate-dirs", false, "show candidate conf directories")

	// modify conf
	flags.StringVar(&opt.watch, "watch", "", "add watching repository to conf")
	flags.StringVar(&opt.unwatch, "unwatch", "", "remove watching repository in conf")

	// setting
	flags.StringVar(&opt.git, "git", "git", "command name of git or full path")
	flags.StringVar(&opt.key, "key", "", "specify target repository")
	flags.StringVar(&opt.conf, "conf", defConfName, "accept base name or full path, to json format watchlist")
	flags.StringVar(&opt.confdir, "conf-dir", defConfDir, "specify conf directory")
	flags.DurationVar(&opt.timeout, "timeout", time.Minute*30, "set timeout for running git")
	flags.Parse(args[1:])

	var confpath string
	if opt.conf != "" {
		if filepath.IsAbs(opt.conf) {
			confpath = opt.conf
		} else {
			confpath = filepath.Join(opt.confdir, filepath.Base(opt.conf))
		}
	}

	wl := &watchList{Map: make(map[string]repoInfo)}
	if confpath != "" {
		var err error
		wl, err = readWatchList(confpath)
		if err != nil {
			fmt.Fprintln(errw, err)
			return exitWithErr
		}
	}

	if opt.key != "" {
		info, ok := wl.Map[opt.key]
		if !ok {
			fmt.Fprintf(errw, "not found [%s] in repository map", opt.key)
			return exitWithErr
		}
		wl.Map = map[string]repoInfo{opt.key: info}
	}

	if flags.NArg() == 0 {
		switch {
		case opt.version:
			fmt.Fprintf(w, "version %s\n", version)
			return validExit
		case opt.showConfPath:
			fmt.Fprintln(w, confpath)
			return validExit
		case opt.showConfDirs:
			fmt.Fprintln(w, strings.Join(defConfDirList, "\n"))
			return validExit
		case opt.template:
			if err := template(w); err != nil {
				// unreachable?
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			return validExit
		case opt.list:
			fmt.Fprintf(w, "conf:[%s]\n%s\n", confpath, wl)
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
			if err := wl.writeFile(confpath); err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			fmt.Fprintf(w, "conf:[%s]\n%s\n", confpath, wl)
			fmt.Fprintf(w, "appended [%s] in [%s]\n", key, confpath)
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
			if err := wl.writeFile(confpath); err != nil {
				// maybe unraechable
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			fmt.Fprintf(w, "conf:[%s]\n%s\n", confpath, wl)
			fmt.Fprintf(w, "removed [%s] in [%s]\n", key, confpath)
			return validExit
		default:
			flags.Usage()
			return exitWithErr
		}
	}

	if !wl.isAllow(flags.Arg(0)) {
		fmt.Fprintf(errw, "Configuration file path:\n\t[%s]\n%s\n", confpath, wl)
		fmt.Fprintf(errw, "This argument is not allowd: %+v\n", flags.Args())
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
