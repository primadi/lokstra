# Router Performance Comparison Chart

## Performance Overview (Lower is Better)

```
┌─────────────────────────────────────────────────────────────────┐
│ Static Routes (ns/op)                                           │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ████ 200.7ns                                      │
│ ServeMuxPlus  ███████ 346.6ns                                   │
│ ChiRouter     ███████ 350.7ns                                   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Path Parameters (ns/op)                                         │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      █████ 278.2ns                                     │
│ ServeMuxPlus  █████████ 468.0ns                                 │
│ ChiRouter     █████████████ 689.7ns                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Wildcard Routes (ns/op)                                         │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ████████████████ 783.7ns                          │
│ ServeMuxPlus  ████████████████ 814.2ns                          │
│ ChiRouter     █████████████ 655.9ns ⭐ FASTEST                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ OPTIONS Requests (ns/op)                                        │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ████████████████████████████████████████ 2,639ns  │
│ ServeMuxPlus  ███████████████████████████████████████████ 3,209 │
│ ChiRouter     ██████████████ 686.9ns ⭐ FASTEST                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Mixed Routes (ns/op)                                            │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ████████████ 593.6ns                              │
│ ServeMuxPlus  ███████████████ 770.2ns                           │
│ ChiRouter     █████████████████ 860.4ns                         │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Large Route Table (100 routes, ns/op)                          │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ██████ 302.9ns                                    │
│ ServeMuxPlus  ██████████ 487.3ns                                │
│ ChiRouter     ██████████████████ 882.4ns                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Parallel Requests (ns/op)                                       │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ██ 48.25ns                                        │
│ ServeMuxPlus  ███ 69.95ns                                       │
│ ChiRouter     ████████ 204.9ns                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Memory Allocations (Lower is Better)

```
┌─────────────────────────────────────────────────────────────────┐
│ Static Routes (bytes/op)                                        │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      █ 7 B                                             │
│ ServeMuxPlus  ████████ 134 B                                    │
│ ChiRouter     ████████████████████ 375 B                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Path Parameters (bytes/op)                                      │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ██ 32 B                                           │
│ ServeMuxPlus  █████████ 163 B                                   │
│ ChiRouter     ████████████████████████████████████████ 721 B    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Wildcard Routes (bytes/op)                                      │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ████████ 151 B                                    │
│ ServeMuxPlus  ███████████████ 283 B                             │
│ ChiRouter     ████████████████████████████████████████ 704 B    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Parallel Requests (bytes/op)                                    │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ██ 34 B                                           │
│ ServeMuxPlus  █████████ 160 B                                   │
│ ChiRouter     ████████████████████████████████████████ 718 B    │
└─────────────────────────────────────────────────────────────────┘
```

## Throughput Comparison (Higher is Better)

```
┌─────────────────────────────────────────────────────────────────┐
│ Static Routes (million ops/sec)                                │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ████████████████████████████ 5.0 M ops/s ⭐       │
│ ServeMuxPlus  ████████████████ 2.9 M ops/s                      │
│ ChiRouter     ███████████████ 2.8 M ops/s                       │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Path Parameters (million ops/sec)                              │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ████████████████████████ 3.6 M ops/s ⭐           │
│ ServeMuxPlus  ██████████████ 2.1 M ops/s                        │
│ ChiRouter     ████████ 1.2 M ops/s                              │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Parallel Requests (million ops/sec)                            │
├─────────────────────────────────────────────────────────────────┤
│ ServeMux      ████████████████████████████████ 20.7 M ops/s ⭐  │
│ ServeMuxPlus  ████████████████████ 14.3 M ops/s                 │
│ ChiRouter     ███████ 4.9 M ops/s                               │
└─────────────────────────────────────────────────────────────────┘
```

## Speed Multiplier vs ServeMux

```
┌──────────────────────────────────────────────────────────────┐
│                     ServeMux = 1.0x (baseline)               │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Static Routes:                                              │
│    ServeMux      ████████████████████████████████ 1.00x ⭐   │
│    ServeMuxPlus  █████████████████ 0.58x (1.7x slower)      │
│    ChiRouter     ████████████████ 0.57x (1.75x slower)      │
│                                                              │
│  Path Parameters:                                            │
│    ServeMux      ████████████████████████████████ 1.00x ⭐   │
│    ServeMuxPlus  ███████████████ 0.59x (1.68x slower)       │
│    ChiRouter     ████████ 0.40x (2.48x slower)              │
│                                                              │
│  Wildcard:                                                   │
│    ServeMux      ████████████████████████████████ 1.00x     │
│    ServeMuxPlus  █████████████████████████████ 0.96x        │
│    ChiRouter     █████████████████████████████████ 1.19x ⭐  │
│                                                              │
│  OPTIONS:                                                    │
│    ServeMux      ████████████████████████████████ 1.00x     │
│    ServeMuxPlus  ███████████████████████ 0.82x              │
│    ChiRouter     ████████████████████████████████████ 3.84x ⭐│
│                                                              │
│  Parallel:                                                   │
│    ServeMux      ████████████████████████████████ 1.00x ⭐   │
│    ServeMuxPlus  ████████████████████ 0.69x                 │
│    ChiRouter     ██████ 0.24x (4.25x slower)                │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## Winner Summary

