//go:build !(windows || unix)

package window

func GetSenderWindowAvailable() bool {
	return false
}
func GetSenderWindow(basicConn net.Conn) (uint32, error) {
	return 0, fmt.Errorf("GetSenderWindow unavailable on this platform.")
}
