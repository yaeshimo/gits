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

const version = "0.0.7"

const (
	validExit = iota
	exitWithErr
)

type option struct {
	// one shot
	version  bool
	template bool
	list     bool
	edit     bool

	// setting
	git     string
	sync    bool
	match   string
	conf    string
	confdir string
	timeout time.Duration

	// show contents of conf
	showConfPath bool
	showConfDirs bool
	showConfList bool

	// new conf
	confNew string // create configuration file to confDir

	// TODO: add flags?
	// out string // specify output to file
	// execute string // specify commands: -e [command]

	// modify conf
	watch   string /// add watch to conf
	unwatch string /// delte watch in conf
}

func (opt *option) init(errw io.Writer, args []string, ehandle flag.ErrorHandling) (*flag.FlagSet, error) {
	// TODO: consider need case args == nil?
	f := flag.NewFlagSet(args[0], ehandle)
	f.SetOutput(errw)

	// one shot
	f.BoolVar(&opt.version, "version", false, "")
	f.BoolVar(&opt.template, "template", false, "output the template of watchlist")
	f.BoolVar(&opt.list, "list", false, "list of accept first argument and repository")
	f.BoolVar(&opt.edit, "edit", false, "open conf on your editor(default:"+DefEditor+")")

	// setting
	f.StringVar(&opt.git, "git", "git", "command name of git or full path")
	f.BoolVar(&opt.sync, "sync", false, "run on sync")
	f.StringVar(&opt.match, "match", "", "specify target repositories")
	f.StringVar(&opt.conf, "conf", DefConfName, "accept base name or full path, to json format watchlist")
	f.StringVar(&opt.conf, "c", DefConfName, "alias of [-conf]")
	f.StringVar(&opt.confdir, "conf-dir", DefConfDir, "specify conf directory")
	f.DurationVar(&opt.timeout, "timeout", 0, "set timeout for running git, 0 means no limit")

	// show contents of conf
	f.BoolVar(&opt.showConfPath, "conf-path", false, "show default conf path")
	f.BoolVar(&opt.showConfDirs, "candidate-dirs", false, "show candidate conf directories")
	f.BoolVar(&opt.showConfList, "conf-list", false, "show configuration set list")

	// new conf
	f.StringVar(&opt.confNew, "conf-new", "", "generate new configuration file to conf directory")

	// modify conf
	f.StringVar(&opt.watch, "watch", "", "add watching repository to conf")
	f.StringVar(&opt.unwatch, "unwatch", "", "remove watching repository in conf")
	return f, f.Parse(args[1:])
}

// repository walker
// use: git --git-dir=/path/to/work/.git --work-tree=/path/to/work
// TODO: sync consider to join structure of subcmd
func gitWalker(git *Subcmd, runOnSync bool, wl *watchList, args []string) []error {
	// work on current directory
	// TODO: need it? case len(w.Map) == 0
	if len(wl.Map) == 0 {
		msg := fmt.Sprintf("not found git repositories:\n\twork on current directory\n")
		git.WriteErrString(msg)
		if err := git.Run("", args); err != nil {
			return []error{err}
		}
		return nil
	}

	var (
		errs         []error
		mux          = new(sync.Mutex)
		wg           = new(sync.WaitGroup)
		argsWithRepo []string
	)

	var do func(string, []string)
	if runOnSync {
		do = func(key string, argsWithRepo []string) {
			premsg := fmt.Sprintf("\n[%s]\n", key)
			if err := git.Run(premsg, argsWithRepo); err != nil {
				errs = append(errs, fmt.Errorf("[%s]:%+v", key, err))
			}
		}
	} else {
		do = func(key string, argsWithRepo []string) {
			wg.Add(1)
			go func(key string, argsWithRepo []string) {
				defer wg.Done()
				premsg := fmt.Sprintf("\n[%s]\n", key)
				if err := git.Run(premsg, argsWithRepo); err != nil {
					mux.Lock()
					errs = append(errs, fmt.Errorf("[%s]:%+v", key, err))
					mux.Unlock()
				}
			}(key, argsWithRepo)
		}
	}

	for key, repoInfo := range wl.Map {
		argsWithRepo = append(
			[]string{"--git-dir=" + repoInfo.Gitdir,
				"--work-tree=" + repoInfo.Workdir},
			args...,
		)
		do(key, argsWithRepo)
	}
	wg.Wait()
	return errs
}