```
┌─────────────────────┬─────────────────┬──────────────────┐
│ Category            │ Winner          │ Speed vs Others  │
├─────────────────────┼─────────────────┼──────────────────┤
│ Static Routes       │ ServeMux ⭐      │ 1.7x faster      │
│ Path Parameters     │ ServeMux ⭐      │ 1.68x faster     │
│ Wildcard Routes     │ ChiRouter ⭐     │ 1.19x faster     │
│ OPTIONS Requests    │ ChiRouter ⭐     │ 3.8x faster!     │
│ Mixed Routes        │ ServeMux ⭐      │ 1.3x faster      │
│ Large Route Table   │ ServeMux ⭐      │ 1.6x faster      │
│ Parallel Requests   │ ServeMux ⭐      │ 1.45x faster     │
│ Memory Efficiency   │ ServeMux ⭐      │ 2-22x less       │
└─────────────────────┴─────────────────┴──────────────────┘

Overall Winner: ServeMux (Go 1.22+ stdlib) 🏆
  - Wins 6 out of 8 categories
  - Best memory efficiency
  - Zero external dependencies

Special Mentions:
  - ChiRouter: Best for CORS-heavy APIs (OPTIONS) and SPA serving
  - ServeMuxPlus: Good middle ground with extras
```

## Real-World Impact

### High-Traffic API (1M req/sec)

**ServeMux**:
- Latency: ~200-300ns per request
- Memory: ~30-100 bytes per request
- Throughput: Can handle 3-5M ops/sec per core

**ChiRouter**:
- Latency: ~350-700ns per request
- Memory: ~375-720 bytes per request  
- Throughput: Can handle 1.2-2M ops/sec per core

**Conclusion**: ServeMux saves ~150-400ns per request
- At 1M req/sec → saves 150-400ms of CPU time per second
- At 10M req/sec → saves 1.5-4 seconds of CPU time per second!

### Memory Impact (1K concurrent requests)

**ServeMux**:
- Static: 7 KB (1K × 7 bytes)
- Params: 32 KB (1K × 32 bytes)

**ChiRouter**:
- Static: 375 KB (1K × 375 bytes) - 53x more!
- Params: 721 KB (1K × 721 bytes) - 22x more!

**Conclusion**: ServeMux is much more memory-efficient at scale.

---

## When to Choose What?

```
┌─────────────────────────────────────────────────────────────┐
│                   Decision Tree                             │
└─────────────────────────────────────────────────────────────┘

Do you need CORS preflight performance?
├─ YES → ChiRouter (3.8x faster OPTIONS)
└─ NO ↓

Do you need wildcard routes for SPA?
├─ YES → ChiRouter (1.19x faster wildcard)
└─ NO ↓

Do you want Chi's middleware ecosystem?
├─ YES → ChiRouter
└─ NO ↓

Do you need "ANY" method or better OPTIONS than stdlib?
├─ YES → ServeMuxPlus
└─ NO ↓

Do you want best performance?
└─ YES → ServeMux ⭐ (default choice)

```

---

**Generated**: October 7, 2025
**Benchmark Tool**: Go 1.22+ `go test -bench`
**CPU**: AMD Ryzen 9 5900HX (16 threads)
