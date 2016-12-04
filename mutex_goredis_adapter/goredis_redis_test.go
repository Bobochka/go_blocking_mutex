package mutex_goredis_adapter

import (
	"log"
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
	r.client.FlushDb()
	mutex := mutex.New(r, mutexKey)

	locked, err := mutex.Lock()

	if !locked {
		t.Errorf("Expected to be locked, but it is not")
	}

	if err != nil {
		t.Errorf("Error on locking: %+v", err.Error())
	}
}

func TestUnlockWhenLocked(t *testing.T) {
	r.client.FlushDb()
	mutex := mutex.New(r, mutexKey)

	if locked, err := mutex.Lock(); !locked || err != nil {
		t.Errorf("Expected to be locked, but it is not")
	}

	if err := mutex.Unlock(); err != nil {
		t.Errorf("Got error while unlocking: %+v", err.Error())
	}
}

func TestUnlockWhenNotLocked(t *testing.T) {
	r.client.FlushDb()
	mutex := mutex.New(r, mutexKey)

	if err := mutex.Unlock(); err != nil {
		t.Errorf("Got error while unlocking: %+v", err.Error())
	}
}

func TestWaitsForPreviousOwnerToUnlock(t *testing.T) {
	r.client.FlushDb()
	mtx := mutex.New(r, mutexKey)

	resource := 1

	go func(mtx mutex.Mutex) {
		mtx.Lock()
		time.Sleep(1200 * time.Millisecond)
		resource = 2
		mtx.Unlock()
	}(*mtx)

	time.Sleep(100 * time.Millisecond)
	if resource != 1 {
		t.Errorf("Resource had to be 1, but was %+v", resource)
	}

	mtx.Lock()
	if resource != 2 {
		log.Printf("resource2: %+v\n", resource)
		t.Errorf("Resource had to be 2, but was %+v", resource)
	}
}
