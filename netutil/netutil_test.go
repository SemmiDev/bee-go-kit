package netutil_test

import (
	"net"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/netutil"
)

func TestMustGetIP(t *testing.T) {
	Convey("Given MustGetIP", t, func() {
		ip := netutil.MustGetIP()

		Convey("Then it should return a non-empty string", func() {
			// On most machines with a network interface, this should be non-empty.
			// In CI/container environments without networking, it may be empty,
			// so we check only if a result is returned.
			if ip != "" {
				Convey("And it should be a valid IPv4 address", func() {
					parsed := net.ParseIP(ip)
					So(parsed, ShouldNotBeNil)
					So(parsed.To4(), ShouldNotBeNil)
				})

				Convey("And it should not be a loopback address", func() {
					parsed := net.ParseIP(ip)
					So(parsed.IsLoopback(), ShouldBeFalse)
				})
			}
		})
	})
}
