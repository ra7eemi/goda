# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Run all tests
go test ./...

# Run a single test
go test -run TestFunctionName ./...

# Build/verify compilation
go build ./...

# Get dependencies
go get ./...
```

## Architecture Overview

GODA (Golang Optimized Discord API) is a Discord API wrapper with these core components:

**Client** (client.go) - Main entry point that manages configuration, shards, REST API, cache, and event dispatcher. Created via `goda.New()` with functional options pattern (`WithToken`, `WithIntents`, etc.).

**Shard** (shard.go) - Manages a single WebSocket connection to Discord Gateway. Handles connection lifecycle, heartbeats, session resumption, and reconnection with exponential backoff.

**Dispatcher** (dispatcher.go) - Routes Gateway events to registered handlers. Event handlers are registered via `OnXxx` methods (e.g., `OnMessageCreate`). Events are dispatched asynchronously through a worker pool.

**CacheManager** (cache.go) - Interface for caching Discord entities (guilds, users, channels, members, roles, voice states). `DefaultCache` provides a thread-safe in-memory implementation with configurable cache flags.

**RestAPI** (restapi.go) - HTTP client for Discord REST endpoints. Uses bytedance/sonic for JSON serialization.

## Key Patterns

- **Functional Options**: Client configuration uses `clientOption` functions
- **Interface-Based Extensibility**: `CacheManager`, `Logger`, `WorkerPool`, `ShardsIdentifyRateLimiter` can be swapped with custom implementations
- **Gateway Intents**: Bit flags control which events the bot receives (e.g., `GatewayIntentGuilds | GatewayIntentMessageContent`)
- **Snowflake IDs**: Discord IDs are typed as `Snowflake` for type safety

## Dependencies

- `github.com/bytedance/sonic` - Fast JSON serialization
- `github.com/gobwas/ws` - WebSocket client for Gateway connection
