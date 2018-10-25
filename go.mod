module github.com/go-ego/riot

require (
	github.com/AndreasBriese/bbloom v0.0.0-20170702084017-28f7e881ca57
	// github.com/BurntSushi/toml v0.3.0
	github.com/StackExchange/wmi v0.0.0-20180725035823-b12b22c5341f
	// github.com/coreos/bbolt v1.3.0
	// github.com/coreos/etcd v3.3.9+incompatible
	github.com/dgraph-io/badger v1.5.4
	github.com/dgryski/go-farm v0.0.0-20180109070241-2de33835d102
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-ego/gpy v0.0.0-20180905182312-c381ef5b982f
	github.com/go-ego/gse v0.0.0-20181025134840-c539b6c2f5ae
	github.com/go-ego/murmur v0.0.0-20181019172721-24868d0e6e66
	github.com/go-ole/go-ole v1.2.1
	github.com/go-vgo/grpclb v0.0.0-20180520124126-434b4da1cea2
	github.com/go-vgo/gt v0.0.0-20180924174828-283b79133891
	github.com/gogo/protobuf v1.1.1
	github.com/golang/protobuf v1.2.0
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.8.0
	github.com/shirou/gopsutil v0.0.0-20180801053943-8048a2e9c577
	github.com/syndtr/goleveldb v0.0.0-20180815032940-ae2bd5eed72d
	github.com/vcaesar/tt v0.0.0-20181014173808-65366af8a7be
	// not github
	go.etcd.io/bbolt v1.3.1-etcd.7
	go.etcd.io/etcd v3.3.9+incompatible
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1 // indirect
	golang.org/x/net v0.0.0-20180702212446-ed29d75add3d
	golang.org/x/sys v0.0.0-20180329131831-378d26f46672 // indirect
	golang.org/x/text v0.3.0 // indirect
	google.golang.org/grpc v1.14.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0-20170531160350-a96e63847dc3 // indirect
)

replace (
	go.etcd.io/bbolt v1.3.1-etcd.7 => github.com/etcd-io/bbolt v1.3.1-etcd.7
	go.etcd.io/etcd v3.3.9+incompatible => github.com/etcd-io/etcd v3.3.9+incompatible
	golang.org/x/net v0.0.0-20180702212446-ed29d75add3d => github.com/golang/net v0.0.0-20180702212446-ed29d75add3d
	golang.org/x/sys v0.0.0-20180329131831-378d26f46672 => github.com/golang/sys v0.0.0-20180329131831-378d26f46672
	golang.org/x/text v0.3.0 => github.com/golang/text v0.3.0
	google.golang.org/grpc v1.14.0 => github.com/grpc/grpc-go v1.14.0
)
