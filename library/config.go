package library

import (
	"fmt"
	"os"
	"path/filepath"
)

func InitConfig() bool {
	ex, err := os.Executable()
	if err != nil {
		return false
	}
	exPath := filepath.Dir(ex)

	fmt.Println(exPath)

	return true
}
