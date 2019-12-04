package otakumo

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEntityUpdate(t *testing.T) {
	os.Setenv("REDIS_OTAKUMO_UPDATE_ADDRESS", "127.0.0.1:6379")
	os.Setenv("REDIS_OTAKUMO_UPDATE_CHANNEL", "objectUpdate")
	InitEntityUpdateListener()

	Convey("Subscribe to a entityType mrs-serie", t, func() {
		var called = false
		SubscribeEntityType("mrs-serie", func(otakumoID string) {
			if otakumoID == "mrs-serie-123" {
				called = true
			}
		})

		SendEntityUpdated("mrs-chapter-123")
		So(called, ShouldBeFalse)

		SendEntityUpdated("mrs-serie-123")
		time.Sleep(100 * time.Millisecond)
		So(called, ShouldBeTrue)

	})

	Convey("Unsubscribe to a command", t, func() {
		var called = false
		SubscribeEntityType("mrs-serie", func(otakumoID string) {
			if otakumoID == "mrs-serie-123" {
				called = true
			}
		})

		SendEntityUpdated("mrs-serie-123")
		time.Sleep(100 * time.Millisecond)
		So(called, ShouldBeTrue)

		UnsubscribeEntityType("mrs-serie")
		called = false
		SendEntityUpdated("mrs-serie-123")
		time.Sleep(100 * time.Millisecond)
		So(called, ShouldBeFalse)

	})
}
