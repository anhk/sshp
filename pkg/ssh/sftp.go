package ssh

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/sftp"
)

type Sftp struct {
	*Terminal

	client *sftp.Client
}

func NewSftp(target, user, password, privateKey, privateKeyPass string) *Sftp {
	sftp := &Sftp{}
	sftp.Terminal = NewTerminal(target, user, password, privateKey, privateKeyPass)
	return sftp
}

func (s *Sftp) Dial() (err error) {
	if err := s.Terminal.Dial(); err != nil {
		return err
	}
	s.client, err = sftp.NewClient(s.Terminal.client, sftp.MaxPacket(1<<15))
	return err
}

func (s *Sftp) Close() {
	s.Terminal.Close()
	s.client.Close()
}

//func (s *Sftp) OpenFile(path string) (*sftp.File, error) {
//	return s.client.Create(path)
//}

func (s *Sftp) Mkdir(path string) {
	s.client.MkdirAll(path)
}

func (s *Sftp) Stat(path string) (os.FileInfo, error) {
	return s.client.Stat(path)
}

func (s *Sftp) ParseRemoteDirectory(path string) string {
	if strings.HasPrefix(path, "~") {
		sess, err := s.Terminal.client.NewSession()
		if err != nil {
			fmt.Printf("can't open %v: %v", path, err)
			return path
		}
		defer sess.Close()
		if data, _ := sess.Output(`pwd`); len(data) > 0 {
			return strings.TrimSpace(string(data)) + path[1:]
		}
	}
	return path
}

func (s *Sftp) Upload(localPath, remotePath string) {
	srcFile, err := os.Open(localPath)
	if err != nil {
		fmt.Printf("upload %v failed: %v\n", localPath, err)
		return
	}
	defer srcFile.Close()

	remotePath = s.ParseRemoteDirectory(remotePath)
	dstFile, err := s.client.Create(remotePath)
	if err != nil {
		fmt.Printf("upload to %v failed: %v\n", remotePath, err)
		return
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Printf("upload %v failed: %v\n", localPath, err)
	} else {
		fmt.Printf("upload %s ok\n", localPath)
	}

}
