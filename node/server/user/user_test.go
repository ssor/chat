package user

import (
	"fmt"
	"sync"
	"testing"
)

var lock = &sync.Mutex{}

func TestLock(t *testing.T) {
	lockf(1)
}

func lockf(flag int) {
	fmt.Println("lock -> ", flag)
	lock.Lock()
	defer func() {
		lock.Unlock()
		fmt.Println("unlock -> ", flag)
	}()
	lockf(flag + 1)
}
