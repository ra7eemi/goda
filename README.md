# Goda - Golang Optimized Discord API
> Goda 5alina men lmar9a kat3ref tsaybe bot b goda?

## Overview

Goda is a lightweight and modern Discord API wrapper written in Go, designed for developers building Discord bots or integrations. With a focus on simplicity and performance, Goda provides an intuitive interface to interact with Discord's API, leveraging Go's concurrency features for efficient bot development. 

## Installation

To use Goda in your Go project, install it via:

```bash
go get github.com/ra7eemi/goda
```

Ensure you have Go 1.22 or later installed, as specified in the project’s `go.mod`.

## Usage

Here’s a basic Ping Pong example to get started with Goda:

```go
package main

import (
    "github.com/ra7eemi/goda"
    "context"
    "fmt"
)

func main() {
    // Initialize a new Goda client
    client := goda.New(
        context.TODO(),
		goda.WithToken("YOUR_BOT_TOKEN"),
		goda.WithIntents(goda.GatewayIntentGuildMessages, goda.GatewayIntentMessageContent),
    )

    // Add message create even handlers
    client.OnMessageCreate(func(event goda.MessageCreateEvent) {
        if event.Message.Content == "!ping" {
            fmt.Println("Pong!")
        }
    })

    // Start the bot
    client.Start()
}
```

Replace `YOUR_BOT_TOKEN` with your Discord bot token. Check the [documentation](https://pkg.go.dev/github.com/ra7eemi/goda) for more examples and API details.

## Badges

[![Go Reference](https://pkg.go.dev/badge/github.com/ra7eemi/goda.svg)](https://pkg.go.dev/github.com/ra7eemi/goda)
[![Go Report](https://goreportcard.com/badge/github.com/ra7eemi/goda)](https://goreportcard.com/report/github.com/ra7eemi/goda)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ra7eemi/goda)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://github.com/ra7eemi/goda/blob/master/LICENSE)
[![Yada Version](https://img.shields.io/github/v/tag/ra7eemi/goda?label=release)](https://github.com/ra7eemi/goda/releases/latest)
[![Issues](https://img.shields.io/github/issues/ra7eemi/goda)](https://github.com/ra7eemi/goda/issues)
[![Last Commit](https://img.shields.io/github/last-commit/ra7eemi/goda)](https://github.com/ra7eemi/goda/commits/main)
[![Lines of Code](https://tokei.rs/b1/github/ra7eemi/goda)](https://github.com/ra7eemi/goda)
