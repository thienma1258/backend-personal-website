package mrcommand

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMRCommand(t *testing.T) {
	os.Setenv("REDIS_MRCOMMAND_ADDRESS", "127.0.0.1:6379")
	os.Setenv("REDIS_MRCOMMAND_CHANNEL", "mrchannel")
	Init()

	Convey("Subscribe to a command", t, func() {
		var called = false
		SubscribeCommand("CommandA", func() {
			called = true
		})

		SendCommand("CommandB")
		So(called, ShouldBeFalse)

		SendCommand("CommandA")
		time.Sleep(100 * time.Millisecond)
		So(called, ShouldBeTrue)

	})

	Convey("Unsubscribe to a command", t, func() {
		var called = false
		SubscribeCommand("CommandA", func() {
			called = true
		})

		SendCommand("CommandA")
		time.Sleep(100 * time.Millisecond)
		So(called, ShouldBeTrue)

		UnsubscribeCommand("CommandA")
		called = false
		SendCommand("CommandA")
		time.Sleep(100 * time.Millisecond)
		So(called, ShouldBeFalse)

	})
}
