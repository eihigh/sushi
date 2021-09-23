package sushi

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell/v2"
)

var (
	taskViews []*taskView

	screen  tcell.Screen
	cursor  = -1 // taskID
	prompt  = []rune{}
	state   = promptView
	workdir = ""
	lastCD  = -1 // この値よりexitMsg.taskIDが大きければwd更新

	boxes struct {
		prompt, history, stdout, stderr, tips *box
	}
)

type viewState int

const (
	promptView viewState = iota
	attachView
)

type taskView struct {
	command string
	// taskがchannelから送ってくるデータを貯める場所。後でlibvtermを使う
	exit     bool
	exitCode int

	stdout bytes.Buffer
	stderr bytes.Buffer

	// vtOut *vterm.VTerm
	// vtErr *vterm.VTerm
	// out   *vterm.Screen
	// err   *vterm.Screen
}

func newTaskView(command string) (*taskView, error) {
	t := &taskView{
		// vtOut: vterm.New(boxes.stdout.h, boxes.stdout.w),
		// vtErr: vterm.New(boxes.stderr.h, boxes.stderr.w),
	}
	// t.out = t.vtOut.ObtainScreen()
	// t.err = t.vtErr.ObtainScreen()

	return t, nil
}

func layout(width, height int) {
	/*
		--[prompt]--------
		/path/to/workdir
		$ command; command
		--[history]-------
		_ EXIT 0  | pwd
		>>RUNNING | fzf
		--[stdout]--------
		...

		--[stderr]--------
		...
		--[tips]----------
		(ctrl+z) detach
	*/
	boxes.prompt = &box{0, 0, width, 3}
	height -= 3 /*prompt*/ + 2 /*tips*/
	fh := float64(height)
	hh := int(fh / 4)
	oh := int(fh / 2)
	eh := height - hh - oh // 余り

	y := 3
	boxes.history = &box{0, y, width, hh}
	y += hh
	boxes.stdout = &box{0, y, width, oh}
	y += oh
	boxes.stderr = &box{0, y, width, eh}
	y += eh
	boxes.tips = &box{0, y, width, 2}
}

func draw() {
	s := fmt.Sprintf("--[prompt]--\n%s\n$ %s", workdir, string(prompt))
	boxes.prompt.drawString(s, true)

	s = ""
	b := len(taskViews) - boxes.history.h
	if b < 0 {
		b = 0
	}
	for i, t := range taskViews[b:] {
		c := "  "
		if cursor == i {
			if state == attachView {
				c = ">>"
			} else {
				c = "> "
			}
		}
		r := "RUNNING"
		if t.exit {
			r = fmt.Sprintf("EXIT %d", t.exitCode)
		}
		l := fmt.Sprintf("%s%s | %s\n", c, r, t.command)
		s = l + s
	}
	boxes.history.drawString("--[history]--\n"+s, false)

	if cursor >= 0 {
		t := taskViews[cursor]
		// boxes.stdout.drawVTerm(t.vtOut, t.out)
		// boxes.stderr.drawVTerm(t.vtErr, t.err)
		boxes.stdout.drawString("--[stdout]--\n"+t.stdout.String(), false)
		boxes.stderr.drawString("--[stderr]--\n"+t.stderr.String(), false)
	}

	// tips
	tips := `--[tips]--
(ctrl+c) quit (enter) run command (ctrl+z) attach (up/down) show/select command`
	if state == attachView {
		tips = `--[tips]--
(ctrl+z) detach`
	}
	boxes.tips.drawString(tips, false)

	screen.Show()
}

func Main() error {

	// Init tcell
	var err error
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, err = tcell.NewScreen()
	if err != nil {
		return err
	}
	defer screen.Fini()

	if err := screen.Init(); err != nil {
		return err
	}

	screen.SetStyle(tcell.StyleDefault)
	screen.Clear()

	// Init workdir
	workdir, err = os.Getwd()
	if err != nil {
		return err
	}

	// channelの準備
	// tcellのイベントを受け取るchan
	eventc := make(chan tcell.Event)
	// tcellのポーリングの終了を通知するchan
	done := make(chan struct{})
	defer close(done)
	// osのsignalを受け取るchan
	sigc := make(chan os.Signal)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 非同期ポーリング
	go func() {
		screen.ChannelEvents(eventc, done)
	}()

	// Main loop
loop:
	for {

		layout(screen.Size())
		draw()

		select {
		case <-sigc:
			break loop

		case e := <-eventc:
			err := handleEvent(e)

			if err == io.EOF {
				break loop
			} else if err != nil {
				panic(err)
			}

		case x := <-exitc:
			t := taskViews[x.taskID]
			t.exit = true
			t.exitCode = x.exitCode
			if lastCD < x.taskID {
				// 先に実行したタスクが遅れてcdしてくるのを防ぐ仕組み
				lastCD = x.taskID
				workdir = x.workdir
			}

		case o := <-stdoutc:
			t := taskViews[o.taskID]
			// n, err := t.vtOut.Write(o.p)
			n, err := t.stdout.Write(o.p)
			o.respc <- writeResp{n, err}

		case e := <-stderrc:
			t := taskViews[e.taskID]
			// n, err := t.vtErr.Write(e.p)
			n, err := t.stderr.Write(e.p)
			e.respc <- writeResp{n, err}
		}
	}

	return nil
}

func handleEvent(e tcell.Event) error {

	switch e := e.(type) {
	case *tcell.EventKey:
		switch state {
		case promptView:
			return handleKeyEvent(e)
		case attachView:
			return handleKeyEventAttach(e)
		}
	}

	return nil
}

func handleKeyEventAttach(e *tcell.EventKey) error {
	k, r := e.Key(), e.Rune()

	switch k {
	case tcell.KeyCtrlZ:
		state = promptView

	case tcell.KeyCtrlD:
		CloseStdin(cursor)
		state = promptView

	case tcell.KeyRune:
		if r <= 0xff {
			WriteStdin(cursor, []byte{byte(r)})
		}

	default:
		WriteStdin(cursor, []byte{byte(k)})
	}

	return nil
}

func handleKeyEvent(e *tcell.EventKey) error {
	k, r := e.Key(), e.Rune()

	switch k {
	case tcell.KeyEnter:
		// Run task
		tid, err := RunTask(workdir, string(prompt))
		if err != nil {
			return err
		}
		cursor = tid
		taskViews = append(taskViews, &taskView{command: string(prompt)})

		// reset
		prompt = []rune{}

	case tcell.KeyCtrlC:
		return io.EOF

	case tcell.KeyCtrlZ:
		state = attachView

	case tcell.KeyDown:
		if cursor > 0 {
			cursor--
		}

	case tcell.KeyUp:
		if cursor < len(taskViews)-1 {
			cursor++
		}

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(prompt) > 0 {
			prompt = prompt[:len(prompt)-1]
		}

	case tcell.KeyRune:
		prompt = append(prompt, r)

	default:
		prompt = append(prompt, rune(k))
	}

	return nil
}
