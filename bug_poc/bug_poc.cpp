// clang-format off
#include <Winsock2.h>
#include <mstcpip.h>
#include <ws2tcpip.h>
#include <Windows.h>
#include <stdio.h>
#include <synchapi.h>
#include <tchar.h>
#include <ws2ipdef.h>
// clang-format on

/** Print a stringified version of a system error message.
 * @param last_error The system error id to stringify and print.
 * @note Adapted from
 * https://learn.microsoft.com/en-us/windows/win32/debug/retrieving-the-last-error-code
 */
void PrintError(DWORD last_error) {

  LPTSTR error_string{nullptr};
  FormatMessage(FORMAT_MESSAGE_ALLOCATE_BUFFER | FORMAT_MESSAGE_FROM_SYSTEM |
                    FORMAT_MESSAGE_IGNORE_INSERTS,
                NULL, last_error, 0, (LPTSTR)&error_string, 0, NULL);

  if (!error_string) {
    _tprintf(TEXT("Unknown error occurred (code: %d)!\n"), last_error);
    return;
  }

  _tprintf(TEXT("%s\n"), error_string);
  LocalFree(error_string);
}

/** Get a completion status from the given handle.
 * @param lpvParam A (disguised) pointer to the completion port handle to
 * monitor.
 * @note Timeout after 10 seconds.
 */
DWORD WINAPI CompletionThread(LPVOID lpvParam) {
  LPHANDLE completion_port_handlep{static_cast<LPHANDLE>(lpvParam)};
  DWORD bytes_xferred{};
  ULONG64 completion_key{};
  LPOVERLAPPED overlappedp{};
  if (!GetQueuedCompletionStatus(*completion_port_handlep, &bytes_xferred,
                                 &completion_key, &overlappedp, 10000)) {
    _tprintf(TEXT("GetQueuedCompletionStatus Failed: \n"));
    PrintError(WSAGetLastError());
    return 0;
  }

  printf("Proof! The completion port was signaled.\n");
  return 0;
}

int main() {

  WSADATA wsa_data{};
  WORD wsa_version = MAKEWORD(2, 2);
  struct sockaddr_in sin {};
  const char *server_ip{"100.96.78.66"};

  // Connect to a windowpane server.
  sin.sin_family = AF_INET;
  sin.sin_port = htons(5001);
  if (inet_pton(sin.sin_family, server_ip, &sin.sin_addr) != 1) {
    _tprintf(TEXT("Could not convert the server address to network format.\n"));
  }

  if (WSAStartup(wsa_version, &wsa_data)) {
    _tprintf(TEXT("WSAStartup failed!\n"));
  }

  auto ws_socket{WSASocketA(AF_INET, SOCK_STREAM, IPPROTO_TCP, NULL,
                            SG_UNCONSTRAINED_GROUP, WSA_FLAG_OVERLAPPED)};

  if (ws_socket == INVALID_SOCKET) {
    _tprintf(TEXT("WSASocketA failed:\n"));
    PrintError(WSAGetLastError());
    WSACleanup();
    return -1;
  }

  auto connect_result{WSAConnect(ws_socket, (const sockaddr *)&sin,
                                 sizeof(struct sockaddr_in), NULL, NULL, NULL,
                                 NULL)};

  if (connect_result == SOCKET_ERROR) {
    _tprintf(TEXT("WSAConnect failed:\n"));
    PrintError(WSAGetLastError());
    WSACleanup();
    return -1;
  }

  OVERLAPPED ov{};

  const int BUFFER_LENGTH{4096};
  char buffer_storage[BUFFER_LENGTH]{};
  WSABUF buf{.len = BUFFER_LENGTH, .buf = buffer_storage};
  DWORD bytes_sent{};

  auto send_result{
      WSASend(ws_socket, &buf, 1, &bytes_sent, WS_OVERLAPPED, &ov, NULL)};
  if (send_result == SOCKET_ERROR) {
    _tprintf(TEXT("Failed to WSASend:\n"));
    PrintError(WSAGetLastError());
  }

  auto completion_port{CreateIoCompletionPort((HANDLE)ws_socket, NULL, 0, 0)};

  if (completion_port == INVALID_HANDLE_VALUE) {
    _tprintf(TEXT("CreateIoCompletionPort failed:\n"));
    PrintError(GetLastError());
    closesocket(ws_socket);
    WSACleanup();
    return -1;
  }

  DWORD dwThreadId{};
  auto completion_thread{CreateThread(
      NULL, 0, CompletionThread, (LPVOID)&completion_port, 0, &dwThreadId)};

  if (completion_thread == NULL) {
    _tprintf(TEXT("CreateThread failed, GLE=%d.\n"), GetLastError());
    CloseHandle(completion_port);
    closesocket(ws_socket);
    WSACleanup();
    return -1;
  }

#if 1
  TCP_INFO_v1 tcp_info{};
  DWORD tcp_info_version{1};

  for (;;) {
    DWORD bytesReceivedFromIoctl{};

    /*
     * The only IO on this socket (i.e., the send of 4k bytes [above]) happened
     * before the completion port was created (and the socket was associated with
     * it) -- the server sends back no data. So, in the absence of the WSAIoctl
     * here, there would be no completion events. Let's do a WSAIoctl and see whether
     * there is a completion event generated!
     * 
     * Note: You can set the 1 to a 0 above to test the above assertion -- without
     * the WSAIoctl here, the GetQueuedCompletionStatus in the completion thread
     * will timeout without having received any events.
     */
    auto ioctl_result{WSAIoctl(
        ws_socket, SIO_TCP_INFO, &tcp_info_version, sizeof(tcp_info_version),
        &tcp_info, sizeof(tcp_info), &bytesReceivedFromIoctl, &ov, NULL)};
    if (ioctl_result == SOCKET_ERROR) {
      _tprintf(TEXT("Failed to WSAIoctl:\n"));
      PrintError(WSAGetLastError());
      continue;
    } else if (bytesReceivedFromIoctl != sizeof(tcp_info)) {
      // It is expected that WSAIoctl will receive back 4096 the first time --
      // the result of the first send on the socket. We will do the WSAIoctl
      // again until the results that we receive are the size of the tcp_info
      // struct.
      _tprintf(TEXT("WSAIoctl received %d bytes but expected %llu\n"),
               bytesReceivedFromIoctl, sizeof(tcp_info));
      continue;
    }
    break;
  }
  _tprintf(TEXT("Send Window Size: %d\n"), tcp_info.SndWnd);
#endif

  for (;;) {
    if (WAIT_TIMEOUT != WaitForSingleObject(completion_thread, 100)) {
      _tprintf(TEXT("Finished the completion thread!\n"));
      break;
    }
    _tprintf(
        TEXT("Timeout waiting for the completion thread to finish ... going "
             "around again.\n"));
  }

  CloseHandle(completion_thread);
  CloseHandle(completion_port);
  closesocket(ws_socket);
  WSACleanup();
  return 0;
}
