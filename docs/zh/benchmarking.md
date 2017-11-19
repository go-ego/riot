性能测试
====

测试程序见 [examples/benchmark/benchmark.go](/examples/benchmark/benchmark.go)

测试数据为从 52个微博账号里抓取的十万条微博（请从[这里](https://raw.githubusercontent.com/huichen/wukong/43f20b4c0921cc704cf41fe8653e66a3fcbb7e31/testdata/weibo_data.txt)下载，然后 copy 到 testdata 目录），通过benchmark.go 中的 -num_repeat_text 参数（设为10）重复索引为一百万条，500M 文本。测试环境Intel(R) Xeon(R) CPU E5-2650 v2 @ 2.60GHz 32 核，128G 内存。

改变测试程序中的 NumShards 变量可以改变数据单机裂分（sharding）的数目，裂分越多单请求的并发度越高延迟越小，但相应地每秒能处理的总请求数也会变少，比如：

- 1个 shard 时：每秒索引 1.3M 个索引项，响应时间 1.65毫秒，吞吐量每秒 19.3K 次搜索
- 2个 shard 时：每秒索引 1.8M 个索引项，响应时间 0.87毫秒，吞吐量每秒 18.4K 次搜索
- 4个 shard 时：每秒索引 1.9M 个索引项，响应时间 0.56毫秒，吞吐量每秒 14.3K 次搜索
- 8个 shard 时：每秒索引 2.0M 个索引项，响应时间 0.39毫秒，吞吐量每秒 10.3K 次搜索

这里的索引项是指一个不重复的“搜索键”-“文档”对，比如当一个文档中有N个不一样的搜索键时，该文档会产生N个索引项。

程序默认使用 8个 shard，你可以根据具体的需求在初始化引擎时改变这个值，见[types.EngineInitOptions.NumShards](/types/engine_init_options.go)

# 性能分析

benchmark.go 也可以帮助你找到引擎的 CPU 和内存瓶颈在哪里。

分析性能瓶颈：
```
go build benchmark.go
./benchmark -cpuprofile=cpu.prof
go tool pprof benchmark cpu.prof
```

进入 pprof 终端后输入 web 命令可以生成类似下面的图，清晰地表示了每个组件消耗的 CPU 时间

![](https://raw.github.com/go-ego/riot/master/docs/zh/cpu.png)

分析内存占用：
```
go build benchmark.go
./benchmark -memprofile=mem.prof
go tool pprof benchmark mem.prof
```

pprof 的使用见[这篇文章](http://blog.golang.org/profiling-go-programs)。
