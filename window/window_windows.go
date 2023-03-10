//go:build windows

// Please NOTE: This is blatant self plagairism of code within
// goresponsiveness.
package window

import (
	"fmt"
	"net"
	"unsafe"

	"golang.org/x/sys/windows"
)

type TCPINFO_BASE struct {
	State             uint32
	Mss               uint32
	ConnectionTimeMs  uint64
	TimestampsEnabled bool
	RttUs             uint32
	MinRttUs          uint32
	BytesInFlight     uint32
	Cwnd              uint32
	SndWnd            uint32
	RcvWnd            uint32
	RcvBuf            uint32
	BytesOut          uint64
	BytesIn           uint64
	BytesReordered    uint32
	BytesRetrans      uint32
	FastRetrans       uint32
	DupAcksIn         uint32
	TimeoutEpisodes   uint32
	SynRetrans        byte // UCHAR
}

// https://github.com/tpn/winsdk-10/blob/9b69fd26ac0c7d0b83d378dba01080e93349c2ed/Include/10.0.16299.0/shared/mstcpip.h#L289
type TCPINFO_V0 struct {
	TCPINFO_BASE
}

// https://docs.microsoft.com/en-us/windows/win32/api/mstcpip/ns-mstcpip-tcp_info_v1
type TCPINFO_V1 struct {
	TCPINFO_BASE
	SndLimTransRwin uint32
	SndLimTimeRwin  uint32
	SndLimBytesRwin uint64
	SndLimTransCwnd uint32
	SndLimTimeCwnd  uint32
	SndLimBytesCwnd uint64
	SndLimTransSnd  uint32
	SndLimTimeSnd   uint32
	SndLimBytesSnd  uint64
}

func getTCPInfoRaw(basicConn net.Conn) (*TCPINFO_V1, error) {
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

	// SIO_TCP_INFO
	// https://docs.microsoft.com/en-us/windows/win32/winsock/sio-tcp-info
	// https://github.com/tpn/winsdk-10/blob/master/Include/10.0.16299.0/shared/mstcpip.h
	iocc := uint32(windows.IOC_INOUT | windows.IOC_VENDOR | 39)

	// Should be a DWORD, 0 for version 0, 1 for version 1 tcp_info:
	// 0: https://docs.microsoft.com/en-us/windows/win32/api/mstcpip/ns-mstcpip-tcp_info_v0
	// 1: https://docs.microsoft.com/en-us/windows/win32/api/mstcpip/ns-mstcpip-tcp_info_v1
	inbuf := uint32(1)

	// Size of the inbuf variable
	cbif := uint32(4)

	outbuf := TCPINFO_V1{}

	cbob := uint32(unsafe.Sizeof(outbuf)) // Size = 136 for V1 and 88 for V0

	// Size pointer of return object
	cbbr := uint32(0)

	completionRoutine := uintptr(0)

	ov := windows.Overlapped{}

	rawConn.Control(func(fd uintptr) {
		err = windows.WSAIoctl(
			windows.Handle(fd),
			iocc,
			(*byte)(unsafe.Pointer(&inbuf)),
			cbif,
			(*byte)(unsafe.Pointer(&outbuf)),
			cbob,
			&cbbr,
			&ov,
			completionRoutine,
		)
	})
	return &outbuf, err
}

func GetSenderWindowAvailable() bool {
	return true
}
func GetSenderWindow(connection net.Conn) (uint32, error) {
	info, err := getTCPInfoRaw(connection)
	if err != nil {
		return 0, err
	}
	return info.SndWnd, nil
}
