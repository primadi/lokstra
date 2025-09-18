# Gzip Compression Middleware

Gzip compression middleware for compressing HTTP responses. This middleware automatically compresses responses using gzip when supported by the client, improving bandwidth usage and load times.

## Features

- Automatic gzip compression for supported clients
- Configurable minimum size for compression
- Customizable compression level
- Typed middleware and config recommended
- Easy integration with router

## Configuration

```go
type Config struct {
    MinSize         int  `json:"min_size" yaml:"min_size"`         // Minimum response size to trigger compression (bytes)
    Level int `json:"level" yaml:"level"` // Gzip compression level (1-9)
}
```

### Configuration Parameters

- `min_size` (int): Minimum response size in bytes to enable compression (default: `1024`)
- `level` (int): Gzip compression level (1 = fastest, 9 = best compression, default: `5`)

## Usage

### 1. Typed Middleware (Recommended)

It is recommended to use typed middleware with typed config:

```go
import "github.com/primadi/lokstra/middleware/gzipcompression"

config := &gzipcompression.Config{
    MinSize:          1024, // Only compress responses larger than 1KB
    CompressionLevel: 6,    // Use moderate compression
}
router.Use(gzipcompression.GetMidware(config))
```

### 2. Basic Usage (default config)

```go
// Use default gzip compression config
router.Use("gzipcompression")
```

### 3. With Map Configuration

```go
config := map[string]any{
    "min_size": 2048,
    "compression_level": 9,
}
router.Use("gzipcompression", config)
```

### 4. With Struct Config

```go
config := &gzipcompression.Config{
    MinSize:          512,
    CompressionLevel: 3,
}
router.Use("gzipcompression", config)
```

## Example Response Header

When compression is applied, the following header is set:

```
Content-Encoding: gzip
```

## Best Practices

- Use typed middleware and config for type safety and clarity
- Adjust `min_size` and `level` based on performance needs
- Enable compression only for clients that support gzip
- Avoid compressing already compressed content types (e.g., images, zip files)

## Testing

Run tests for the gzip compression middleware:

```bash
go test ./middleware/gzipcompression -v
```

Tests cover:
- Configuration parsing in various formats
- Compression behavior for different response sizes
- Compression level effects
- Edge cases and error handling

(See gzipcompression_test.go for details.)
