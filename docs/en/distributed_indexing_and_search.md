Distributed indexing and searching
===

The principle of distributed search is as follows：

When a large number of documents can not be indexed in one machine's memory, the document can be sharded according to the hash value of the text content, and different blocks are indexed by different servers. The same request is distributed to all split servers at lookup time, and the results returned by all servers are then reordered to be output as the final search result.

In order to ensure the uniformity of split, it is recommended to use the Go language Murmur3 hash function:

https://github.com/go-ego/murmur

In accordance with the above principle is very easy to use the riot engine for distributed search (split server running a riot engine), but most of such distributed systems are highly customized, such as task scheduling depends on the distributed environment, and sometimes need to add The extra layer of server to load balance, it is not here to achieve.
