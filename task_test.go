package sushi

import (
	"bytes"
	"testing"
)

func TestRunTask(t *testing.T) {
	id, err := RunTask(".", "cd ..")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(id)
	ob := &bytes.Buffer{}
	eb := &bytes.Buffer{}

loop:
	for {
		select {
		case e := <-exitc:
			t.Logf("%+v", e)
			break loop
		case o := <-stdoutc:
			n, err := ob.Write(o.p)
			o.respc <- writeResp{n, err}
		case e := <-stderrc:
			n, err := eb.Write(e.p)
			e.respc <- writeResp{n, err}
		}
	}

	t.Log(ob.String())
	t.Log(eb.String())
}
