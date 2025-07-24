# Listener Module Unit Tests

This directory contains comprehensive unit tests for the Lokstra HTTP Listener module located in `modules/coreservice/listener`.

## Overview

The listener module provides an **HTTP Server Abstraction Layer** with multiple backend implementations:

- **NetHttpListener** - Standard Go `net/http` server
- **FastHttpListener** - High-performance `fasthttp` server  
- **SecureNetHttpListener** - HTTPS server with TLS/SSL support
- **Http3Listener** - HTTP/3 server with QUIC protocol

## Test Coverage

### 1. NetHttpListener Tests (`nethttp_listener_test.go`)
- ✅ Factory function validation (`NewNetHttpListener`)
- ✅ Running state management (`IsRunning`, `ActiveRequest`)
- ✅ TCP socket support (`ListenAndServe`)
- ✅ Unix socket support (skipped on Windows)
- ✅ Graceful shutdown with active request tracking
- ✅ Shutdown timeout handling
- ✅ Concurrent request processing
- ✅ Request counting and active request tracking
- ✅ Service unavailable responses during shutdown

**Key Features Tested:**
- Standard HTTP server functionality
- Unix socket support (platform-dependent)
- Graceful shutdown waits for active requests
- Service unavailable (503) responses during shutdown
- Active request counter accuracy
- Shutdown timeout handling

### 2. FastHttpListener Tests (`fasthttp_listener_test.go`)  
- ✅ Factory function with timeout configuration
- ✅ Custom timeout handling (read, write, idle)
- ✅ High-performance fasthttp integration
- ✅ TCP socket support
- ✅ Unix socket support (skipped on Windows)
- ✅ Graceful shutdown with FastHTTP specifics
- ✅ Configuration validation and defaults
- ✅ Active request tracking
- ✅ Timeout configuration testing

**Key Features Tested:**
- FastHTTP-specific request handling
- Timeout configuration (read/write/idle)
- FastHTTP graceful shutdown behavior
- Performance-oriented request processing
- FastHTTP request adapter functionality
- Configuration error handling

### 3. SecureNetHttpListener Tests (`secure_nethttp_listener_test.go`)
- ✅ TLS certificate and key file validation
- ✅ Configuration parsing (map, array, string array)
- ✅ CA certificate support for client authentication
- ✅ HTTPS server functionality
- ✅ TLS timeout configuration
- ✅ Graceful shutdown with TLS
- ✅ Active request tracking over HTTPS
- ✅ Certificate file error handling

**Key Features Tested:**
- TLS/SSL certificate handling
- Multiple configuration formats
- CA certificate validation
- HTTPS request processing
- TLS-specific timeout handling
- Secure connection management

### 4. Http3Listener Tests (`http3_listener_test.go`)
- ✅ HTTP/3 with QUIC protocol support
- ✅ TLS 1.3 certificate validation
- ✅ Configuration parsing and validation
- ✅ HTTP/3 server functionality
- ✅ QUIC connection handling
- ✅ Graceful shutdown with HTTP/3
- ✅ Active request tracking over HTTP/3
- ✅ Idle timeout configuration

**Key Features Tested:**
- HTTP/3 protocol implementation
- QUIC transport layer
- TLS 1.3 requirement
- HTTP/3-specific request handling
- Advanced graceful shutdown
- Modern protocol features

### 5. Helper Tests (`helper_test.go`)
- ✅ Timeout constants validation
- ✅ Default timeout values verification
- ✅ Configuration key constants

**Key Features Tested:**
- Constant value correctness
- Default timeout values
- Configuration key naming

### 6. Integration Tests (`integration_test.go`)
- ✅ Interface compliance for all listeners
- ✅ Basic functionality across implementations
- ✅ TLS listener specialized testing
- ✅ Configuration parsing uniformity
- ✅ Timeout behavior consistency
- ✅ Cross-listener compatibility testing

**Key Features Tested:**
- ServiceApi.HttpListener interface compliance
- Consistent behavior across implementations
- Configuration format compatibility
- Timeout behavior standardization
- Integration between different listeners

## Test Statistics

