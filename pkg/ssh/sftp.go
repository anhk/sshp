package ssh

import "github.com/pkg/sftp"

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

func (s *Sftp) OpenFile(path string) (*sftp.File, error) {
	return s.client.Create(path)
}
