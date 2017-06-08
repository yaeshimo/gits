package main

import (
	"io/ioutil"
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
		sub := newSubcmd(ioutil.Discard, ioutil.Discard, test.cmdname, time.Second)
		if err := sub.run(test.args); err != nil {
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
