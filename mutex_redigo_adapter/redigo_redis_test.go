package mutex_redigo_adapter

import (
	"mutex/mutex"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	BeforeAllTests()
	retCode := m.Run()
	AfterAllTests()

	os.Exit(retCode)
}

var r Redis

const mutexKey = "resource"

func BeforeAllTests() {
	r = New("127.0.0.1:6379", "0", 3)
}

func AfterAllTests() {
}

func TestLock(t *testing.T) {
	r.execute("FLUSHDB")
	mutex := mutex.New(r, mutexKey, 0)

	locked, err := mutex.Lock()

	if !locked {
		t.Errorf("Expected to be locked, but it is not")
	}

	if err != nil {
		t.Errorf("Error on locking: %+v", err.Error())
	}
}

func TestUnlockWhenLocked(t *testing.T) {
	r.execute("FLUSHDB")
	mutex := mutex.New(r, mutexKey, 0)

	if locked, err := mutex.Lock(); !locked || err != nil {
		t.Errorf("Expected to be locked, but it is not")
	}

	if err := mutex.Unlock(); err != nil {
		t.Errorf("Got error while unlocking: %+v", err.Error())
	}
}

func TestUnlockWhenNotLocked(t *testing.T) {
	r.execute("FLUSHDB")
	mutex := mutex.New(r, mutexKey, 0)

	if err := mutex.Unlock(); err != nil {
		t.Errorf("Got error while unlocking: %+v", err.Error())
	}
}

func TestWaitsForPreviousOwnerToUnlock(t *testing.T) {
	r.execute("FLUSHDB")
	mtx := mutex.New(r, mutexKey, 0)

	resource := 1

	go func(mtx mutex.Mutex) {
		mtx.Lock()
		time.Sleep(200 * time.Millisecond)
		resource = 2
		mtx.Unlock()
	}(mtx)

	time.Sleep(100 * time.Millisecond)
	if resource != 1 {
		t.Errorf("Resource had to be 1, but was %+v", resource)
	}

	mtx.Lock()
	if resource != 2 {
		t.Errorf("Resource had to be 2, but was %+v", resource)
	}
}
