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
	"encoding/json"
)

/***********************
 *  Channel Endpoints  *
 ***********************/

// ChannelEditOptions are options for editing a channel.
type ChannelEditOptions struct {
	// Name is the channel name (1-100 characters).
	Name string `json:"name,omitempty"`
	// Type is the type of channel (only conversion between text and announcement is supported).
	Type ChannelType `json:"type,omitempty"`
	// Position is the position of the channel in the left-hand listing.
	Position *int `json:"position,omitempty"`
	// Topic is the channel topic (0-1024 characters for text/announcement, 0-4096 for forum/media).
	Topic string `json:"topic,omitempty"`
	// NSFW indicates whether the channel is nsfw.
	NSFW *bool `json:"nsfw,omitempty"`
	// RateLimitPerUser is the slowmode rate limit in seconds (0-21600).
	RateLimitPerUser *int `json:"rate_limit_per_user,omitempty"`
	// Bitrate is the bitrate for voice channels (8000-96000 or up to 384000 for VIP servers).
	Bitrate *int `json:"bitrate,omitempty"`
	// UserLimit is the user limit for voice channels (0-99, 0 is unlimited).
	UserLimit *int `json:"user_limit,omitempty"`
	// PermissionOverwrites are the channel permission overwrites.
	PermissionOverwrites []PermissionOverwrite `json:"permission_overwrites,omitempty"`
	// ParentID is the id of the parent category for a channel.
	ParentID *Snowflake `json:"parent_id,omitempty"`
	// RTCRegion is the voice region id for the voice channel, automatic when set to nil.
	RTCRegion *string `json:"rtc_region,omitempty"`
	// VideoQualityMode is the camera video quality mode of the voice channel.
	VideoQualityMode *int `json:"video_quality_mode,omitempty"`
	// DefaultAutoArchiveDuration is the default duration (in minutes) for newly created threads.
	DefaultAutoArchiveDuration *int `json:"default_auto_archive_duration,omitempty"`
	// Flags are channel flags combined as a bitfield.
	Flags *ChannelFlags `json:"flags,omitempty"`
	// AvailableTags are tags that can be used in a forum or media channel (max 20).
	AvailableTags []ForumTag `json:"available_tags,omitempty"`
	// DefaultReactionEmoji is the emoji to show in the add reaction button on a thread.
	DefaultReactionEmoji *DefaultReactionEmoji `json:"default_reaction_emoji,omitempty"`
	// DefaultThreadRateLimitPerUser is the default slowmode for threads.
	DefaultThreadRateLimitPerUser *int `json:"default_thread_rate_limit_per_user,omitempty"`
	// DefaultSortOrder is the default sort order type for forum posts.
	DefaultSortOrder *int `json:"default_sort_order,omitempty"`
	// DefaultForumLayout is the default forum layout view for forum channels.
	DefaultForumLayout *int `json:"default_forum_layout,omitempty"`
}

// EditChannel modifies a channel's settings. Returns the updated channel.
// Requires MANAGE_CHANNELS permission.
//
// Usage example:
//
//	channel, err := client.EditChannel(channelID, ChannelEditOptions{
//	    Name: "new-channel-name",
//	    Topic: "Updated topic",
//	}, "Channel update")
func (r *restApi) EditChannel(channelID Snowflake, opts ChannelEditOptions, reason string) (Channel, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("PATCH", "/channels/"+channelID.String(), reqBody, true, reason)
	if err != nil {
		return nil, err
	}
	return UnmarshalChannel(body)
}

// DeleteChannel deletes a channel or closes a DM.
// Requires MANAGE_CHANNELS permission for guild channels.
// Deleting a category does not delete its child channels.
//
// Usage example:
//
//	err := client.DeleteChannel(channelID, "No longer needed")
func (r *restApi) DeleteChannel(channelID Snowflake, reason string) error {
	_, err := r.doRequest("DELETE", "/channels/"+channelID.String(), nil, true, reason)
	return err
}

// EditChannelPermissions edits permissions for a role or user in a channel.
// Requires MANAGE_ROLES permission.
//
// Usage example:
//
//	err := client.EditChannelPermissions(channelID, roleID, PermissionOverwrite{
//	    ID: roleID,
//	    Type: PermissionOverwriteTypeRole,
//	    Allow: PermissionSendMessages,
//	    Deny: 0,
//	}, "Allow sending messages")
func (r *restApi) EditChannelPermissions(channelID Snowflake, overwrite PermissionOverwrite, reason string) error {
	reqBody, _ := json.Marshal(overwrite)
	_, err := r.doRequest("PUT", "/channels/"+channelID.String()+"/permissions/"+overwrite.ID.String(), reqBody, true, reason)
	return err
}

