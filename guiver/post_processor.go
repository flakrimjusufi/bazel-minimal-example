package guiver

import (
	"log"
	"sync"
)

type (
	// ProcessorFunc takes in slice of weaver events and then processes them
	ProcessorFunc = func([]*WeaverEvent)
)

var (
	currProcessorFunc = nilProcessor

	lock = &sync.RWMutex{}
)

// SetPostProcessor is a threadsafe function to set the ProcessorFunc that is called after every ACK
// from sending a batch of events upstream. SetPostProcessor will be called a lot so include questionable
// code at your own discretion
func SetPostProcessor(f ProcessorFunc) {
	if f == nil {
		log.Println("tried to setup nil guiver postprocessor. shame")
		return
	}
	lock.Lock()
	defer lock.Unlock()
	currProcessorFunc = f
}

func postProcess(events []*WeaverEvent) {
	lock.RLock()
	defer lock.RUnlock()

	currProcessorFunc(events)
}

// nilProcessor does nothing
func nilProcessor([]*WeaverEvent) {}
