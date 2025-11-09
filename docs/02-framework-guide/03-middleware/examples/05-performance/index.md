# Middleware Performance

# Middleware Performance Example

Compare performance impact of different middleware configurations.

## Running

```bash
go run main.go
```

Server starts on `http://localhost:3004`

## Benchmarking

Use ApacheBench to test performance:

```bash
ab -n 1000 -c 10 http://localhost:3004/baseline
ab -n 1000 -c 10 http://localhost:3004/light
ab -n 1000 -c 10 http://localhost:3004/heavy
```

## Endpoints

- `/baseline` - No middleware
- `/light` - 1 lightweight middleware
- `/light-5` - 5 lightweight middleware
- `/heavy` - 1 heavy middleware (10ms delay)

## Key Insights

- Middleware overhead is minimal for simple operations
- Multiple lightweight middleware have negligible impact
- Heavy processing in middleware directly affects response time
- Keep middleware focused and efficient
