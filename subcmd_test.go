package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
	"time"
)

func TestSubcmd(t *testing.T) {
	tests := []struct {
		cmdname string
		args    []string
		wanterr bool
	}{
		{
			cmdname: "",
			args:    []string{},
			wanterr: true,
		},
		{
			cmdname: "git",
			args:    []string{"version"},
			wanterr: false,
		},
	}

	for i, test := range tests {
		sub := NewSubcmd(ioutil.Discard, ioutil.Discard, nil, test.cmdname, time.Second)
		if err := sub.Run("", test.args); err != nil {
			if test.wanterr {
				t.Logf("t.Logf err: %+v", err)
			} else {
				t.Errorf("t.Errorf [%d] err: %+v", i, err)
			}
		} else if test.wanterr {
			t.Errorf("expected error but nil")
		}
	}
}

func BenchmarkSubcmd(b *testing.B) {
	b.Run("goroutine", func(b *testing.B) {
		var s, errs string
		buf := bytes.NewBufferString(s)
		errbuf := bytes.NewBufferString(errs)
		git := NewSubcmd(buf, errbuf, nil, "git", time.Hour)
		args := []string{"version"}

		wg := new(sync.WaitGroup)
		once := new(sync.Once)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if err := git.Run(fmt.Sprintln(i), args); err != nil {
					once.Do(func() { b.Fatal(err) })
				}
			}(i)
		}
		wg.Wait()
	})

	b.Run("single", func(b *testing.B) {
		var s, errs string
		buf := bytes.NewBufferString(s)
		errbuf := bytes.NewBufferString(errs)
		git := NewSubcmd(buf, errbuf, nil, "git", time.Hour)
		args := []string{"version"}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := git.Run(fmt.Sprintln(i), args); err != nil {
				b.Fatal(err)
			}
		}
	})
}
