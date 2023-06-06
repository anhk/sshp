package app

import "strings"

// Config
//   - ssh -p 22 anhk@10.226.133.9
type Config struct {
	Port   int
	target string
	user   string
}

// Parse 解析SSH命令
//   - addr => anhk@10.226.133.9
func (cfg *Config) Parse(addr string) *Config {
	arr := strings.Split(addr, "@")
	if len(arr) == 1 {
		cfg.target = arr[0]
	} else if len(arr) >= 2 {
		cfg.user = arr[0]
		cfg.target = arr[1]
	}

	return cfg
}
