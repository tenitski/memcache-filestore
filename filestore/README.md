# File store library

## Supported backends

- Memcache


### Memcache

Memcache backend requires server details and allows to configure certain parameters:

```go
s := store.NewMemcache("127.0.0.1:11211", store.MemcacheConfig{
    Timeout:     100 * time.Millisecond,
    ChunkSize:   1024 * 1024,
    MaxFileSize: 50 * 1024 * 1024,
})
```

## Usage

```go
// Init client
c := store.NewMemcache("127.0.0.1:11211", store.MemcacheConfig{})

// Store file
err := c.Store(filename, data)
if err != nil {
    fmt.Printf("Unable to store file: %s", err.Error())
}

// Get file 
value, err := c.Retrieve(filename)
if err != nil {
    fmt.Printf("Unable to retrieve file: %s", err.Error())
}
log.WithField("contents", string(value)).Info("File contents")

// Delete file
err = c.Delete(filename)
if err != nil {
    fmt.Printf("Unable to delete file: %s", err.Error())
}
```

## Example

See [this example](example/main.go)


## Testing

```bash
cd filestore
go test ./...
```
