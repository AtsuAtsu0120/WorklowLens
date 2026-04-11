package lock

import (
	"errors"
	"fmt"
	"net"
)

// DefaultLockPort はロック用のデフォルトポート。
const DefaultLockPort = 59099

// ErrAlreadyRunning は既に別インスタンスが起動中であることを示すエラー。
var ErrAlreadyRunning = errors.New(
	"workflow_lensは既に起動中です。複数のインスタンスを同時に実行することはできません。",
)

// InstanceLock はインスタンスロック。内部でTCPリスナーを保持し、
// Close()されるまでポートを占有する。
type InstanceLock struct {
	listener net.Listener
}

// Acquire はデフォルトポート（59099）でロックを取得する。
func Acquire() (*InstanceLock, error) {
	return AcquireOnPort(DefaultLockPort)
}

// AcquireOnPort は指定ポートでロックを取���する（テスト用）。
func AcquireOnPort(port int) (*InstanceLock, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, ErrAlreadyRunning
	}
	return &InstanceLock{listener: listener}, nil
}

// Port はロックに使用しているポート番号を返す。
func (l *InstanceLock) Port() int {
	return l.listener.Addr().(*net.TCPAddr).Port
}

// Close はロックを解放する。
func (l *InstanceLock) Close() error {
	return l.listener.Close()
}
