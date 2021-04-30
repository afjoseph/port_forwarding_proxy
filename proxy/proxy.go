package proxy

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/afjoseph/commongo/print"
)

var ERR_LISTEN = errors.New("LISTEN")
var ERR_REMOTE_DIAL = errors.New("REMOTE_DIAL")
var ERR_FORWARD_STREAM = errors.New("FORWARD_STREAM")

func Run(localPort, remoteIp, remotePort string) error {
	print.DebugFunc()
	ln, err := net.Listen("tcp4", ":"+localPort)
	if err != nil {
		return print.ErrorWrapf(ERR_LISTEN, err.Error())
	}
	// This loops, so that we can keep accepting new connections
	for {
		// Accept incoming request to proxy
		conn, err := ln.Accept()
		if err != nil {
			return print.ErrorWrapf(ERR_LISTEN, err.Error())
		}
		go func() {
			err := establishConnectionToRemote(conn, remoteIp, remotePort)
			// XXX Don't die if a connection is lost: just mention the error
			if err != nil {
				print.Warnln(err)
			}
			err = conn.Close()
			if err != nil {
				print.Warnln(err)
			}
		}()
	}
}

// establishConnectionToRemote takes a 'conn', establishes a new connection to
// the remoteIp and remotePort.
// establishConnectionToRemote should also close conn when it is done with it.
func establishConnectionToRemote(
	conn net.Conn,
	remoteIp, remotePort string) error {
	print.DebugFunc()

	proxy, err := net.Dial("tcp4", fmt.Sprintf("%s:%s", remoteIp, remotePort))
	if err != nil {
		return print.ErrorWrapf(ERR_REMOTE_DIAL, err.Error())
	}
	// We wanna establish a bi-directional connection
	// So, both those functions need to run together.
	// When either of them returns an error, we need to shut-down both
	// When either of them is done, we need to close the other
	// Apart from that, we just keep looping them
	errChan := make(chan error, 10)
	go forwardStream(conn, proxy, errChan)
	go forwardStream(proxy, conn, errChan)
loop:
	select {
	case err := <-errChan:
		switch err {
		case nil:
			// Do nothing
		case io.EOF:
			// We reached the end: close the proxy and exit this connection
			print.Debugf("Found an EOF: closing this connection\n")
			proxy.Close()
			return nil
		default:
			return print.ErrorWrapf(ERR_FORWARD_STREAM, err.Error())
		}
	default:
		time.Sleep(2 * time.Second)
		goto loop
	}
	return nil
}

// forwardStream enters an infinite loop to read from "srcConn" and write to
// "dstConn". The idea here is that this will continue working **until** we get
// an EOF from a read or a write: afterwards, we'll exit.
func forwardStream(
	srcConn, dstConn net.Conn,
	errChan chan error) {
	srcReader := bufio.NewReader(srcConn)
	dstWriter := bufio.NewWriter(dstConn)
	buf := make([]byte, 1024)
	for {
		n, err := srcReader.Read(buf)
		if err != nil {
			errChan <- err
			return
		}
		if n > 0 {
			_, err = dstWriter.Write(buf[:n])
			if err != nil {
				errChan <- err
				return
			}
			err = dstWriter.Flush()
			if err != nil {
				errChan <- err
				return
			}
			fmt.Printf("%s -> %s\n",
				srcConn.LocalAddr(), dstConn.LocalAddr())
			fmt.Printf("%s", hex.Dump(buf[:n]))
			fmt.Printf("-------------------\n")
		}
	}
}
