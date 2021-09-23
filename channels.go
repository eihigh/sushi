package sushi

var (
	exitc   = make(chan exitMsg)
	stdoutc = make(chan writeReq)
	stderrc = make(chan writeReq)
)

type exitMsg struct {
	taskID   int
	exitCode int
	workdir  string
}

type writeReq struct {
	taskID int
	p      []byte

	respc chan writeResp
}

type writeResp struct {
	n   int
	err error
}

type stdoutHandler struct{ id int }
type stderrHandler struct{ id int }

func (h *stdoutHandler) Write(p []byte) (int, error) {
	respc := make(chan writeResp)
	req := writeReq{taskID: h.id, p: p, respc: respc}
	stdoutc <- req
	resp := <-respc
	return resp.n, resp.err
}

func (h *stderrHandler) Write(p []byte) (int, error) {
	respc := make(chan writeResp)
	req := writeReq{taskID: h.id, p: p, respc: respc}
	stderrc <- req
	resp := <-respc
	return resp.n, resp.err
}
