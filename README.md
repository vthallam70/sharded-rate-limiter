# Sharded rate limiter

A Go HTTP service shards keys across a static cluster using rendezvous hashing. Each owner enforces a thread-safe token bucket; nodes proxy requests to the owner. An unavailable owner returns HTTP 503 (fail closed). It is not replicated: an owner restart loses its in-memory buckets, and changing membership remaps keys, so this is not a production-safe distributed quota service.

```sh
go test -race ./...
go test -bench=. ./...
go run ./cmd/server -listen :8080 -self http://127.0.0.1:8080 -nodes http://127.0.0.1:8080,http://127.0.0.1:8081
go run ./cmd/server -listen :8081 -self http://127.0.0.1:8081 -nodes http://127.0.0.1:8080,http://127.0.0.1:8081
curl -X POST http://127.0.0.1:8080/allow -d '{"key":"client-a"}'
```

All nodes require the same membership and a trusted network. There is no transport authentication, durable state, consensus, automatic recovery, or bucket eviction. The test includes an unavailable owner. The benchmark measures local operations, not network latency.
