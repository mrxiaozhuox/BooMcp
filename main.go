package main

import (
	"fmt"
	"os"

	"fkyos.com/mcp/library"
	"fkyos.com/mcp/service"
)

func main() {

	library.InitConfig()

	fmt.Println(os.Getenv("APPDATA"))

	service.InitServer()

}
