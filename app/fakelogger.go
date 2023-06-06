package app

import "github.com/cihub/seelog"

type fakeLogger struct {
}

func (f *fakeLogger) Write(p []byte) (n int, err error) {
	seelog.Info(string(p))
	return len(p), nil
}
