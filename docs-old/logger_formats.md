# Logger Service Format Support

The Lokstra logger service now supports multiple output formats for different use cases.

## Supported Formats

### 1. JSON Format (Default)
- **Use Case**: Production environments, log aggregation systems
- **Output**: Structured JSON logs that are easy to parse by log collectors
- **Example Output**:
  ```json
  {"level":"info","time":"2025-08-19T10:30:45Z","message":"User created successfully","user_id":123}
  ```

### 2. Console/Text Format
- **Use Case**: Development, debugging, human readability
- **Output**: Colored, formatted text that's easy to read in terminal
- **Example Output**:
  ```
  10:30AM INF User created successfully user_id=123
  ```

## Configuration Examples

### Basic JSON Logger
```yaml
services:
  - name: "logger"
    type: "lokstra.logger"
    config:
      level: "info"
      format: "json"
      output: "stdout"
```

### Development Console Logger
```yaml
services:
  - name: "logger"
    type: "lokstra.logger"
    config:
      level: "debug"
      format: "console"
      output: "stdout"
      caller: true      # Show file/line information
      stacktrace: true  # Show stack trace for errors
```

### File Output with Rotation
```yaml
services:
  - name: "logger"
    type: "lokstra.logger"
    config:
      level: "info"
      format: "json"
      output: "file"
      file_path: "./logs/app.log"
      max_size: 100     # Max file size in MB
      max_backups: 5    # Number of old files to keep
      max_age: 30       # Days to keep old files
      compress: true    # Compress rotated files
```

## Format Comparison

| Feature | JSON | Console/Text |
|---------|------|--------------|
| Machine Readable | ✅ | ❌ |
| Human Readable | ❌ | ✅ |
| Colored Output | ❌ | ✅ |
| Structured Fields | ✅ | ✅ |
| Log Aggregation | ✅ | ❌ |
| Development Debug | ❌ | ✅ |

## Best Practices

1. **Production**: Use JSON format for production environments
2. **Development**: Use console format for local development
3. **File Logging**: Enable rotation to prevent disk space issues
4. **Performance**: JSON format is slightly faster for high-throughput logging
5. **Debugging**: Enable caller and stacktrace for development environments

## Implementation Details

The logger service uses [zerolog](https://github.com/rs/zerolog) library:
- JSON format uses zerolog's default JSON encoder
- Console format uses `zerolog.ConsoleWriter` with colored output
- All formats support structured logging with fields
