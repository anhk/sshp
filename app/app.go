package app

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/anhk/sshp/pkg/exp"
	"github.com/anhk/sshp/pkg/ssh"
	"github.com/cihub/seelog"
	"github.com/jedib0t/go-pretty/v6/table"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

//type PasswordDict struct {
//	Id       int64
//	Password string    `xorm:"varchar(256) not null comment('密码')"`
//	LastUsed time.Time `xorm:"timestamp"`
//
//	CreatedAt time.Time `xorm:"created"`
//	UpdatedAt time.Time `xorm:"updated"`
//}

type SshInfo struct {
	Id        int64
	UserName  string    `xorm:"varchar(64) not null comment('用户名')"`
	Password  string    `xorm:"varchar(256) null comment('密码')"`
	Address   string    `xorm:"varchar(256) not null comment('服务器地址')"`
	Port      int       `xorm:"int default 22 comment('端口')"`
	Tag       string    `xorm:"varchar(64) null comment('标签')'"`
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
}

type App struct {
	defaultUser string
	engine      *xorm.Engine
	sftp        *ssh.Sftp
	priKey      string
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
		//exp.Throw(engine.Sync2(PasswordDict{}))
	}
	app.engine = engine
	app.defaultUser = "root"
	app.loadDefaultPrivateKey()
}

func (app *App) getPasswordFromDatabase(cfg *SshConfig) (int64, string) {
	info := &SshInfo{
		UserName: cfg.user,
		Address:  cfg.target,
		Port:     cfg.Port,
	}
	if has, err := app.engine.Get(info); err != nil {
		return 0, ""
	} else if !has {
		return 0, ""
	}
	return info.Id, info.Password
}

func (app *App) getPasswordFromStdin(cfg *SshConfig) string {
	fmt.Printf("%v@%v's password: ", cfg.user, cfg.target)
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password)
}

func (app *App) savePasswordToDatabase(cfg *SshConfig, uid int64, password string) {
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

func (app *App) DoSSH(cfg *SshConfig) {
	cfg.user = If(cfg.user == "", app.defaultUser, cfg.user)
	seelog.Infof("ssh -p %d %s@%s", cfg.Port, cfg.user, cfg.target)

	uid, password := app.getPasswordFromDatabase(cfg)

	for i := 0; i < 4; i++ {
		if password == "" && app.priKey == "" {
			password = app.getPasswordFromStdin(cfg) // read from stdin
		}

		t := ssh.NewTerminal(fmt.Sprintf("%s:%d", cfg.target, cfg.Port), cfg.user,
			password, strings.TrimSpace(app.priKey), "")

		if err := t.Dial(); err != nil {
			password = ""
			app.priKey = ""
			continue
		}

		app.savePasswordToDatabase(cfg, uid, password)
		app.engine.Close()

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

func (app *App) DoSCP(cfg *ScpConfig) {
	seelog.Infof("scp -P %d %s@%s [%s <-> %s] %s",
		cfg.Port, cfg.user, cfg.target, cfg.localFile, cfg.remoteFile,
		If(cfg.upload, "upload", "download"))

	cfg.user = If(cfg.user == "", app.defaultUser, cfg.user)
	uid, password := app.getPasswordFromDatabase(&cfg.SshConfig)

	for i := 0; i < 4; i++ {
		if password == "" {
			password = app.getPasswordFromStdin(&cfg.SshConfig) // read from stdin
		}

		sftp := ssh.NewSftp(fmt.Sprintf("%s:%d", cfg.target, cfg.Port), cfg.user,
			password, strings.TrimSpace(app.priKey), "")
		if err := sftp.Dial(); err != nil {
			password = ""
			continue
		}
		app.savePasswordToDatabase(&cfg.SshConfig, uid, password)
		app.engine.Close()

		app.sftp = sftp
		if cfg.upload {
			app.doUpload(cfg)
		} else {
			app.doDownload(cfg)
		}
		break
	}
}

func (app *App) doUpload(cfg *ScpConfig) {
	st, err := os.Stat(cfg.localFile)
	if err != nil {
		fmt.Printf("file doesn't existed: %v\n", cfg.localFile)
		return
	}

	if st.IsDir() && cfg.Dir {
		app.UploadDirectory(cfg.localFile, cfg.remoteFile)
	} else if !st.IsDir() {
		app.UploadFile(cfg.localFile, cfg.remoteFile)
	} else {
		fmt.Printf("can't upload directory: %v, please add `-r` flag.\n", cfg.localFile)
	}
}

func (app *App) doDownload(cfg *ScpConfig) {

}

func (app *App) UploadDirectory(localFile, remoteFile string) {
	// target 必须是文件夹
	remoteFile = app.sftp.ParseRemoteDirectory(remoteFile)
	st, err := app.sftp.Stat(remoteFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) { // 非 不存在
		fmt.Printf("can't stat %v: %v\n", remoteFile, err)
		return
	} else if err == nil && !st.IsDir() {
		fmt.Printf("%v is not directory\n", remoteFile)
		return
	}

	remoteFile = path.Join(remoteFile, path.Base(localFile))
	app.sftp.Mkdir(remoteFile)

	localFiles, err := os.ReadDir(localFile)
	if err != nil {
		fmt.Printf("open %v failed: %v\n", localFile, err)
		return
	}

	for _, fi := range localFiles {
		localFilePath := path.Join(localFile, fi.Name())
		remoteFilePath := path.Join(remoteFile, fi.Name())

		if fi.IsDir() {
			app.UploadDirectory(localFilePath, remoteFilePath)
		} else {
			app.UploadFile(localFilePath, remoteFilePath)
		}
	}
}

func (app *App) UploadFile(localFile, remoteFile string) {
	st, err := app.sftp.Stat(remoteFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) { // 非 不存在
		fmt.Printf("can't stat %v: %v\n", remoteFile, err)
		return
	}

	if (st != nil && st.IsDir()) || strings.HasSuffix(remoteFile, "/") {
		remoteFile = path.Join(remoteFile, path.Base(localFile))
	}

	app.sftp.Upload(localFile, remoteFile)
}

func (app *App) loadDefaultPrivateKey() {
	u, _ := user.Current()
	rsaFile := fmt.Sprintf("%v/.ssh/id_rsa", u.HomeDir)
	bytes, err := os.ReadFile(rsaFile)
	if err != nil {
		return
	}
	app.priKey = string(bytes)
}

func (app *App) ShowLastLoggedInDevices(n int) {
	var infoList []SshInfo
	err := app.engine.Desc("updated_at").Limit(n).Find(&infoList)
	exp.Throw(err)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Id", "Tag", "Target", "Password"})
	for _, v := range infoList {
		tableRow := table.Row{}
		tableRow = append(tableRow, v.Id, v.Tag, fmt.Sprintf("%s@%s:%d", v.UserName, v.Address, v.Port), v.Password)
		t.AppendRow(tableRow)
	}
	t.Render()
}

func (app *App) DoTag(idStr string, tagName string) {
	id, err := strconv.Atoi(idStr)
	exp.Throw(err)

	_, err = app.engine.Update(SshInfo{Tag: tagName}, SshInfo{Id: int64(id)})
	exp.Throw(err)
}
