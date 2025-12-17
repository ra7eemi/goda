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

// EntityBase provides common functionality for all Discord entities
// that need to interact with the Discord API.
//
// Entities embedding EntityBase can call action methods like Reply(), Delete(), etc.
// The client reference is set automatically when entities are received from events
// or fetched from the API.
type EntityBase struct {
	client *Client
}

// SetClient sets the client reference for this entity.
// This is called internally when entities are received from events or API responses.
func (e *EntityBase) SetClient(c *Client) {
	e.client = c
}

// Client returns the client reference for this entity.
// Returns nil if the entity was not received from a client context.
func (e *EntityBase) Client() *Client {
	return e.client
}

// HasClient returns true if this entity has a client reference.
func (e *EntityBase) HasClient() bool {
	return e.client != nil
}

// ImageFile represents a base64-encoded image for Discord API requests.
// Used for guild icons, splashes, banners, role icons, etc.
type ImageFile string

// ApplicationCommandOptionChoice represents a choice for a command option.
// This is a placeholder type - the full definition is in application_command.go.
type ApplicationCommandOptionChoice struct {
	Name              string      `json:"name"`
	NameLocalizations StringMap   `json:"name_localizations,omitempty"`
	Value             interface{} `json:"value"`
}

// StringMap is a map of locale codes to strings for localizations.
type StringMap map[string]string
