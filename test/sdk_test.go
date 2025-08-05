package test

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
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

var (
	endpoint = "http://127.0.0.1:18080"
	jwtToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmVzIjoxNzU0NDA5OTUwLCJ1aWQiOiJhYXJvbiJ9.63e-SoAsLK9WkngTFnbQEJtMbROPg6ASw-NSaiOIxIU"
	hcli     = &http.Client{
		Timeout: time.Second * 10,
	}
)

const (
	loginPath   = "/login"
	applyUpload = "/media/upload/apply"
)

func Test_Login(t *testing.T) {
	convey.Convey("登录", t, func() {
		loginReq := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: "aaron",
			Password: "aaron519",
		}

		body, err := json.Marshal(loginReq)
		convey.So(err, convey.ShouldBeNil)
		req, err := http.NewRequest(http.MethodPost, endpoint+loginPath, bytes.NewBuffer(body))
		convey.So(err, convey.ShouldBeNil)
		req.Header.Set("Content-Type", "application/json")
		resp, err := hcli.Do(req)
		convey.So(err, convey.ShouldBeNil)
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		convey.So(err, convey.ShouldBeNil)
		t.Logf("raw: %s", string(raw))
	})
}

func LoadFileInfoFromLocal(path string) (*core.InitUpload, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 计算文件的md5
	h := md5.New()
	partBytes := make([]byte, 1024)
	count := 0
	contentType := "application/octet-stream"
	for {
		count++
		n, err := f.Read(partBytes)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		h.Write(partBytes[:n])
		if count == 1 {
			// 第一次读取文件的前1024字节，用于检测文件的内容类型
			contentType = http.DetectContentType(partBytes)
		}
	}
	md5Sum := hex.EncodeToString(h.Sum(nil))

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &core.InitUpload{
		FileName:      ptr.String(stat.Name()),
		ContentLength: ptr.Int64(stat.Size()),
		ContentMd5:    ptr.String(md5Sum),
		ContentType:   ptr.String(contentType),
	}, nil
}

func Test_ApplyUpload(t *testing.T) {
	convey.Convey("申请上传", t, func() {
		path := "/Users/aaron/Downloads/assets/photoes/头像01.jpg"
		init, err := LoadFileInfoFromLocal(path)
		convey.So(err, convey.ShouldBeNil)

		init.BoxId = ptr.String("default")

		body, err := json.Marshal(init)
		convey.So(err, convey.ShouldBeNil)

		req, err := http.NewRequest(http.MethodPost, endpoint+applyUpload, bytes.NewBuffer(body))
		convey.So(err, convey.ShouldBeNil)
		req.Header.Set("Authorization", jwtToken)

		resp, err := hcli.Do(req)
		convey.So(err, convey.ShouldBeNil)
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		convey.So(err, convey.ShouldBeNil)
		t.Logf("raw: %s", string(raw))
	})
}
