package guiver_test

import (
	eventCore "github.com/medialab-ai/guiver/api/core"
	gc "github.com/medialab-ai/guiver/pkg/client"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	AppVersion = "1.0.0.1"
	AppName    = "Whisper"
	Event      = "imgur_events_guiver"
	AppID      = "sh.whisper"
)

func TestSendEvent(t *testing.T) {

	eventToGuiver := &eventCore.WeaverEvent{
		AppVersion: AppVersion,
		AppName:    AppName,
		Event:      Event,
		ExtraJSON: map[string]interface{}{
			"showsAds":        true,
			"device":          "unknown",
			"showAdLevel":     2,
			"inGallery":       true,
			"unsafeFlags":     []string{"sixth_mod_unsafe"},
			"wallUnsafeFlags": []string{"onsfw_mod_unsafe_wall"},
		},
		AppID: AppID,
	}

	err := gc.Publish(eventToGuiver)
	Convey("When calling gc.Publish", t, func() {
		Convey("err ShouldBeNil", func() {
			So(err, ShouldBeNil)
		})
	})
}
