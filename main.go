package main

//	TODO: impl tests
//		: consider to separate to some functions for to simpl

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	name    = "gits"
	version = "0.1.0dev"
)

// Default values
// TODO: separate to conf_*.go files?
var (
	// TODO: consider change to []string{"$HOME/gits.json"}
	CandidateConfPaths = []string{}
	EditorWithArgs     = []string{"vim", "--"}
)

type option struct {
	version bool
	conf    string
	// TODO: consider to remove
	exec string

	// TODO: consider to remove
	key   string
	match string

	edit  bool
	add   string
	rm    string
	prune bool

	list         bool
	listRepo     bool
	listRepoFull bool
	listAlias    bool
	listConfig   bool

	template bool
}

var opt = &option{}

// Edit edit configuration file
func Edit(w, errw io.Writer, r io.Reader, path string) error {
	if len(EditorWithArgs) < 1 {
		return fmt.Errorf("invalid [EditorWithArgs]: %v", EditorWithArgs)
	}
	editor := exec.Command(EditorWithArgs[0], append(EditorWithArgs[1:], path)...)
	editor.Stdout = w
	editor.Stderr = errw
	editor.Stdin = r
	if _, err := fmt.Fprintln(w, editor.Args); err != nil {
		return err
	}
	return editor.Run()
}

func init() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("[" + name + "]:")
	flag.BoolVar(&opt.version, "version", false, "show version")
	flag.StringVar(&opt.conf, "config", "", "specify path to configuration JSON format files")
	flag.StringVar(&opt.exec, "exec", "git", "specify execute command name")

	flag.StringVar(&opt.key, "key", "", "specify repository name for append repository")
	flag.StringVar(&opt.match, "match", "", "match for pick repostories")

	flag.BoolVar(&opt.edit, "edit", false, "edit config")
	flag.StringVar(&opt.add, "add", "", "specify path to directory for add to configuration files")
	flag.StringVar(&opt.rm, "rm", "", "specify key to remove from configuration file")
	flag.BoolVar(&opt.prune, "prune", false, "prune invalid worktree from configuration file")

	flag.BoolVar(&opt.list, "list", false, "show content of configuration file")
	flag.BoolVar(&opt.listRepo, "list-repo", false, "list repositories")
	flag.BoolVar(&opt.listRepoFull, "list-repo-full", false, "list repositories with full path")
	flag.BoolVar(&opt.listAlias, "list-alias", false, "list alias")
	flag.BoolVar(&opt.listConfig, "list-config", false, "list candidate paths to the configuration file")

	flag.BoolVar(&opt.template, "template", false, "show configuration template")
}

// TODO: error handling for fmt
func main() {
	flag.Parse()
	if opt.version {
		fmt.Fprintf(os.Stdout, "%s version %s\n", name, version)
		os.Exit(0)
	}
	if opt.conf == "" {
		for _, path := range CandidateConfPaths {
			if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
				opt.conf = path
			}
		}
	}
	gits, err := ReadJSON(opt.conf)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: consider to split to functions from flags
	// 1. run
	// 2. check error
	// 3. output err or valid message
	switch {
	case opt.add != "":
		root, err := GetGitToplevel(opt.add)
		if err != nil {
			log.Fatal(err)
		}
		if err := gits.AddRepository(opt.key, root); err != nil {
			log.Fatal(err)
		}
		if err := gits.Update(); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Appended Repositories:\n\t[%s]\nCurrent List:\n", root)
		if err := gits.FprintIndent(os.Stdout, "", "\t"); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Updated:\n\t[%s]\n", gits.path)
	case opt.edit:
		if err := Edit(os.Stdout, os.Stderr, os.Stdin, opt.conf); err != nil {
			log.Fatal(err)
		}
	case opt.rm != "":
		if err := gits.RemoveRepository(opt.rm); err != nil {
			log.Fatal(err)
		}
		if err := gits.Update(); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Removed Repositories:\n\t[%s]\nCurrent List:\n", opt.rm)
		if err := gits.FprintIndent(os.Stdout, "", "\t"); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Updated:\n\t[%s]\n", gits.path)
	case opt.prune:
		if removed, err := gits.Prune(); err != nil {
			log.Fatal(err)
		} else if len(removed) != 0 {
			fmt.Fprintf(os.Stdout, "Pruned:\n\t\"%s\"\n", strings.Join(removed, "\n\t"))
		}
		if err := gits.Update(); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Current List:\n")
		if err := gits.FprintIndent(os.Stdout, "", "\t"); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "Updated:\n\t[%s]\n", gits.path)
	case opt.list:
		if err := gits.FprintIndent(os.Stdout, "", "\t"); err != nil {
			log.Fatal(err)
		}
	case opt.listRepo:
		gits.ListRepositories(os.Stdout)
	case opt.listRepoFull:
		gits.ListRepositoriesFull(os.Stdout)
	case opt.listAlias:
		if err := gits.ListAlias(os.Stdout, opt.exec); err != nil {
			log.Fatal(err)
		}
	case opt.listConfig:
		fmt.Fprintf(os.Stdout, "Candidates:\n[high priority]\n")
		for i, s := range CandidateConfPaths {
			fmt.Fprintf(os.Stdout, "\t%d. %s\n", i+1, s)
		}
		fmt.Fprintln(os.Stdout, "[low priority]")
	case opt.template:
		b, err := Template()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", string(b))
	default:
		args := append([]string{opt.exec}, flag.Args()...)
		os.Exit(gits.Run(os.Stdout, os.Stderr, os.Stdin, opt.match, args))
	}
}
