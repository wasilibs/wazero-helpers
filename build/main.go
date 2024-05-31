package main

import (
	"flag"

	"github.com/curioswitch/go-build"
	"github.com/goyek/x/boot"
)

func main() {
	_ = flag.Lookup("v").Value.Set("true") // Force verbose output
	build.DefineTasks()
	boot.Main()
}
