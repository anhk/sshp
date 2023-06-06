package app

import (
	"fmt"
	"os"
	"strings"
)

// SshConfig
//   - ssh -p 22 anhk@10.226.133.9
type SshConfig struct {
	Port   int
	target string
	user   string
}

// Parse 解析SSH命令
//   - addr => anhk@10.226.133.9
//   - addr => anhk@10.226.133.9:/tmp/test
func (cfg *SshConfig) Parse(addr string) *SshConfig {
	arr := strings.FieldsFunc(addr, func(r rune) bool {
		return If(r == '@' || r == ':', true, false)
	})
	if len(arr) == 1 {
		cfg.target = arr[0]
	} else if len(arr) >= 2 {
		cfg.user = arr[0]
		cfg.target = arr[1]
	}

	return cfg
}

// ScpConfig
//   - scp -P 22 -r ./file.tar anhk@10.226.133.9:/file.tar
type ScpConfig struct {
	SshConfig
	localFile  string
	remoteFile string
	upload     bool // download or upload
	Dir        bool // 上传/下载文件夹
}

func getFileFromTarget(target string) string {
	arr := strings.Split(target, ":")
	if len(arr) != 2 {
		fmt.Println("invalid target: ", target)
		os.Exit(-1)
	}
	return arr[1]
}

// Parse 解析SCP命令
//   - addr => anhk@10.226.133.9
func (cfg *ScpConfig) Parse(src, dst string) *ScpConfig {
	download := strings.Contains(src, ":")
	cfg.upload = strings.Contains(dst, ":")

	if (cfg.upload && download) || (!cfg.upload && !download) {
		fmt.Printf("invalid src or dst\n")
		os.Exit(-1)
	}

	if cfg.upload {
		cfg.localFile = src
		cfg.SshConfig.Parse(dst)
		cfg.remoteFile = getFileFromTarget(dst)
	} else {
		cfg.SshConfig.Parse(src)
		cfg.localFile = dst
		cfg.remoteFile = getFileFromTarget(src)
	}
	return cfg
}
