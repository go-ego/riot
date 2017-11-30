benchmark test
====

Test code: [examples/benchmark/benchmark.go](/examples/benchmark/benchmark.go)

The test data is one hundred thousand tweets from 52 Weibo accounts (Please download from [here](https://raw.githubusercontent.com/huichen/wukong/43f20b4c0921cc704cf41fe8653e66a3fcbb7e31/testdata/weibo_data.txt), then copy to the testdata directory), by benchmark.go in the-num_repeat_text parameter (set to 10) repeat the index for one million, 500M text。Test environment: Intel (R) Xeon (R) CPU E5-2650 v2 @ 2.60GHz 32-core, 128G memory。

Changing the NumShards variable in the test program can change the number of single data sharding. The greater the split, the higher the single request, the smaller the delay. However, the total number of requests that can be processed per second also decreases, such as：

- 1 shard: indexing 1.3M index entries per second, 1.65 milliseconds response time , throughput of 19.3K searches per second
- 2 shards: indexing 1.8M index entries per second, 0.87 milliseconds response time, throughput 18.4K searches per second
- 4 shards: indexing 1.9M index entries per second, 0.56 milliseconds response time, throughput 14.3K searches per second
- 8 shards: indexing 2.0M index entries per second, 0.39 milliseconds response time, throughput 10.3K searches per second

The index item here refers to a non-repeating "search key" - "document" pair. For example, when there are N different search keys in a document, the document generates N index items.

The program uses 8 shards by default, and you can change this value when initializing the engine based on your specific requirements, see [types.EngineOpts.NumShards](/types/engine_init_options.go)

# performance analysis

benchmark.go can also help you find the engine's CPU and memory bottlenecks.

Analyze performance bottlenecks：
```
go build benchmark.go
./benchmark -cpuprofile=cpu.prof
go tool pprof benchmark cpu.prof
```

Enter the pprof terminal after entering the web command can generate a similar figure below, clearly shows the CPU time consumed by each component:
![](https://raw.github.com/go-ego/riot/master/docs/zh/cpu.png)

Analyze memory usage：
```
go build benchmark.go
./benchmark -memprofile=mem.prof
go tool pprof benchmark mem.prof
```

The use of pprof [see this article](http://blog.golang.org/profiling-go-programs).
