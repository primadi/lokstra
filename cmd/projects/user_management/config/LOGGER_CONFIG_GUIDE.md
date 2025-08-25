# Logger Configuration Examples for Lokstra Framework

## 1. Development Configuration (user_management.yaml)
```yaml
server:
  global_setting:
    # Development settings - easier to read
    log_level: debug          # Detailed logging for development
    log_format: text          # Human-readable format
    log_output: stdout        # Console output
    flow_logger: logger
```

## 2. Production Configuration
```yaml
server:
  global_setting:
    # Production settings - structured and efficient
    log_level: ${ENV:LOG_LEVEL:info}        # Environment variable with default
    log_format: ${ENV:LOG_FORMAT:json}      # JSON for log aggregation tools
    log_output: ${ENV:LOG_OUTPUT:/var/log/user_management.log}  # File output
    flow_logger: logger
```

## 3. Available Log Levels (in order of verbosity)
- `debug`: Detailed debugging information
- `info`: General information (default)
- `warn`: Warning messages
- `error`: Error messages only
- `fatal`: Fatal errors only

## 4. Available Log Formats
- `text`: Human-readable format (good for development)
- `json`: Structured JSON format (good for production/log aggregation)

## 5. Available Log Outputs
- `stdout`: Standard output (console)
- `stderr`: Standard error output
- `/path/to/file.log`: File path for log file output

## 6. Environment Variables for Production
```bash
# .env file or system environment
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=/var/log/user_management.log
```

## 7. Sample Log Outputs

### Text Format (Development)
```
2025-08-25T10:30:00Z INFO [user-management] Starting server on :8080
2025-08-25T10:30:01Z DEBUG [user-management] Handler registered: user.create
2025-08-25T10:30:05Z INFO [user-management] User created: johndoe
```

### JSON Format (Production)
```json
{"timestamp":"2025-08-25T10:30:00Z","level":"info","service":"user-management","message":"Starting server on :8080"}
{"timestamp":"2025-08-25T10:30:01Z","level":"debug","service":"user-management","message":"Handler registered: user.create"}
{"timestamp":"2025-08-25T10:30:05Z","level":"info","service":"user-management","message":"User created: johndoe","user_id":"123"}
```

## 8. Why NOT to Define Logger Service Explicitly

❌ **WRONG** - Don't do this in services.yaml:
```yaml
services:
  - name: "logger"          # This will conflict with auto-created logger
    type: "lokstra.logger"
    config:
      level: "info"
```

✅ **CORRECT** - Use global_setting instead:
```yaml
server:
  global_setting:
    log_level: info         # Configure the auto-created logger
    log_format: json
    log_output: stdout
```

## 9. Benefits of Using global_setting
1. **No Conflicts**: Framework automatically creates logger service
2. **Consistent**: All flows and handlers use the same logger configuration
3. **Simple**: One place to configure logging for entire application
4. **Environment Support**: Easy to override with environment variables

## 10. Advanced Configuration Example
```yaml
server:
  name: user_management_service
  global_setting:
    # Logger configuration with environment override
    log_level: ${ENV:LOG_LEVEL:info}
    log_format: ${ENV:LOG_FORMAT:json}
    log_output: ${ENV:LOG_OUTPUT:stdout}
    
    # Additional flow settings
    flow_dbPool: db_global
    flow_logger: logger
    
    # Request timeout and other global settings
    default_timeout: "30s"
    max_request_size: "10MB"
```

This approach ensures that:
- Logger service is automatically created and configured
- No naming conflicts occur
- Configuration is centralized and environment-friendly
- Easy to override for different deployment environments
