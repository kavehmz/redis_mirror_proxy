package main

import (
	"sync"
)

var onlyOneSwitch = make(chan bool, 1)

func switchConnections() {
	onlyOneSwitch <- true
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(n int) {
			pauseMain[n] <- true
			syncMirror[n] <- true
			<-doneMirror[n]
			wg.Done()
		}(i)
	}

	// Wait for all the mirror goroutines to confirm the did their sync
	wg.Wait()

	// switch the main and mirror connection selector
	// switch is based on odd/even so adding one to the number will do the switch
	mode.Add(1)

	// Now continue the normal operation
	for i := 0; i < N; i++ {
		contMain[i] <- true
		contMirror[i] <- true
	}
	<-onlyOneSwitch
}
