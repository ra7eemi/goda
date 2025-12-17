/************************************************************************************
 *
 * goda (Golang Optimized Discord API), A Lightweight Go library for Discord API
 *
 * SPDX-License-Identifier: BSD-3-Clause
 *
 * Copyright 2025 Marouane Souiri
 *
 * Licensed under the BSD 3-Clause License.
 * See the LICENSE file for details.
 *
 ************************************************************************************/

package goda

import "errors"

// Common errors returned by the goda library.
var (
	// ErrNoClient is returned when an entity action method is called
	// but the entity has no client reference.
	ErrNoClient = errors.New("goda: entity has no client reference")

	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("goda: resource not found")

	// ErrUnauthorized is returned when the bot lacks permission for an action.
	ErrUnauthorized = errors.New("goda: unauthorized")

	// ErrRateLimited is returned when the API rate limit has been exceeded.
	ErrRateLimited = errors.New("goda: rate limited")

	// ErrInvalidToken is returned when the bot token is invalid.
	ErrInvalidToken = errors.New("goda: invalid token")

	// ErrMissingPermissions is returned when the bot lacks required permissions.
	ErrMissingPermissions = errors.New("goda: missing permissions")

	// ErrInvalidSnowflake is returned when a snowflake ID is invalid.
	ErrInvalidSnowflake = errors.New("goda: invalid snowflake")

	// ErrChannelNotText is returned when a text channel operation is attempted
	// on a non-text channel.
	ErrChannelNotText = errors.New("goda: channel is not a text channel")

	// ErrChannelNotVoice is returned when a voice channel operation is attempted
	// on a non-voice channel.
	ErrChannelNotVoice = errors.New("goda: channel is not a voice channel")

	// ErrChannelNotAnnouncement is returned when an announcement channel operation
	// is attempted on a non-announcement channel.
	ErrChannelNotAnnouncement = errors.New("goda: channel is not an announcement channel")

	// ErrChannelNotStage is returned when a stage channel operation is attempted
	// on a non-stage channel.
	ErrChannelNotStage = errors.New("goda: channel is not a stage channel")

	// ErrChannelNotThread is returned when a thread channel operation is attempted
	// on a non-thread channel.
	ErrChannelNotThread = errors.New("goda: channel is not a thread channel")

	// ErrChannelNotForum is returned when a forum channel operation is attempted
	// on a non-forum channel.
	ErrChannelNotForum = errors.New("goda: channel is not a forum channel")

	// ErrChannelNotMedia is returned when a media channel operation is attempted
	// on a non-media channel.
	ErrChannelNotMedia = errors.New("goda: channel is not a media channel")

	// ErrDMNotAllowed is returned when a DM cannot be sent to a user.
	ErrDMNotAllowed = errors.New("goda: cannot send DM to this user")
)

// DiscordAPIError represents an error returned by the Discord API.
type DiscordAPIError struct {
	// Code is the Discord error code.
	Code int `json:"code"`

	// Message is the error message from Discord.
	Message string `json:"message"`

	// HTTPStatus is the HTTP status code.
	HTTPStatus int `json:"-"`

	// Errors contains nested validation errors.
	Errors map[string]interface{} `json:"errors,omitempty"`
}

// Error implements the error interface.
func (e *DiscordAPIError) Error() string {
	return e.Message
}

// IsNotFound returns true if this is a 404 Not Found error.
func (e *DiscordAPIError) IsNotFound() bool {
	return e.HTTPStatus == 404
}

// IsRateLimited returns true if this is a 429 Rate Limited error.
func (e *DiscordAPIError) IsRateLimited() bool {
	return e.HTTPStatus == 429
}

// IsUnauthorized returns true if this is a 401 Unauthorized error.
func (e *DiscordAPIError) IsUnauthorized() bool {
	return e.HTTPStatus == 401
}

// IsForbidden returns true if this is a 403 Forbidden error.
func (e *DiscordAPIError) IsForbidden() bool {
	return e.HTTPStatus == 403
}
