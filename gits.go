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
type Gits struct {
	// path to configuration files
	path string

	// execName->alias->cmdLine
	AllowCommands map[string]map[string][]string `json:"allow_commands"`
	Repositories  map[string]Repository          `json:"repositories"`
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
	for key, rep := range gits.Repositories {
		_, err = fmt.Fprintf(w, "[%s]\n\t%s\n", key, rep.WorkTree)
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
		AllowCommands: make(map[string]map[string][]string),
		Repositories:  make(map[string]Repository),
	}
	gits.AllowCommands["git"] = make(map[string][]string)
	gits.AllowCommands["git"]["status"] = []string{"-c", "color.status=always", "status"}
	gits.AllowCommands["git"]["fetch"] = []string{"fetch"}
	gits.AllowCommands["git"]["diff"] = []string{"diff", "--stat"}
	gits.AllowCommands["git"]["ls"] = []string{"ls-files"}
	gits.AllowCommands["ls"] = make(map[string][]string)
	gits.AllowCommands["ls"]["la"] = []string{"-lah", "--color=auto"}
	gits.AllowCommands["pwd"] = make(map[string][]string)
	if err := gits.AddRepository("", ""); err != nil {
		return nil, fmt.Errorf("%v\nRecurire run on the git repository", err)
	}
	return json.MarshalIndent(gits, "", "\t")
}

// ParseArgs if args is invalid then return name is ""
func (gits *Gits) ParseArgs(executable string, alias string) (name string, allowArgs []string) {
	if executable == "" {
		return "", nil
	}
	// check command cmdName
	aliasMap, ok := gits.AllowCommands[executable]
	if !ok {
		return "", nil
	}
	if alias == "" {
		// accept only single command
		return executable, nil
	}
	allowArgs, ok = aliasMap[alias]
	if !ok {
		return "", nil
	}
	return executable, allowArgs
}

// Run args[0] == executable, args[1 == alias
func (gits *Gits) Run(w, errw io.Writer, r io.Reader, executable string, alias string) error {
	name, allowArgs := gits.ParseArgs(executable, alias)
	if name == "" {
		return fmt.Errorf("invalid arguments:%v %v", executable, alias)
	}

	// TODO: remove?
	if name == "git" && len(allowArgs) == 0 {
		return fmt.Errorf("need specify alias. see [gits -list-alias]")
	}

	fmt.Fprintf(w, "exec=[%s] args=%q\n", name, allowArgs)
	var errors []error
	for key, rep := range gits.Repositories {
		fmt.Fprintf(w, "\n[Repository]: %#v\n", key)
		if err := rep.Exec(w, errw, r, name, allowArgs); err != nil {
			errors = append(errors, fmt.Errorf("[%v %v]", key, err))
		}
	}
	if len(errors) != 0 {
		return fmt.Errorf("Errors: %v", errors)
	}
	return nil
}
