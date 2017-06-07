package main

import (
	"io/ioutil"
	"testing"
	"time"
)

func TestGit(t *testing.T) {
	tests := []struct {
		aliasOfGit string
		wanterr    bool
	}{
		{
			aliasOfGit: "",
			wanterr:    true,
		},
		{
			aliasOfGit: "git",
			wanterr:    false,
		},
	}

	for _, test := range tests {
		_, err := newGit(ioutil.Discard, ioutil.Discard, test.aliasOfGit, time.Second)
		switch {
		case err == nil && test.wanterr:
			t.Errorf("expected error but nil")
		case err != nil && !test.wanterr:
			t.Errorf("err: %+v", err)
		default:
			t.Logf("passed err: %+v", err)
		}
	}
}
