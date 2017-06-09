package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

// TODO: delete this, be graceful
var TestGitDir = ""
var TestNotGitDir = ""
var TestRoot = ""
var TestPWD = ""

// TODO: delete this, be graceful
func cd(dir string) {
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
}

// TODO: be graceful
func TestMain(m *testing.M) {
	exitCode := func() int {
		git, err := exec.LookPath("git")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		TestPWD, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		mktmpdir := func(root string, prefix string) string {
			s, err := ioutil.TempDir(root, prefix)
			if err != nil {
				panic(err)
			}
			return s
		}
		TestRoot = mktmpdir("", "gits_test")
		defer os.RemoveAll(TestRoot)
		TestGitDir = mktmpdir(TestRoot, "gitdir")
		TestNotGitDir = mktmpdir(TestRoot, "notgitdir")

		cd(TestGitDir)
		if err := exec.Command(git, "init").Run(); err != nil {
			panic(err)
		}
		cd(TestPWD)
		return m.Run()
	}()
	os.Exit(exitCode)
}

func TestRun(t *testing.T) {
	// TODO: be graceful
	cd(TestGitDir)
	defer cd(TestPWD)

	tests := []struct {
		args    []string
		wanterr bool
	}{
		{
			args:    []string{"gits", "version"},
			wanterr: false,
		},
		{
			args:    []string{"gits", "status"},
			wanterr: false,
		},
		{
			args:    []string{"gits", "-version"},
			wanterr: false,
		},
		{
			args:    []string{"gits", `-git=""`, "version"},
			wanterr: true,
		},
		{
			args:    []string{"gits", "status", "--invalid--git--flags"},
			wanterr: true,
		},
		{
			args:    []string{"gits"},
			wanterr: true,
		},
		{
			args:    []string{"gits", "not impl"},
			wanterr: true,
		},
	}

	var s, errs string
	buf := bytes.NewBufferString(s)
	errbuf := bytes.NewBufferString(errs)
	for i, test := range tests {
		exitCode := run(buf, errbuf, nil, test.args)
		switch exitCode {
		case validExit:
			if test.wanterr {
				t.Errorf("t.Errorf [%d] expected error but nil", i)
			}
		case exitWithErr:
			if test.wanterr {
				t.Logf("t.Logf [%d] passed error: %+v", i, errbuf)
			} else {
				t.Errorf("t.Errorf [%d] err: %+v", i, errbuf)
			}
		}
		t.Logf("t.Logf [%d] outbuf: %+v", i, buf)
		buf.Reset()
		errbuf.Reset()
	}
}
