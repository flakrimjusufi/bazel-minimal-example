package guiver

import (
	eventCore "github.com/medialab-ai/guiver/api/core"
	gc "github.com/medialab-ai/guiver/pkg/client"
)

const (
	AppVersion = "1.0.0.1"
	AppName    = "Whisper"
	Event      = "imgur_events_guiver"
	AppID      = "sh.whisper"
)

func SendEvent(eventProperties map[string]interface{}) {

	eventToGuiver := &eventCore.WeaverEvent{
		AppVersion: AppVersion,
		AppName:    AppName,
		Event:      Event,
		ExtraJSON:  eventProperties,
		AppID:      AppID,
	}

	gc.Publish(eventToGuiver)
}
