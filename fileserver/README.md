# File store server

Depends on `filestore` library with memcache backend.

## Usage

Build:
```bash 
go build
```

Start the server:

```bash
LOG_LEVEL=debug ./fileserver MEMCACHE_HOST:MEMCACHE_PORT 
```

Make requests:

```bash
# Store file
curl --data-binary "@/path/to/myfile.dat" http://127.0.0.1:8080/file/myfile.dat

# Retrieve file
curl http://127.0.0.1:8080/file/myfile.dat > myfile.dat

# Delete file
curl -X DELETE -v http://127.0.0.1:8080/file/myfile.dat
```

## Testing

```bash
cd filestore
go test ./...
```
