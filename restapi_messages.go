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

import (
	"net/url"
	"strconv"

	"encoding/json"
)

/***********************
 *   Message Endpoints *
 ***********************/

// FetchMessagesOptions are options for fetching messages from a channel.
type FetchMessagesOptions struct {
	// Around gets messages around this message ID.
	Around Snowflake
	// Before gets messages before this message ID.
	Before Snowflake
	// After gets messages after this message ID.
	After Snowflake
	// Limit is the maximum number of messages to return (1-100). Default is 50.
	Limit int
}

// FetchMessages retrieves messages from a channel.
//
// Usage example:
//
//	messages, err := client.FetchMessages(channelID, FetchMessagesOptions{
//	    Limit: 10,
//	})
func (r *restApi) FetchMessages(channelID Snowflake, opts FetchMessagesOptions) ([]Message, error) {
	query := url.Values{}
	if opts.Limit > 0 {
		if opts.Limit > 100 {
			opts.Limit = 100
		}
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if !opts.Around.UnSet() {
		query.Set("around", opts.Around.String())
	}
	if !opts.Before.UnSet() {
		query.Set("before", opts.Before.String())
	}
	if !opts.After.UnSet() {
		query.Set("after", opts.After.String())
	}

	endpoint := "/channels/" + channelID.String() + "/messages"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	body, err := r.doRequest("GET", endpoint, nil, true, "")
	if err != nil {
		return nil, err
	}

	var messages []Message
	if err := json.Unmarshal(body, &messages); err != nil {
		r.logger.Error("Failed parsing response for GET /channels/{id}/messages: " + err.Error())
		return nil, err
	}
	return messages, nil
}

// FetchMessage retrieves a single message by ID from a channel.
//
// Usage example:
//
//	message, err := client.FetchMessage(channelID, messageID)
func (r *restApi) FetchMessage(channelID, messageID Snowflake) (Message, error) {
	body, err := r.doRequest("GET", "/channels/"+channelID.String()+"/messages/"+messageID.String(), nil, true, "")
	if err != nil {
		return Message{}, err
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		r.logger.Error("Failed parsing response for GET /channels/{id}/messages/{id}: " + err.Error())
		return Message{}, err
	}
	return message, nil
}

// MessageEditOptions are options for editing a message.
type MessageEditOptions struct {
	// Content is the new message content (up to 2000 characters).
	Content string `json:"content,omitempty"`
	// Embeds are the new embedded rich content (up to 10 embeds).
	Embeds []Embed `json:"embeds,omitempty"`
	// Flags are edit flags to set on the message.
	Flags MessageFlags `json:"flags,omitempty"`
	// AllowedMentions are the allowed mentions for the message.
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`
	// Components are the components to include with the message.
	Components []LayoutComponent `json:"components,omitempty"`
	// Attachments are the attachments to keep or add.
	Attachments []Attachment `json:"attachments,omitempty"`
}

// EditMessage edits a previously sent message.
//
// Usage example:
//
//	message, err := client.EditMessage(channelID, messageID, MessageEditOptions{
//	    Content: "Updated content",
//	})
func (r *restApi) EditMessage(channelID, messageID Snowflake, opts MessageEditOptions) (Message, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("PATCH", "/channels/"+channelID.String()+"/messages/"+messageID.String(), reqBody, true, "")
	if err != nil {
		return Message{}, err
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		r.logger.Error("Failed parsing response for PATCH /channels/{id}/messages/{id}: " + err.Error())
		return Message{}, err
	}
	return message, nil
}

// DeleteMessage deletes a message from a channel.
//
// Usage example:
//
//	err := client.DeleteMessage(channelID, messageID, "Spam")
func (r *restApi) DeleteMessage(channelID, messageID Snowflake, reason string) error {
	_, err := r.doRequest("DELETE", "/channels/"+channelID.String()+"/messages/"+messageID.String(), nil, true, reason)
	return err
}

// BulkDeleteMessages deletes multiple messages in a single request.
// This endpoint can only be used on messages that are less than 2 weeks old.
// Between 2 and 100 messages may be deleted at once.
//
// Usage example:
//
//	err := client.BulkDeleteMessages(channelID, messageIDs, "Cleanup")
func (r *restApi) BulkDeleteMessages(channelID Snowflake, messageIDs []Snowflake, reason string) error {
	reqBody, _ := json.Marshal(map[string][]Snowflake{"messages": messageIDs})
	_, err := r.doRequest("POST", "/channels/"+channelID.String()+"/messages/bulk-delete", reqBody, true, reason)
	return err
}

/***********************
 *  Reaction Endpoints *
 ***********************/

// CreateReaction adds a reaction to a message.
// The emoji must be URL encoded (e.g., %F0%9F%91%8D for thumbs up).
// For custom emoji, use the format name:id.
//
// Usage example:
//
//	err := client.CreateReaction(channelID, messageID, "👍")
//	err := client.CreateReaction(channelID, messageID, "custom_emoji:123456789")
func (r *restApi) CreateReaction(channelID, messageID Snowflake, emoji string) error {
	encodedEmoji := url.PathEscape(emoji)
	_, err := r.doRequest("PUT", "/channels/"+channelID.String()+"/messages/"+messageID.String()+"/reactions/"+encodedEmoji+"/@me", nil, true, "")
	return err
}

// DeleteOwnReaction removes the bot's own reaction from a message.
//
// Usage example:
//
//	err := client.DeleteOwnReaction(channelID, messageID, "👍")
func (r *restApi) DeleteOwnReaction(channelID, messageID Snowflake, emoji string) error {
	encodedEmoji := url.PathEscape(emoji)
	_, err := r.doRequest("DELETE", "/channels/"+channelID.String()+"/messages/"+messageID.String()+"/reactions/"+encodedEmoji+"/@me", nil, true, "")
	return err
}

// DeleteUserReaction removes another user's reaction from a message.
// Requires MANAGE_MESSAGES permission.
//
// Usage example:
//
//	err := client.DeleteUserReaction(channelID, messageID, userID, "👍")
func (r *restApi) DeleteUserReaction(channelID, messageID, userID Snowflake, emoji string) error {
	encodedEmoji := url.PathEscape(emoji)
	_, err := r.doRequest("DELETE", "/channels/"+channelID.String()+"/messages/"+messageID.String()+"/reactions/"+encodedEmoji+"/"+userID.String(), nil, true, "")
	return err
}

// GetReactionsOptions are options for getting reactions on a message.
type GetReactionsOptions struct {
	// After gets users after this user ID.
	After Snowflake
	// Limit is the maximum number of users to return (1-100). Default is 25.
	Limit int
}

// GetReactions gets a list of users that reacted with a specific emoji.
//
// Usage example:
//
//	users, err := client.GetReactions(channelID, messageID, "👍", GetReactionsOptions{Limit: 10})
func (r *restApi) GetReactions(channelID, messageID Snowflake, emoji string, opts GetReactionsOptions) ([]User, error) {
	encodedEmoji := url.PathEscape(emoji)
	query := url.Values{}
	if opts.Limit > 0 {
		if opts.Limit > 100 {
			opts.Limit = 100
		}
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if !opts.After.UnSet() {
		query.Set("after", opts.After.String())
	}

	endpoint := "/channels/" + channelID.String() + "/messages/" + messageID.String() + "/reactions/" + encodedEmoji
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	body, err := r.doRequest("GET", endpoint, nil, true, "")
	if err != nil {
		return nil, err
	}

	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		r.logger.Error("Failed parsing response for GET reactions: " + err.Error())
		return nil, err
	}
	return users, nil
}

// DeleteAllReactions removes all reactions from a message.
// Requires MANAGE_MESSAGES permission.
//
// Usage example:
//
//	err := client.DeleteAllReactions(channelID, messageID)
func (r *restApi) DeleteAllReactions(channelID, messageID Snowflake) error {
	_, err := r.doRequest("DELETE", "/channels/"+channelID.String()+"/messages/"+messageID.String()+"/reactions", nil, true, "")
	return err
}

// DeleteAllReactionsForEmoji removes all reactions for a specific emoji.
// Requires MANAGE_MESSAGES permission.
//
// Usage example:
//
//	err := client.DeleteAllReactionsForEmoji(channelID, messageID, "👍")
func (r *restApi) DeleteAllReactionsForEmoji(channelID, messageID Snowflake, emoji string) error {
	encodedEmoji := url.PathEscape(emoji)
	_, err := r.doRequest("DELETE", "/channels/"+channelID.String()+"/messages/"+messageID.String()+"/reactions/"+encodedEmoji, nil, true, "")
	return err
}

/***********************
 *    Pin Endpoints    *
 ***********************/

// PinMessage pins a message in a channel.
// Requires MANAGE_MESSAGES permission.
// Maximum of 50 pinned messages per channel.
//
// Usage example:
//
//	err := client.PinMessage(channelID, messageID, "Important message")
func (r *restApi) PinMessage(channelID, messageID Snowflake, reason string) error {
	_, err := r.doRequest("PUT", "/channels/"+channelID.String()+"/pins/"+messageID.String(), nil, true, reason)
	return err
}

// UnpinMessage unpins a message from a channel.
// Requires MANAGE_MESSAGES permission.
//
// Usage example:
//
//	err := client.UnpinMessage(channelID, messageID, "No longer important")
func (r *restApi) UnpinMessage(channelID, messageID Snowflake, reason string) error {
	_, err := r.doRequest("DELETE", "/channels/"+channelID.String()+"/pins/"+messageID.String(), nil, true, reason)
	return err
}

// GetPinnedMessages retrieves all pinned messages in a channel.
//
// Usage example:
//
//	messages, err := client.GetPinnedMessages(channelID)
func (r *restApi) GetPinnedMessages(channelID Snowflake) ([]Message, error) {
	body, err := r.doRequest("GET", "/channels/"+channelID.String()+"/pins", nil, true, "")
	if err != nil {
		return nil, err
	}

	var messages []Message
	if err := json.Unmarshal(body, &messages); err != nil {
		r.logger.Error("Failed parsing response for GET /channels/{id}/pins: " + err.Error())
		return nil, err
	}
	return messages, nil
}
