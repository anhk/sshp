package loginit

import (
	"fmt"
	"os"
	"os/user"

	"github.com/cihub/seelog"
)

var logStr = `
<seelog minlevel="trace">
	<outputs formatid="main">
		<filter levels="trace,debug,info,warn,error,critical">
			<splitter formatid="formatter">
				<rollingfile type="size" filename="%s/sshp.log" maxsize="20487560" namemode="postfix"  maxrolls="2"/>
			</splitter>
		</filter>
	</outputs>
	<formats>
		<format id="main" format="[%%Date %%Time][%%LEVEL][%%RelFile:%%Line] %%Msg%%n"/>
		<format id="formatter" format="[%%Date %%Time][%%LEVEL] %%Msg%%n"/>
	</formats>
</seelog>`

func Init() error {
	u, _ := user.Current()
	logPath := fmt.Sprintf("%v/.log/", u.HomeDir)
	_ = os.MkdirAll(logPath, 0755)

	if logger, err := seelog.LoggerFromConfigAsString(fmt.Sprintf(logStr, logPath)); err != nil {
		_ = seelog.Critical("err parsing log config", err)
		return err
	} else {
		return seelog.ReplaceLogger(logger)
	}
}
