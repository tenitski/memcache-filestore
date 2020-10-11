# FileStore

*FileStore* is available as REST API app [fileserver](./fileserver/README.md) 
and Go lang library [filestore](./filestore/README.md).

## Concept

*FileStore* uses Memcache backend to store files.

Memcache has a limitation of 1MB per key so we have to chunk file contents and store it across multiple keys.

*FileStore* implements a simple approach with the first key storing metadata (number of chunks) and 
the subsequent keys storing the actual chunks.

No read/write locks are used because due to the nature of the storage backend (specifically the 
fact that Memcache can evict keys when it runs out of memory) files can get corrupted at any 
moment anyway. While `store` operation verifies that all chunks have been successfully written,
there is no guarantee `retrieve` would be able to read a file later.

For the same reason there is not much point worrying about issues caused by race conditions -
files getting corrupted is business as usual. 

## Notes 

- Both API handlers and the library could use streams to read/write data for better efficiency 
rather than read files into the memory and pass them around. Using streams would require 
some changes to the way *FileStore* tracks how many chunks there are: when we start storing a file 
it is yet unknown how big the file is and how many chunks would be created.

- For some reason Memcache didn't like `1048576` byte values in my setup. The max value it would 
take is `1048470`... Memcache logs show that it gets exactly the specified number of bytes, 
no envelop is added by the third party Memcache client lib. To be resolved later.

