package main

import (
	"github.com/vivek-yadav/rabbit/cmd"
	"github.com/vivek-yadav/rabbit/zlog"
)

func main() {
	zlog.InitLogger()
	cmd.Execute()
}
