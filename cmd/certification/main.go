package main

import (
	"flag"

	"k8s.io/klog"
)

var (
	showVersion, help bool
)

func main() {
	flag.BoolVar(&showVersion, "version", false, "Show release version and exit")
	flag.BoolVar(&help, "help", false, "Show flag defaults and exit")
	klog.InitFlags(nil)
	flag.Parse()

	// TODO
}
