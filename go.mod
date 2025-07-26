module github.com/dzjyyds666/mediaStorage

go 1.23.3

replace (
	github.com/dzjyyds666/Allspark-go => ../Allspark-go
	github.com/dzjyyds666/vortex/v2 => ../vortex/v2
)

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/dzjyyds666/Allspark-go v0.0.0-20250726101904-c957bdf200df
	github.com/dzjyyds666/vortex/v2 v2.0.0-00010101000000-000000000000
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.3 // indirect
	github.com/labstack/echo/v4 v4.13.4 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)
