package lock

import (
	"errors"
	"testing"
)

func TestAcquire_Success(t *testing.T) {
	// ポート0でOS自動割り当てを使用
	lock, err := AcquireOnPort(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer lock.Close()

	if lock.Port() == 0 {
		t.Error("Port() should return a non-zero port")
	}
}

func TestAcquire_DoubleLockFails(t *testing.T) {
	lock1, err := AcquireOnPort(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer lock1.Close()

	// 同じポートで2回目のロック取得は失敗する
	_, err = AcquireOnPort(lock1.Port())
	if err == nil {
		t.Fatal("expected error for double lock")
	}
	if !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("error = %v, want ErrAlreadyRunning", err)
	}
}

func TestAcquire_ReacquireAfterClose(t *testing.T) {
	lock1, err := AcquireOnPort(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	port := lock1.Port()

	// ロック解放
	if err := lock1.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	// 解放後は同じポートで再取得できる
	lock2, err := AcquireOnPort(port)
	if err != nil {
		t.Fatalf("unexpected error after close: %v", err)
	}
	defer lock2.Close()
}
