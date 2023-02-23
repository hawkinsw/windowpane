//go:build unix

package window

import (
	"fmt"
	"net"
)

func GetSenderWindowAvailable() bool {
	return false
}

func GetSenderWindow(basicConn net.Conn) (uint32, error) {
	return 0, fmt.Errorf("GetSenderWindow not available on UNIX platforms.")
}
