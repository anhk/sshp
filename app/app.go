package app

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"xorm.io/xorm/log"

	"github.com/cihub/seelog"

	"golang.org/x/crypto/ssh/terminal"

	"golang.org/x/term"

	"github.com/anhk/sshp/pkg/ssh"

	"github.com/anhk/sshp/pkg/exp"

	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

type PasswordDict struct {
	Id       int64
	Password string    `xorm:"varchar(256) not null comment('密码')"`
	LastUsed time.Time `xorm:"timestamp"`

	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
}

type SshInfo struct {
	Id        int64
	UserName  string    `xorm:"varchar(64) not null comment('用户名')"`
	Password  string    `xorm:"varchar(256) null comment('密码')"`
	Address   string    `xorm:"varchar(256) not null comment('服务器地址')"`
	Port      int       `xorm:"int default 22 comment('端口')"`
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
}

type App struct {
	defaultUser string
	engine      *xorm.Engine
}

func (app *App) Init() {
	u, err := user.Current()
	exp.Throw(err)

	needSync := false

	dbFile := fmt.Sprintf("%v/.ssh/sshp.db", u.HomeDir)
	if _, err := os.Stat(dbFile); err != nil && errors.Is(err, os.ErrNotExist) {
		_, _ = os.Create(dbFile)
		needSync = true
	} else {
		exp.Throw(err)
	}

	engine, err := xorm.NewEngine("sqlite3", fmt.Sprintf("file:%s?cache=shared", dbFile))
	exp.Throw(err)

	logger := log.NewSimpleLogger(&fakeLogger{})
	logger.ShowSQL(true)
	engine.SetLogger(logger)
	if needSync {
		exp.Throw(engine.Sync2(SshInfo{}))
		exp.Throw(engine.Sync2(PasswordDict{}))
	}
	app.engine = engine
	app.defaultUser = "root"
}

func (app *App) getPasswordFromDatabase(cfg *Config) (int64, string) {
	info := &SshInfo{
		UserName: cfg.user,
		Address:  cfg.target,
		Port:     cfg.Port,
	}
	if has, err := app.engine.Get(info); err != nil {
		seelog.Errorf("-- %v", err)
		return 0, ""
	} else if !has {
		seelog.Errorf("!has")
		return 0, ""
	}
	return info.Id, info.Password
}

func (app *App) getPasswordFromStdin(cfg *Config) string {
	fmt.Printf("%v@%v's password: ", cfg.user, cfg.target)
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	return string(password)
}

func (app *App) savePasswordToDatabase(cfg *Config, uid int64, password string) {
	info := &SshInfo{
		UserName: cfg.user,
		Address:  cfg.target,
		Password: password,
		Port:     cfg.Port,
	}

	seelog.Infof("save password to database: %v", uid)
	if uid == 0 {
		_, err := app.engine.InsertOne(info)
		if err != nil {
			seelog.Errorf("InsertOne: %v", err)
		}
	} else {
		app.engine.Update(info, &SshInfo{Id: uid})
	}
}

func (app *App) DoSSH(cfg *Config) {
	cfg.user = If(cfg.user == "", app.defaultUser, cfg.user)

	seelog.Infof("ssh -p %d %s@%s", cfg.Port, cfg.user, cfg.target)

	uid, password := app.getPasswordFromDatabase(cfg)

	for i := 0; i < 4; i++ {
		if password == "" {
			password = app.getPasswordFromStdin(cfg) // read from stdin
		}

		t := ssh.NewTerminal(fmt.Sprintf("%s:%d", cfg.target, cfg.Port), cfg.user,
			password, "", "")
		if err := t.Dial(); err != nil {
			password = ""
			continue
		}

		app.savePasswordToDatabase(cfg, uid, password)

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGHUP, syscall.SIGABRT, syscall.SIGTERM)
		go func() {
			for {
				s := <-sigs
				if s == syscall.SIGWINCH {
					if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
						t.WindowChange(h, w)
						continue
					}
				}
			}
		}()

		if err := t.Start(); err != nil {
			break
		}

		_ = t.Wait()
		break
	}
}
