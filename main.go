package main

import (
	"flag"
	"os"

	"github.com/afjoseph/commongo/print"
	"github.com/afjoseph/port_forward_proxy/proxy"
)

var (
	localPortFlag  = flag.String("local_port", "4444", "sss")
	remoteIpFlag   = flag.String("remote_ip", "127.0.0.1", "sss")
	remotePortFlag = flag.String("remote_port", "12345", "sss")
	verboseFlag    = flag.Bool("verbose", false, "verbosity")
)

func main() {
	flag.Parse()
	if *verboseFlag {
		print.SetLevel(print.LOG_DEBUG)
	}
	err := proxy.Run(*localPortFlag, *remoteIpFlag, *remotePortFlag)
	if err != nil {
		print.Warnln(err)
		os.Exit(1)
	}
}
