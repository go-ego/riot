module github.com/oGre222/tea

require (
	cloud.google.com/go v0.34.0 // indirect
	github.com/AndreasBriese/bbloom v0.0.0-20180913140656-343706a395b7
	// github.com/BurntSushi/toml v0.3.0
	github.com/StackExchange/wmi v0.0.0-20180725035823-b12b22c5341f
	github.com/coreos/etcd v3.3.12+incompatible // indirect
	// github.com/coreos/etcd v3.3.10+incompatible // indirect
	// github.com/coreos/bbolt v1.3.0
	// github.com/coreos/etcd v3.3.9+incompatible
	github.com/dgraph-io/badger v1.5.4
	github.com/dgryski/go-farm v0.0.0-20180109070241-2de33835d102
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-ego/gpy v0.0.0-20181128170341-b6d42325845c
	github.com/go-ego/gse v0.0.0-20190129002529-45bfc8a37d9a
	github.com/go-ego/murmur v0.0.0-20181129155752-fac557227e04
	github.com/go-ole/go-ole v1.2.1
	github.com/go-vgo/grpclb v0.0.0-20181128171039-89526b0a742e
	github.com/go-vgo/gt v0.0.0-20181207163017-e40d098f9006
	github.com/gogo/protobuf v1.2.0
	github.com/golang/mock v1.2.0 // indirect
	// github.com/golang/lint v0.0.0-20181026193005-c67002cb31c3 // indirect
	github.com/golang/protobuf v1.2.0
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db
	github.com/google/btree v1.0.0 // indirect
	github.com/json-iterator/go v1.1.6
	github.com/juju/errors v0.0.0-20190207033735-e65537c515d7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pelletier/go-toml v1.2.0
	github.com/pingcap/kvproto v0.0.0-20190415114642-9811bf233a01 // indirect
	github.com/pingcap/tidb v0.0.0-20190325083614-d6490c1cab3a
	github.com/pingcap/tipb v0.0.0-20190415042426-af5d98495e74 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20190321074620-2f0d2b0e0001 // indirect
	github.com/shirou/gopsutil v2.18.11+incompatible
	github.com/sirupsen/logrus v1.4.1 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/syndtr/goleveldb v0.0.0-20181128100959-b001fa50d6b2
	github.com/uber/jaeger-client-go v2.16.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	github.com/vcaesar/tt v0.0.0-20190128173424-2874d9a357d3
	// not github
	go.etcd.io/bbolt v1.3.1-etcd.7
	go.etcd.io/etcd v3.3.10+incompatible
	// golang.org/x/lint v0.0.0-20181026193005-c67002cb31c3 // indirect
	golang.org/x/net v0.0.0-20190108225652-1e06a53dbb7e
	golang.org/x/oauth2 v0.0.0-20181203162652-d668ce993890 // indirect
	google.golang.org/grpc v1.17.0
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
