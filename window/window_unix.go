//go:build unix && !darwin

package window

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

func GetSenderWindowAvailable() bool {
	return true
}

func GetSenderWindow(basicConn net.Conn) (uint32, error) {
	var err error = nil
	var info *unix.TCPInfo = nil
	if info, err = getTCPInfo(basicConn); err == nil {
		return info.Snd_wnd, nil
	}
	return 0, err
}

/*
 * Blatant self plagiarism from goresponsiveness.
 */
func getTCPInfo(basicConn net.Conn) (*unix.TCPInfo, error) {
	tcpConn, ok := basicConn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf(
			"OOPS: Could not get the TCP info for the connection (not a TCP connection)",
		)
	}
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return nil, err
	}
	var info *unix.TCPInfo = nil
	rawConn.Control(func(fd uintptr) {
		info, err = unix.GetsockoptTCPInfo(int(fd), unix.SOL_TCP, unix.TCP_INFO)
	})
	return info, err
}
