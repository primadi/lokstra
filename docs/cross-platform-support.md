# Cross-Platform Support

Lokstra framework now supports multiple operating systems and architectures through platform-specific implementations.

## Supported Platforms

- **Windows**: AMD64
- **Linux**: AMD64, ARM64  
- **macOS/Darwin**: AMD64, ARM64 (Apple Silicon)

## Platform-Specific Features

### Health Check System

The health check system provides platform-specific implementations for optimal performance:

#### Windows (`checkers.go`)
- Uses Windows API `GetDiskFreeSpaceEx` for accurate disk usage
- Native Windows syscalls for system monitoring
- Build constraint: `//go:build windows`

#### Linux (`checkers_linux.go`)
- Uses `syscall.Statfs` for disk usage monitoring
- Linux-specific system calls
- Build constraint: `//go:build linux`

#### macOS/Darwin (`checkers_darwin.go`)
- Uses `syscall.Statfs` for disk usage monitoring
- Darwin-specific system calls  
- Build constraint: `//go:build darwin`

#### Other Platforms (`checkers_fallback.go`)
- Fallback implementation for unsupported platforms
- Provides mock values for disk usage
- Build constraint: `//go:build !windows && !linux && !darwin`

## Building for Different Platforms

### Using Build Scripts

#### PowerShell (Windows)
```powershell
# Build all platforms
.\build.ps1

# Build with custom version
.\build.ps1 -Version "2.0.0" -AppName "myapp"
```

#### Bash (Linux/macOS)
```bash
# Build all platforms
./build.sh

# Build with custom version
VERSION=2.0.0 ./build.sh
```

### Manual Cross-Compilation

#### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o lokstra-windows.exe .
```

#### Linux AMD64
```bash
GOOS=linux GOARCH=amd64 go build -o lokstra-linux-amd64 .
```

#### Linux ARM64
```bash
GOOS=linux GOARCH=arm64 go build -o lokstra-linux-arm64 .
```

#### macOS AMD64
```bash
GOOS=darwin GOARCH=amd64 go build -o lokstra-darwin-amd64 .
```

#### macOS ARM64 (Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -o lokstra-darwin-arm64 .
```

## Runtime Behavior

### Health Checks

All platforms provide identical health check APIs:

```go
// Database health check (cross-platform)
dbChecker := DatabaseHealthChecker(dbPool)

// Memory health check (cross-platform)
memChecker := MemoryHealthChecker(1024) // 1GB limit

// Disk health check (platform-specific implementation)
diskChecker := DiskHealthChecker("/var/log", 80.0) // 80% threshold

// Redis health check (cross-platform)
redisChecker := RedisHealthChecker(redisClient)
```

### Platform Detection

The appropriate implementation is automatically selected at compile time using Go build constraints.

### Disk Usage Monitoring

| Platform | Implementation | API Used |
|----------|----------------|----------|
| Windows | `checkers.go` | `GetDiskFreeSpaceEx` |
| Linux | `checkers_linux.go` | `syscall.Statfs` |
| macOS | `checkers_darwin.go` | `syscall.Statfs` |
| Others | `checkers_fallback.go` | Mock values |

## Deployment Considerations

### Linux Deployment
- Supports both AMD64 and ARM64 architectures
- Optimized for container environments
- Uses native Linux syscalls for performance

### Windows Deployment
- Full Windows API integration
- Service deployment support
- Native Windows performance monitoring

### macOS Deployment
- Supports both Intel and Apple Silicon Macs
- Native Darwin syscalls
- Compatible with macOS security restrictions

## Configuration

No platform-specific configuration is required. The health check thresholds and paths are configured the same way across all platforms:

```yaml
health_check:
  memory_limit_mb: 1024
  disk_usage_threshold: 80.0
  disk_path: "/var/log"  # or "C:\\" on Windows
```

## Troubleshooting

### Build Issues
1. Ensure Go 1.16+ for embed support
2. Check GOOS/GOARCH environment variables
3. Verify platform-specific dependencies

### Runtime Issues
1. Check file permissions on Unix systems
2. Verify disk paths exist and are accessible
3. Ensure syscall permissions are available

### Platform Detection
Use `runtime.GOOS` and `runtime.GOARCH` to detect the current platform at runtime if needed.
