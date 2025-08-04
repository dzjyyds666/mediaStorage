package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/core"
	"github.com/smartystreets/goconvey/convey"
)

const (
	applyUpload = "/media/upload/apply"
)

var (
	hcli = http.Client{
		Timeout: time.Second * 10,
	}

	baseURL = "http://127.0.0.1:18080"
)

func Test_Applyload(t *testing.T) {
	convey.Convey("Applyload", t, func() {
		file, err := os.Open("/Users/aaron/Downloads/assets/photoes/头像01.jpg")
		convey.So(err, convey.ShouldBeNil)
		defer file.Close()

		stat, err := file.Stat()
		convey.So(err, convey.ShouldBeNil)
		var init core.InitUpload

		init.ContentLength = ptr.Int64(stat.Size())
		init.FileName = ptr.String(stat.Name())
		init.ContentType = ptr.String(http.DetectContentType([]byte(stat.Name())))
		init.Uploader = ptr.String("aaron")

		buf, err := json.Marshal(init)
		convey.So(err, convey.ShouldBeNil)

		req, err := http.NewRequest(http.MethodPost, baseURL+applyUpload, bytes.NewBuffer(buf))
		convey.So(err, convey.ShouldBeNil)

		resp, err := hcli.Do(req)
		convey.So(err, convey.ShouldBeNil)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		convey.So(err, convey.ShouldBeNil)
		t.Logf("body: %s", string(body))
	})
}
