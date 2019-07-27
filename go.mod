module github.com/go-ego/riot

require (
	// github.com/coreos/etcd v3.3.10+incompatible // indirect
	// github.com/coreos/bbolt v1.3.0
	// github.com/coreos/etcd v3.3.9+incompatible
	github.com/dgraph-io/badger v1.6.0
	github.com/go-ego/gpy v0.0.0-20181128170341-b6d42325845c
	github.com/go-ego/murmur v0.0.0-20181129155752-fac557227e04
	github.com/go-vgo/grpclb v0.0.0-20181128171039-89526b0a742e
	github.com/go-vgo/gt v0.0.0-20181207163017-e40d098f9006
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/shirou/gopsutil v2.18.11+incompatible
	github.com/syndtr/goleveldb v0.0.0-20181128100959-b001fa50d6b2
	github.com/vcaesar/tt v0.0.0-20190128173424-2874d9a357d3
	// not github
	go.etcd.io/bbolt v1.3.3
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	// golang.org/x/lint v0.0.0-20181026193005-c67002cb31c3 // indirect
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
	golang.org/x/sys v0.0.0-20190726091711-fc99dfbffb4e // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/grpc v1.22.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
// honnef.co/go/tools v0.0.0-20180920025451-e3ad64cb4ed3 // indirect
)

replace (
	go.etcd.io/bbolt v1.3.1-etcd.7 => github.com/etcd-io/bbolt v1.3.1-etcd.7
	go.etcd.io/etcd v3.3.10+incompatible => github.com/etcd-io/etcd v3.3.10+incompatible
	golang.org/x/net v0.0.0-20181114220301-adae6a3d119a => github.com/golang/net v0.0.0-20181114220301-adae6a3d119a
	golang.org/x/sys v0.0.0-20181122145206-62eef0e2fa9b => github.com/golang/sys v0.0.0-20181122145206-62eef0e2fa9b
	golang.org/x/text v0.3.0 => github.com/golang/text v0.3.0
	google.golang.org/grpc v1.17.0 => github.com/grpc/grpc-go v1.17.0
)
