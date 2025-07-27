package core

import (
	"testing"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/smartystreets/goconvey/convey"
)

func Test_LoadConfig(t *testing.T) {
	convey.Convey("LoadConfig", t, func() {
		cfg, err := LoadConfig("/Users/aaron/code/GolandProjects/mediaStorage/conf/storage.toml")
		convey.So(err, convey.ShouldBeNil)
		convey.So(cfg, convey.ShouldNotBeNil)
		t.Logf("cfg: %+v", conv.ToJsonWithoutError(cfg))
	})
}