// DeleteChannelPermission deletes a channel permission overwrite for a user or role.
// Requires MANAGE_ROLES permission.
//
// Usage example:
//
//	err := client.DeleteChannelPermission(channelID, roleID, "Removing permission override")
func (r *restApi) DeleteChannelPermission(channelID, overwriteID Snowflake, reason string) error {
	_, err := r.doRequest("DELETE", "/channels/"+channelID.String()+"/permissions/"+overwriteID.String(), nil, true, reason)
	return err
}

// GetChannelInvites retrieves a list of invites for a channel.
// Requires MANAGE_CHANNELS permission.
//
// Usage example:
//
//	invites, err := client.GetChannelInvites(channelID)
func (r *restApi) GetChannelInvites(channelID Snowflake) ([]Invite, error) {
	body, err := r.doRequest("GET", "/channels/"+channelID.String()+"/invites", nil, true, "")
	if err != nil {
		return nil, err
	}

	var invites []Invite
	if err := json.Unmarshal(body, &invites); err != nil {
		r.logger.Error("Failed parsing response for GET /channels/{id}/invites: " + err.Error())
		return nil, err
	}
	return invites, nil
}

// Invite represents a Discord invite.
type Invite struct {
	// Code is the invite code (unique ID).
	Code string `json:"code"`
	// Guild is a partial guild object the invite is for.
	Guild *PartialGuild `json:"guild,omitempty"`
	// Channel is a partial channel object the invite is for.
	Channel *PartialChannel `json:"channel,omitempty"`
	// Inviter is the user who created the invite.
	Inviter *User `json:"inviter,omitempty"`
	// TargetType is the type of target for the invite.
	TargetType int `json:"target_type,omitempty"`
	// TargetUser is the user whose stream to display for this voice channel invite.
	TargetUser *User `json:"target_user,omitempty"`
	// ApproximatePresenceCount is the approximate count of online members.
	ApproximatePresenceCount int `json:"approximate_presence_count,omitempty"`
	// ApproximateMemberCount is the approximate count of total members.
	ApproximateMemberCount int `json:"approximate_member_count,omitempty"`
	// ExpiresAt is the expiration date of this invite.
	ExpiresAt *string `json:"expires_at,omitempty"`
	// Uses is the number of times this invite has been used.
	Uses int `json:"uses,omitempty"`
	// MaxUses is the max number of times this invite can be used.
	MaxUses int `json:"max_uses,omitempty"`
	// MaxAge is the duration (in seconds) after which the invite expires.
	MaxAge int `json:"max_age,omitempty"`
	// Temporary indicates whether this invite only grants temporary membership.
	Temporary bool `json:"temporary,omitempty"`
	// CreatedAt is when this invite was created.
	CreatedAt string `json:"created_at,omitempty"`
}

// PartialChannel represents a partial channel object.
type PartialChannel struct {
	ID   Snowflake   `json:"id"`
	Name string      `json:"name"`
	Type ChannelType `json:"type"`
}

// CreateInviteOptions are options for creating an invite.
type CreateInviteOptions struct {
	// MaxAge is the duration of invite in seconds, 0 for never. Default 86400 (24 hours).
	MaxAge int `json:"max_age,omitempty"`
	// MaxUses is the max number of uses, 0 for unlimited. Default 0.
	MaxUses int `json:"max_uses,omitempty"`
	// Temporary indicates whether this invite grants temporary membership.
	Temporary bool `json:"temporary,omitempty"`
	// Unique indicates whether to try to reuse a similar invite (when false).
	Unique bool `json:"unique,omitempty"`
	// TargetType is the type of target for this voice channel invite.
	TargetType int `json:"target_type,omitempty"`
	// TargetUserID is the id of the user whose stream to display.
	TargetUserID Snowflake `json:"target_user_id,omitempty"`
	// TargetApplicationID is the id of the embedded application to open.
	TargetApplicationID Snowflake `json:"target_application_id,omitempty"`
}

// CreateChannelInvite creates a new invite for a channel.
// Requires CREATE_INSTANT_INVITE permission.
//
// Usage example:
//
//	invite, err := client.CreateChannelInvite(channelID, CreateInviteOptions{
//	    MaxAge: 3600,
//	    MaxUses: 10,
//	}, "Event invite")
func (r *restApi) CreateChannelInvite(channelID Snowflake, opts CreateInviteOptions, reason string) (Invite, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("POST", "/channels/"+channelID.String()+"/invites", reqBody, true, reason)
	if err != nil {
		return Invite{}, err
	}

	var invite Invite
	if err := json.Unmarshal(body, &invite); err != nil {
		r.logger.Error("Failed parsing response for POST /channels/{id}/invites: " + err.Error())
		return Invite{}, err
	}
	return invite, nil
}

// TriggerTypingIndicator triggers the typing indicator in a channel.
// Generally bots should not use this, but it's available if needed.
//
// Usage example:
//
//	err := client.TriggerTypingIndicator(channelID)
func (r *restApi) TriggerTypingIndicator(channelID Snowflake) error {
	_, err := r.doRequest("POST", "/channels/"+channelID.String()+"/typing", nil, true, "")
	return err
}