// open configuration files on editor
func editConf(w, errw io.Writer, r io.Reader, editor, confpath string) error {
	if info, err := os.Stat(confpath); err != nil {
		return err
	} else if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not regular file", confpath)
	}
	sub := NewSubcmd(w, errw, r, editor, 0)
	return sub.Run("", []string{confpath})
}

func run(w io.Writer, errw io.Writer, r io.Reader, args []string) int {
	opt := &option{}
	flags, err := opt.init(errw, args, flag.ExitOnError)
	if err != nil {
		fmt.Fprintln(errw, err)
		return exitWithErr
	}
	var confpath string
	if opt.conf != "" {
		if filepath.IsAbs(opt.conf) {
			confpath = opt.conf
		} else {
			confpath = filepath.Join(opt.confdir, filepath.Base(opt.conf))
		}
	}

	gits := newGits(confpath)

	if confpath != "" {
		var err error
		gits.wl, err = readWatchList(gits.path)
		if err != nil {
			fmt.Fprintln(errw, err)
			return exitWithErr
		}
	}
	if opt.match != "" {
		for key := range gits.wl.Map {
			matched, err := filepath.Match(opt.match, key)
			if err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			if matched {
				continue
			}
			delete(gits.wl.Map, key)
		}
	}

	// one shot
	if flags.NArg() == 0 {
		switch {
		case opt.version:
			fmt.Fprintf(w, "version %s\n", version)
			return validExit
		case opt.showConfPath:
			fmt.Fprintln(w, confpath)
			return validExit
		case opt.showConfDirs:
			fmt.Fprintln(w, strings.Join(DefConfDirList, "\n"))
			return validExit
		case opt.showConfList:
			confList, err := gits.getConfList()
			if err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			fmt.Fprintln(w, strings.Join(confList, "\n"))
			return validExit
		case opt.confNew != "":
			mkpath := filepath.Join(DefConfDir, filepath.Base(opt.confNew))
			if err := createConf(mkpath); err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			fmt.Fprintln(w, "configuration file was written: "+mkpath)
			return validExit
		case opt.template:
			if err := writeTemplate(w); err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			return validExit
		case opt.list:
			fmt.Fprintf(w, "conf:[%s]\n%s\n", confpath, gits.wl)
			return validExit
		case opt.edit:
			if err := editConf(w, errw, r, DefEditor, confpath); err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			return validExit
		case opt.watch != "":
			key, err := gits.watch(opt.watch)
			if err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			fmt.Fprintf(w, "conf:[%s]\n%s\nappended [%s]\n", gits.path, gits.wl, key)
			return validExit
		case opt.unwatch != "":
			key, err := gits.unwatch(opt.unwatch)
			if err != nil {
				fmt.Fprintln(errw, err)
				return exitWithErr
			}
			fmt.Fprintf(w, "conf:[%s]\n%s\nremoved [%s]\n", gits.path, gits.wl, key)
			return validExit
		default:
			flags.Usage()
			return exitWithErr
		}
	}

	if !gits.wl.isAllow(flags.Arg(0)) {
		fmt.Fprintf(errw, "Configuration file path:\n\t[%s]\n%s\n", confpath, gits.wl)
		fmt.Fprintf(errw, "This argument is not allowed: %+v\n", flags.Args())
		return exitWithErr
	}

	git := NewSubcmd(w, errw, r, opt.git, opt.timeout)
	if errs := gitWalker(git, opt.sync, gits.wl, flags.Args()); errs != nil {
		fmt.Fprintln(errw, "---------- found error ----------")
		for _, err := range errs {
			fmt.Fprintln(errw, err)
		}
		return exitWithErr
	}
	return validExit
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Stdin, os.Args))
}
