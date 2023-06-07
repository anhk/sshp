package ssh

import (
	"errors"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Terminal ssh 终端
type Terminal struct {
	Target         string
	User           string
	Password       string
	PrivateKey     string
	PrivateKeyPass string

	client   *ssh.Client
	session  *ssh.Session
	oldState *term.State
	w, h     int // 宽与高

	exit bool
}

func NewTerminal(target, user, password, privateKey, privateKeyPass string) *Terminal {
	return &Terminal{
		Target:         target,
		User:           user,
		Password:       password,
		PrivateKey:     privateKey,
		PrivateKeyPass: privateKeyPass,
	}
}

func (t *Terminal) Dial() (err error) {
	cfg := &ssh.ClientConfig{
		Config:          ssh.Config{},
		User:            t.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		BannerCallback:  nil,
		Timeout:         5 * time.Second,
	}

	if t.Password != "" {
		cfg.Auth = append(cfg.Auth, ssh.Password(t.Password))
	}
	if t.PrivateKey != "" {
		if t.PrivateKeyPass != "" { // 私钥带密码
			if s, err := ssh.ParsePrivateKeyWithPassphrase([]byte(t.PrivateKey), []byte(t.PrivateKeyPass)); err == nil {
				cfg.Auth = append(cfg.Auth, ssh.PublicKeys(s))
			}
		} else {
			if s, err := ssh.ParsePrivateKey([]byte(t.PrivateKey)); err == nil {
				cfg.Auth = append(cfg.Auth, ssh.PublicKeys(s))
			}
		}
	}

	if len(cfg.Auth) == 0 {
		return errors.New("empty auth information")
	}

	if t.client, err = ssh.Dial("tcp", t.Target, cfg); err != nil {
		return err
	}
	return nil
}

func (t *Terminal) Start() error {
	if err := t.newSession(); err != nil {
		return err
	}
	if err := t.startPty(); err != nil {
		return err
	}
	return t.startShell()
}

func (t *Terminal) WindowChange(h, w int) {
	if t.session != nil {
		t.session.WindowChange(h, w)
	}
}

func (t *Terminal) Wait() error {
	defer func() {
		term.Restore(int(os.Stdin.Fd()), t.oldState)
	}()
	if err := t.session.Wait(); err != nil {
		if err, ok := err.(*ssh.ExitError); ok {
			if err.ExitStatus() == 130 || err.ExitStatus() == 127 {
				return nil
			}
		}
		return err
	}
	return nil
}

func (t *Terminal) Close() {
	if t.session != nil {
		t.session.Close()
	}
	t.client.Close()
}

func (t *Terminal) newSession() (err error) {

	if t.session, err = t.client.NewSession(); err != nil {
		return err
	}

	fd := int(os.Stdin.Fd())
	if t.oldState, err = term.MakeRaw(fd); err != nil {
		return err
	}

	t.w, t.h, err = term.GetSize(fd)
	return err
}

func (t *Terminal) startPty() error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	return t.session.RequestPty("xterm", t.h, t.w, modes)
}

func (t *Terminal) startShell() error {
	t.session.Stdout = os.Stdout
	t.session.Stdin = os.Stdin
	t.session.Stderr = os.Stderr
	if err := t.session.Shell(); err != nil {
		return err
	}
	return nil
}
