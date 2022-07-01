package guiver

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	maxFillWait     = time.Millisecond * 500
	ErrBackPressure = errors.New("guiver queue backpressure")
	eventsBackLog   chan *WeaverEvent
	// backlogSize is number of events to keep in memory at a time
	backlogSize int = 20000
	workers         = 10
	// helps gurantee that 12 events get published in
	// a batch of 12 and not 12 batches of 1
	readLock = &sync.Mutex{}
)

const (
	NANO_TO_MS     = 1000000
	miniBundleSize = 200
)

type (
	// WeaverEvent is the basic struct for all events
	WeaverEvent struct {
		AppVersion string      `json:"app_version" bigquery:"app_version" parquet:"name=app_version, type=BYTE_ARRAY, convertedtype=UTF8"`
		AppName    string      `json:"app_name" bigquery:"app_name" parquet:"name=app_name, type=BYTE_ARRAY, convertedtype=UTF8"`
		Event      string      `json:"event" parquet:"name=event, type=BYTE_ARRAY, convertedtype=UTF8"`
		ExtraJSON  interface{} `json:"extra_json" bigquery:"-"` // not in parquet files
		AppID      string      `json:"app_id" bigquery:"app_id" parquet:"name=app_id, type=BYTE_ARRAY, convertedtype=UTF8"`
		Ts         int64       `json:"ts" bigquery:"ts" parquet:"name=ts, type=INT64"`
		EventType  string      `json:"event_type" bigquery:"-"`
	}
)

func Publish(event *WeaverEvent) error {
	event.Ts = NowInUTCMS()

	// If app name is not empty and app id is, copy app name to app id.
	if event.AppName != "" && event.AppID == "" {
		event.AppID = event.AppName
	}

	// dont overwrite fields if someone else set them
	if event.ExtraJSON == nil {
		event.ExtraJSON = make(map[string]interface{})
	}
	if event.EventType == "" {
		event.EventType = event.Event
	}

	return addToBatch(event)
}

// ToMS converts to UTC then converts to ms
func ToMS(t time.Time) int64 {
	return t.UTC().UnixNano() / NANO_TO_MS
}

// NowInUTCMS returns current UTC time in ms
func NowInUTCMS() int64 {
	// ms for life
	return ToMS(Now())
}

// Now current UTC time
func Now() time.Time {
	return time.Now().UTC()
}

// addToBatch tries to add the event to the list of events to process.
// waits until maxFillWait trying to get a spot or returns an error
func addToBatch(event *WeaverEvent) error {
	timeout := time.After(maxFillWait)
	select {
	case <-timeout:
		return ErrBackPressure
	case eventsBackLog <- event:
		return nil
	}
}

func init() {
	eventsBackLog = make(chan *WeaverEvent, backlogSize)

	for i := 0; i < workers; i++ {
		go batcherWorker()
	}
}

// batcherWorker
//	1. gain reads lock
//	2. receive up to miniBundleSize messages OR until maxFillWait passes
//	3. release lock
//	4. publish events
func batcherWorker() {
	// collections to re-use
	myBatch := make(chan *WeaverEvent, miniBundleSize)
	events := make([]*WeaverEvent, miniBundleSize)

	for {
		fillBatch(myBatch)
		// no events so try again
		if len(myBatch) == 0 {
			continue
		}

		events = events[:0]
		for {
			if len(myBatch) == 0 {
				break
			}

			events = append(events, <-myBatch)
		}

		retries := 3
		for {
			if err := publishToPubSub(events); err != nil {
				retries--
				time.Sleep(time.Duration(retries) * time.Second)
			} else {
				break
			}

			if retries == 0 {
				log.Printf("dropping %d guiver events, out of retries", len(events))
				break
			}
		}
	}
}

func fillBatch(myBatch chan<- *WeaverEvent) {
	readLock.Lock()
	defer readLock.Unlock()

	timeout := time.After(maxFillWait)
	for {
		select {
		case event := <-eventsBackLog:
			myBatch <- event
			if len(myBatch) == miniBundleSize {
				return
			}
		case <-timeout:
			return
		}
	}
}
