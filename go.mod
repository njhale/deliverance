module github.com/ecordell/bndlr

go 1.12

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	// contains a fix for talking to quay - next semver release should be used when released
	github.com/containerd/containerd v1.3.1-0.20191014151319-9c86b8f5ed49
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6 // indirect
	github.com/deislabs/oras v0.7.0
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	golang.org/x/crypto v0.0.0-20190611184440-5c40567a22f8 // indirect
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859 // indirect
	golang.org/x/sys v0.0.0-20190621203818-d432491b9138 // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	rsc.io/letsencrypt => github.com/dmcgowan/letsencrypt v0.0.0-20160928181947-1847a81d2087
)
