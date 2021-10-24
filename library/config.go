package library

import (
	"os"
	"path/filepath"
)

type generalConfig struct {
	Hostname string
	Port     int8
	TLS      gctls
}

type gctls struct {
	PemFile string
	KeyFile string
}

func InitConfig() bool {

	ex, err := os.Executable()
	if err != nil {
		return false
	}

	exPath := filepath.Dir(ex)

	return true
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func isFile(path string) bool {
	return !isDir(path)
}
