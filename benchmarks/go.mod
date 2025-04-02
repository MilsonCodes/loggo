module loggo/benchmarks

go 1.24.1

require (
	github.com/rs/zerolog v1.31.0
	github.com/sirupsen/logrus v1.9.3
	go.uber.org/zap v1.27.0
	loggo v0.0.0
)

replace loggo => ../

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
)
