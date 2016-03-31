### LiteJob

A Fast & Safely Golang Task Dispatch.

### Features

* more storage engine support
    1. Redis(storage job list with redis list)
    2. Sqlite(storage job list with a table)
    3. Memory(storage job list with a memory map)

    ... others

* more task handle way
    1. Native  (handle a native func)
    2. Http1.1 (handle job with http/1.1 post)
    3. Http2
    4. Rpc Supported
       1. Yar
       2. ... more
    5. ... more

* plugin easy

### RoadMap



1. support Redis,Sqlite,Memory storage
2. support Http And Rpc