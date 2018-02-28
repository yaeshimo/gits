package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// valid "gits.json"
const ValidData = `{
	"allow_commands": {
		"git": {
			"diff": [ "diff", "--stat" ],
			"fetch": [ "fetch" ],
			"log": [ "log", "-p" ],
			"ls": [ "ls-files" ],
			"status": [ "-c", "color.status=always", "status" ]
		}
	},
	"repositories": {}
}`

// invalid "gits.json"
const InvalidData = `{}`

// test directory root
const TestDir = "t"

// TODO: consider to remove
func makeGitRepository(t *testing.T) (gitdir string) {
	tdir := filepath.Join(TestDir, t.Name())
	if err := os.MkdirAll(tdir, 0777); err != nil {
		t.Fatal(err)
	}
	git := exec.Command("git", "init")
	git.Dir = tdir
	git.Stderr = nil
	git.Stdout = nil
	git.Stdin = nil
	if err := git.Run(); err != nil {
		t.Fatal(err)
	}
	return
}

// make "gits.json" return path to test directory and "gits.json"
// TODO: remove tdir?
func makeConf(t *testing.T, wantValid bool) (tdir string, tjson string) {
	tdir = filepath.Join(TestDir, t.Name())
	if err := os.MkdirAll(tdir, 0777); err != nil {
		t.Fatal(err)
	}
	var data string
	if wantValid {
		data = ValidData
	} else {
		data = InvalidData
	}
	tjson = filepath.Join(tdir, "gits.json")
	if err := ioutil.WriteFile(tjson, []byte(data), 0666); err != nil {
		t.Fatal(err)
	}
	return
}

// TODO: validate
func TestReadJSON(t *testing.T) {
	_, tjson := makeConf(t, true)
	gits, err := ReadJSON(tjson)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gits:%#v", gits)
	t.Log("commands")
	for key, args := range gits.AllowCommands {
		t.Log("name:", key, "args:", args)
	}
	t.Log("repositories")
	for key, rep := range gits.Repositories {
		t.Log(key, rep)
	}
}

// TODO: validate
func TestGitsFprintIndent(t *testing.T) {
	_, tjson := makeConf(t, true)
	gits, err := ReadJSON(tjson)
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBufferString("")
	if err := gits.FprintIndent(buf, "with indent", "\t"); err != nil {
		t.Fatal(err)
	}
	t.Logf("buf: %s", buf)
}

// TODO: validate
func TestGitsAdd(t *testing.T) {
	_, tjson := makeConf(t, true)
	gits, err := ReadJSON(tjson)
	if err != nil {
		t.Fatal(err)
	}

	gitdir := makeGitRepository(t)

	t.Logf("before:%v\n", gits)
	if err := gits.AddRepository("", gitdir); err != nil {
		t.Fatal(err)
	}
	t.Logf("after:%v\n", gits)
}

// TODO: validate
func TestGitsListRepositories(t *testing.T) {
	_, tjson := makeConf(t, true)
	gits, err := ReadJSON(tjson)
	if err != nil {
		t.Fatal(err)
	}

	gitdir := makeGitRepository(t)

	// TODO: consider
	if err := gits.AddRepository("", gitdir); err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer([]byte{})
	if err := gits.ListRepositories(buf); err != nil {
		t.Fatal(err)
	}
	t.Log(buf)
}

// TODO: vaildate
func TestStatus(t *testing.T) {
	_, tjson := makeConf(t, true)
	gits, err := ReadJSON(tjson)
	if err != nil {
		t.Fatal(err)
	}

	gitdir := makeGitRepository(t)

	// TODO: consider
	if err := gits.AddRepository("", gitdir); err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBufferString("")
	for key, rep := range gits.Repositories {
		if err := rep.Exec(buf, buf, nil, "git", []string{"status"}); err != nil {
			t.Error(err)
		}
		t.Logf("[Test key:%s]\n%s", key, buf)
		buf.Reset()
	}
}

// TODO: validate
func TestGitsValidArgs(t *testing.T) {
	_, tjson := makeConf(t, true)
	gits, err := ReadJSON(tjson)
	if err != nil {
		t.Fatal(err)
	}

	s1 := []string{"git", "status"}
	t.Log(s1)
	t.Log(gits.ParseArgs(s1[0], s1[1]))

	s2 := []string{"git", "invalid"}
	t.Log(s2)
	t.Log(gits.ParseArgs(s2[0], s2[1]))
}

// TODO: validate
func TestGitsRun(t *testing.T) {
	_, tjson := makeConf(t, true)
	gits, err := ReadJSON(tjson)
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBufferString("")

	gitdir := makeGitRepository(t)

	// TODO: consider
	if err := gits.AddRepository("", gitdir); err != nil {
		t.Fatal(err)
	}

	if err := gits.Run(buf, buf, nil, "git", "status"); err != nil {
		t.Fatal(err)
	}
	t.Log(buf)

	buf.Reset()

	if err := gits.Run(buf, buf, nil, "git", "invalid"); err == nil {
		t.Fatal("expected return error but nil")
	}
	t.Log(buf)
}