| Test File | Tests | Coverage Areas |
|-----------|-------|----------------|
| `nethttp_listener_test.go` | 9 tests | Basic HTTP functionality, graceful shutdown |
| `fasthttp_listener_test.go` | 10 tests | High-performance HTTP, timeout configuration |
| `secure_nethttp_listener_test.go` | 7 tests | HTTPS/TLS, certificate handling |
| `http3_listener_test.go` | 8 tests | HTTP/3, QUIC, advanced protocols |
| `helper_test.go` | 3 tests | Constants and configuration |
| `integration_test.go` | 4 tests | Cross-implementation compatibility |
| **Total** | **41 tests** | **Complete listener ecosystem** |

## Running Tests

### Run All Tests
```bash
go test -v ./modules/coreservice/listener
```

### Run Specific Test Files
```bash
# Test NetHttpListener only
go test -v ./modules/coreservice/listener -run TestNetHttpListener

# Test FastHttpListener only  
go test -v ./modules/coreservice/listener -run TestFastHttpListener

# Test TLS listeners
go test -v ./modules/coreservice/listener -run "TestSecureNetHttpListener|TestHttp3Listener"

# Test integration
go test -v ./modules/coreservice/listener -run TestListener_
```

### Run with Coverage
```bash
go test -v -cover ./modules/coreservice/listener
```

## Platform Considerations

- **Unix Socket Tests**: Automatically skipped on Windows platform
- **HTTP/3 Tests**: Require QUIC support (available on most modern systems)
- **TLS Tests**: Generate temporary certificates automatically
- **Certificate Tests**: Create self-signed certificates for testing

## Test Utilities

### Certificate Generation
All TLS tests automatically generate temporary certificates:
- RSA 2048-bit keys
- Self-signed certificates
- Localhost and 127.0.0.1 SAN entries
- Automatic cleanup after tests

### Error Handling
Tests validate error conditions:
- Invalid configuration formats
- Missing certificate files
- Network binding errors
- Timeout scenarios
- Graceful shutdown edge cases

## Performance Testing

The tests include performance-oriented scenarios:
- Concurrent request handling
- Active request tracking accuracy
- Graceful shutdown timing
- Timeout behavior validation
- Resource cleanup verification

## Mock and Test Data

- **HTTP Handlers**: Simple test handlers for request validation
- **Test Certificates**: Automatically generated TLS certificates
- **Configuration Samples**: Various configuration format examples
- **Timeout Scenarios**: Different timeout values for testing edge cases

## Dependencies

Test dependencies are minimal and focused:
- Standard `testing` package
- `net/http/httptest` for HTTP testing
- `crypto` packages for certificate generation
- No external testing frameworks required

## Debugging Tests

For test debugging:
```bash
# Verbose output with detailed logging
go test -v ./modules/coreservice/listener -run TestName

# Run single test with maximum detail
go test -v ./modules/coreservice/listener -run TestSpecificFunction
```

## Contributing

When adding new tests:
1. Follow existing naming conventions (`TestListenerType_Functionality`)
2. Include both positive and negative test cases
3. Add platform-specific skips where appropriate
4. Update this README with new test descriptions
5. Ensure certificate cleanup in TLS tests
6. Test configuration error handling

## Security Testing

Security-focused test areas:
- TLS certificate validation
- CA certificate chain handling
- Client certificate authentication
- Secure protocol enforcement (TLS 1.3 for HTTP/3)
- Certificate file permission handling
- ✅ TLS timeout configuration
- ✅ Certificate file error handling
- ✅ HTTPS server functionality

**Key Features Tested:**
- TLS certificate management
- Client certificate authentication (mTLS)
- HTTPS-specific configurations
- Certificate file validation
- Secure connection handling

### 4. Http3Listener Tests (`http3_listener_test.go`)
- ✅ HTTP/3 and QUIC protocol support
- ✅ TLS 1.3 certificate requirements
- ✅ HTTP/3 specific configuration
- ✅ QUIC connection management
- ✅ HTTP/3 graceful shutdown

**Key Features Tested:**
- HTTP/3 protocol implementation
- QUIC transport layer
- TLS 1.3 requirements
- HTTP/3 specific timeouts
- Next-generation protocol features

### 5. Helper Functions Tests (`helper_test.go`)
- ✅ Configuration constants validation
- ✅ Default timeout values
- ✅ Configuration key consistency

## Test Scenarios

### Core Functionality
1. **Service Creation** - Factory functions with various configurations
2. **State Management** - Running status and active request counting
3. **Network Protocols** - TCP, Unix sockets, TLS, HTTP/3
4. **Graceful Shutdown** - Proper cleanup and request completion
5. **Error Handling** - Invalid configurations and runtime errors

