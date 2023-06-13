module fajurion.com/voice-node

go 1.20

require (
	fajurion.com/node-integration v0.0.0-00010101000000-000000000000
	github.com/Fajurion/pipes v0.0.0-00010101000000-000000000000
	github.com/dgraph-io/ristretto v0.1.1
	github.com/gofiber/fiber/v2 v2.44.0
)

require nhooyr.io/websocket v1.8.7 // indirect

require (
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/bytedance/sonic v1.8.8
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/cornelk/hashmap v1.0.8 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/klauspost/compress v1.16.5 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/savsgio/dictpool v0.0.0-20221023140959-7bf2e61cea94 // indirect
	github.com/savsgio/gotils v0.0.0-20230208104028-c358bd845dee // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.46.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
)

replace fajurion.com/node-integration => ./node-integration

replace github.com/Fajurion/pipes => ./pipes
