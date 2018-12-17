Persistent storage
====

The riot engine supports saving search data to the hard drive and restoring data from the hard drive when the machine restarts. Just use persistent storage to set the three options in EngineOpts:

```go
type EngineOpts struct {
  // Skip other options

  // Whether to use persistent databases, and the number of directories and splits that database files hold
  UseStore bool
  StoreFolder string
  StoreShards int
  StoreEngine string // bg: badger, lbd: leveldb, bolt: bolt
}
```

Use persistent storage when UseStore is true:

1. At engine startup (engine.Init function), the engine reads the document index data from the directory specified by StoreFolder, recalculates the index table, and injects the sort data into the sequencer. If the tokenizer code or dictionary changes, these changes will be reflected in the engine index table after the start.
2. When the engine.Index is called, the engine writes the index data to the directory specified by StoreFolder.
3. StoreShards defines the number of database splits, the default is 8. For best performance, adjust this parameter so that each split file is less than 100M.
4. After calling engine.RemoveDoc to delete a document, the document is removed from persistent storage and will not be loaded the next time the engine is started.


### Must pay attention to matters

1. If the sorter uses [custom score field](/docs/en/custom_scoring_criteria.md) then that type must be registered in the gob, as in the example on the left which needs to be added before calling engine.Init：
```
gob.Register(MyScoringFields{})
```
Otherwise the program will crash.

2. Please use engine.Close () to shut down the database when the engine exits. If the database is not closed, the database files will be locked, which will cause the engine to restart failed. To unlock it, go to the directory specified by StoreFolder and delete any files that begin with ".".

### Benchmark

[benchmark.go](/examples/benchmark.go) program can be used to test persistent storage read and write speed:

```go
go run benchmark.go --num_repeat_text 1 --use_persistent
```

Do not use the results of persistent storage:

```
The total number of indexes added 3711375
Indexing takes time 3.159781129s
Indexing speed added per second 1.174567 Millions of indexes
The average search response time is 0.051146982400000006 milliseconds
Search throughput of 78205.98229466612 queries per second
```

Use persistent storage:

```
Indexing takes time 13.068458511s
Indexing speed added 0.283995 million indexes per second
The average search response time is 0.05819595780000001 milliseconds
Search throughput of 68733.29611219149 queries per second
The total number of indexes joined from persistent storage is 3711375
It takes 6.406528378999999 to build an index from persistent storage
Indexing from persistent storage adds 0.579311 million indexes per second
```

As you can see, compared to not using persistent storage:

- [persistent storage](#%E6%8C%81%E4%B9%85%E5%AD%98%E5%82%A8)
        - [must pay attention to matters](#%E5%BF%85%E9%A1%BB%E6%B3%A8%E6%84%8F%E4%BA%8B%E9%A1%B9)
        - [performance Testing](#%E6%80%A7%E8%83%BD%E6%B5%8B%E8%AF%95)