### Advanced Features
1. **Concurrent Request Handling** - Multiple simultaneous requests
2. **Active Request Tracking** - Real-time request counting
3. **Timeout Management** - Read, write, and idle timeouts
4. **TLS/SSL Support** - Certificate management and client authentication
5. **Protocol Support** - HTTP/1.1, HTTP/2, HTTP/3

### Edge Cases
1. **Shutdown During Requests** - Graceful handling of active requests
2. **Timeout Scenarios** - Shutdown timeout and request timeouts
3. **Invalid Configurations** - Missing certificates, invalid timeouts
4. **Resource Cleanup** - Proper cleanup of servers and connections
5. **Error Conditions** - Network errors and certificate issues

## Configuration Testing

### Timeout Configuration
```go
config := map[string]any{
    "read_timeout":  "30s",
    "write_timeout": "45s", 
    "idle_timeout":  "60s",
}
```

### TLS Configuration
```go
config := map[string]any{
    "cert_file": "/path/to/cert.pem",
    "key_file":  "/path/to/key.pem",
    "ca_file":   "/path/to/ca.pem", // Optional for client auth
}
```

### Array Configuration
```go
config := []string{"/path/to/cert.pem", "/path/to/key.pem"}
```

## Running Tests

### Run All Listener Tests
```bash
go test ./modules/coreservice/listener/...
```

### Run Specific Listener Tests
```bash
go test ./modules/coreservice/listener/ -run TestNetHttpListener
go test ./modules/coreservice/listener/ -run TestFastHttpListener
go test ./modules/coreservice/listener/ -run TestSecureNetHttpListener
go test ./modules/coreservice/listener/ -run TestHttp3Listener
```

### Run with Verbose Output
```bash
go test -v ./modules/coreservice/listener/
```

### Run with Coverage
```bash
go test -cover ./modules/coreservice/listener/
```

### Skip HTTP/3 Tests (if dependencies missing)
```bash
go test -short ./modules/coreservice/listener/
```

## Test Architecture

### Mock Infrastructure
- **Test Certificates** - Auto-generated TLS certificates for testing
- **Test Servers** - Lightweight HTTP test servers
- **Concurrent Testing** - Goroutine-based concurrent request simulation
- **Resource Cleanup** - Proper cleanup of temporary files and servers

### Testing Patterns
- **Factory Function Testing** - Configuration validation and object creation
- **State Management Testing** - Running status and request counting
- **Lifecycle Testing** - Start, run, shutdown, cleanup
- **Error Injection Testing** - Invalid inputs and error conditions
- **Concurrency Testing** - Multiple goroutines and race condition detection

## Performance Considerations

### FastHTTP Performance
- Tests validate high-performance request handling
- Concurrent request processing validation
- Resource usage optimization testing

### HTTP/3 Performance  
- QUIC protocol efficiency testing
- TLS 1.3 performance validation
- Next-generation protocol benefits

### Graceful Shutdown Performance
- Active request completion timing
- Shutdown timeout effectiveness
- Resource cleanup efficiency

## Security Testing

### TLS/SSL Security
- Certificate validation and management
- Client certificate authentication (mTLS)
- TLS version and cipher suite validation

### Configuration Security
- Secure default configurations
- Certificate file permissions
- CA certificate chain validation

## Implementation Notes

### Test Certificate Generation
Tests automatically generate temporary TLS certificates for secure listener testing, including:
- RSA 2048-bit keys
- Self-signed certificates
- Localhost and 127.0.0.1 SANs
- Proper certificate cleanup

### Network Port Management
Tests use dynamic port allocation and localhost binding to avoid conflicts:
- httptest.NewServer() for automatic port allocation
- Localhost binding for security
- Unix socket testing for IPC scenarios

### Concurrent Testing Strategy
Tests use sync.WaitGroup and atomic operations for:
- Request counting validation
- Graceful shutdown testing  
- Race condition detection
- Active request tracking

## Maintenance

### Adding New Tests
1. Follow existing test patterns and naming conventions
2. Include configuration validation tests
3. Test both success and error scenarios
4. Add concurrency and shutdown testing
5. Update this README with new test descriptions

### Updating Configurations
1. Update configuration constants in tests
2. Validate new timeout and security settings
3. Test backward compatibility
4. Document configuration changes

This comprehensive test suite ensures the reliability, performance, and security of the Lokstra HTTP listener implementations across all supported protocols and configurations.
