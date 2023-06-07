package app

import "github.com/cihub/seelog"

type fakeLogger struct{}

func (f *fakeLogger) Write(p []byte) (n int, err error) {
	if p[len(p)-1] == '\n' {
		p = p[:len(p)-1]
	}
	seelog.Info(string(p))
	return len(p), nil
}
