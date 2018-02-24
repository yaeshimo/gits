package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Repository fields for repository list
// WorkTree is git --work-tree
// TODO: consider join to Gits and remove Git
type Repository struct {
	WorkTree string `json:"worktree"`
}

// Exec run command
// name is cmd name
// args is arguments for cmd name
// w is output target
// TODO: consider add premsg and use mutex?
func (rep *Repository) Exec(w, errw io.Writer, r io.Reader, name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = w
	cmd.Stderr = errw
	cmd.Stdin = r
	cmd.Dir = rep.WorkTree
	return cmd.Run()
}

// Gits is for json configuration files
// AllowCommands is map of allow commands, [exec]->[alias]->args
// Repositories is map to repositories
// TODO: consider join Repository
type Gits struct {
	// path to configuration files
	path string

	// execName->alias->cmdLine
	AllowCommands map[string]map[string]string `json:"allow_commands"`
	Repositories  map[string]Repository        `json:"repositories"`
}

// ReadJSON read json from file
func ReadJSON(path string) (*Gits, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	gits := &Gits{path: path}
	if err := json.Unmarshal(b, gits); err != nil {
		return nil, err
	}
	return gits, nil
}

// FprintIndent write a json with indent to w
func (gits *Gits) FprintIndent(w io.Writer, prefix string, indent string) error {
	b, err := json.MarshalIndent(gits, prefix, indent)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, string(b)); err != nil {
		return err
	}
	return nil
}

// WriteFile write to file
func (gits *Gits) WriteFile(path string) error {
	buf := bytes.NewBuffer([]byte{})
	if err := gits.FprintIndent(buf, "", "\t"); err != nil {
		return err
	}
	return ioutil.WriteFile(path, buf.Bytes(), 0666)
}

// Update update gits.path
func (gits *Gits) Update() error {
	return gits.WriteFile(gits.path)
}

// ParseArgs get cmdname, allowed args and valid state
// TODO: consider to simpl
func (gits *Gits) ParseArgs(args []string) (cmdName string, allowArgs []string, ok bool) {
	n := len(args)
	if n == 0 {
		return "", nil, false
	}
	name := args[0]
	// check command name
	aliasMap, ok := gits.AllowCommands[name]
	if !ok {
		return "", nil, false
	}
	if n == 1 {
		// accept only single command name
		return name, nil, true
	}
	// check alias
	if n != 2 {
		return "", nil, false
	}
	line, ok := aliasMap[args[1]]
	if !ok {
		return "", nil, false
	}
	return name, strings.Fields(line), true
}

// GetGitToplevel depends get the top of worktree
// NOTE: depends on "git"
// TODO: consider
func GetGitToplevel(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = abs
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s\t%s", err, string(b))
	}
	return strings.TrimSpace(string(b)), nil
}

// AddRepository append repository
// TODO: consider to remove the key
func (gits *Gits) AddRepository(key string, path string) error {
	abs, err := GetGitToplevel(path)
	if err != nil {
		return err
	}

	if key == "" {
		key = filepath.Base(abs)
	}
	if _, ok := gits.Repositories[key]; ok {
		return fmt.Errorf("[Repositor]: %#v is already exists in %s", key, gits.path)
	}
	gits.Repositories[key] = Repository{
		WorkTree: abs,
	}
	return nil
}

// RemoveRepository remove repository from Gits.Repositories
func (gits *Gits) RemoveRepository(key string) error {
	if _, ok := gits.Repositories[key]; !ok {
		return fmt.Errorf("not exists: %v", key)
	}
	delete(gits.Repositories, key)
	return nil
}

// RemoveMatchRepositories delete repositories from match strings
func (gits *Gits) RemoveMatchRepositories(match string) error {
	reg, err := regexp.Compile(match)
	if err != nil {
		return err
	}
	for key := range gits.Repositories {
		if !reg.MatchString(key) {
			delete(gits.Repositories, key)
		}
	}
	return nil
}

// Prune for prune invalid worktree from configuraton file
func (gits *Gits) Prune() ([]string, error) {
	var removed []string
	for key, repo := range gits.Repositories {
		if info, err := os.Stat(repo.WorkTree); err != nil || !info.IsDir() {
			if err := gits.RemoveRepository(key); err != nil {
				return nil, err
			}
			removed = append(removed, key)
		}
	}
	return removed, nil
}

// ListRepositories list key of Gits.Repositories
func (gits *Gits) ListRepositories(w io.Writer) error {
	var err error
	for key := range gits.Repositories {
		_, err = fmt.Fprintln(w, key)
	}
	return err
}

// ListRepositoriesFull list key of Gits.Repositories
func (gits *Gits) ListRepositoriesFull(w io.Writer) error {
	var err error
	for _, rep := range gits.Repositories {
		_, err = fmt.Fprintln(w, rep.WorkTree)
	}
	return err
}

// ListAlias list alias
func (gits *Gits) ListAlias(w io.Writer, exec string) error {
	aliasMap, ok := gits.AllowCommands[exec]
	if !ok {
		return fmt.Errorf("%s is do not allow", exec)
	}
	_, err := fmt.Fprintf(w, "[%s]\n", exec)
	for key, line := range aliasMap {
		_, err = fmt.Fprintf(w, "\t\"%s\": \"%s\"\n", key, line)
	}
	return err
}

// Template return template
func Template() ([]byte, error) {
	gits := &Gits{
		AllowCommands: make(map[string]map[string]string),
		Repositories:  make(map[string]Repository),
	}
	gits.AllowCommands["git"] = make(map[string]string)
	gits.AllowCommands["git"]["status"] = "-c color.status=always status"
	gits.AllowCommands["git"]["fetch"] = "fetch"
	gits.AllowCommands["git"]["diff"] = "diff --stat"
	gits.AllowCommands["git"]["ls"] = "ls-files"
	gits.AllowCommands["ls"] = make(map[string]string)
	gits.AllowCommands["pwd"] = make(map[string]string)
	if err := gits.AddRepository("", ""); err != nil {
		return nil, fmt.Errorf("%v\nRecurire run on the git repository", err)
	}
	return json.MarshalIndent(gits, "", "\t")
}

// Run can use
// gits -- git remote set-url origin git@github.com:UserName/$(basename $(pwd)).git
// "allow_commands": {
//   "git": {
//     "seturl": "remote set-url origin git@github.com:${UserName}/$(basename $(pwd)).git"
//   }
// }
func (gits *Gits) Run(w, errw io.Writer, r io.Reader, args []string) int {
	name, allowArgs, ok := gits.ParseArgs(args)
	if !ok {
		fmt.Fprintf(errw, "invalid arguments:%v\n", args)
		return 1
	}
	if name == "git" && len(allowArgs) == 0 {
		fmt.Fprintf(errw, "need specify alias. see [gits -list-alias]\n")
		return 1
	}

	fmt.Fprintf(w, "exec=[%s] args=%v\n", name, allowArgs)
	var exit int
	for key, rep := range gits.Repositories {
		fmt.Fprintf(w, "\n[Repository]: %#v\n", key)
		if err := rep.Exec(w, errw, r, name, allowArgs); err != nil {
			fmt.Fprintln(errw, err)
			exit = 2
		}
	}
	return exit
}
