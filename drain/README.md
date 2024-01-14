# Drain Infinite Loop

Test showcasing `Drain()` method causing an infinite loop in `Next()` if it's waiting for messages.

## Test

```sh
go test -race -v
```

