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
 *   Guild Endpoints   *
 ***********************/

// FetchGuild retrieves a guild by its ID.
//
// Usage example:
//
//	guild, err := client.FetchGuild(guildID)
func (r *restApi) FetchGuild(guildID Snowflake) (Guild, error) {
	body, err := r.doRequest("GET", "/guilds/"+guildID.String()+"?with_counts=true", nil, true, "")
	if err != nil {
		return Guild{}, err
	}

	var guild Guild
	if err := json.Unmarshal(body, &guild); err != nil {
		r.logger.Error("Failed parsing response for GET /guilds/{id}: " + err.Error())
		return Guild{}, err
	}
	return guild, nil
}

// GuildEditOptions are options for editing a guild.
type GuildEditOptions struct {
	// Name is the guild name.
	Name string `json:"name,omitempty"`
	// VerificationLevel is the verification level required for the guild.
	VerificationLevel *VerificationLevel `json:"verification_level,omitempty"`
	// DefaultMessageNotifications is the default message notification level.
	DefaultMessageNotifications *MessageNotificationsLevel `json:"default_message_notifications,omitempty"`
	// ExplicitContentFilter is the explicit content filter level.
	ExplicitContentFilter *ExplicitContentFilterLevel `json:"explicit_content_filter,omitempty"`
	// AFKChannelID is the id of afk channel.
	AFKChannelID *Snowflake `json:"afk_channel_id,omitempty"`
	// AFKTimeout is the afk timeout in seconds.
	AFKTimeout *int `json:"afk_timeout,omitempty"`
	// Icon is the base64 1024x1024 png/jpeg/gif image for the guild icon.
	Icon *ImageFile `json:"icon,omitempty"`
	// OwnerID is the user id to transfer guild ownership to (must be owner).
	OwnerID *Snowflake `json:"owner_id,omitempty"`
	// Splash is the base64 16:9 png/jpeg image for the guild splash.
	Splash *ImageFile `json:"splash,omitempty"`
	// DiscoverySplash is the base64 16:9 png/jpeg image for the discovery splash.
	DiscoverySplash *ImageFile `json:"discovery_splash,omitempty"`
	// Banner is the base64 16:9 png/jpeg image for the guild banner.
	Banner *ImageFile `json:"banner,omitempty"`
	// SystemChannelID is the id of the channel where system messages are sent.
	SystemChannelID *Snowflake `json:"system_channel_id,omitempty"`
	// SystemChannelFlags are system channel flags.
	SystemChannelFlags *int `json:"system_channel_flags,omitempty"`
	// RulesChannelID is the id of the channel where Community guilds display rules.
	RulesChannelID *Snowflake `json:"rules_channel_id,omitempty"`
	// PublicUpdatesChannelID is the id of the channel where public updates are sent.
	PublicUpdatesChannelID *Snowflake `json:"public_updates_channel_id,omitempty"`
	// PreferredLocale is the preferred locale of a Community guild.
	PreferredLocale string `json:"preferred_locale,omitempty"`
	// Features are the enabled guild features.
	Features []string `json:"features,omitempty"`
	// Description is the description for the guild (Community only).
	Description *string `json:"description,omitempty"`
	// PremiumProgressBarEnabled indicates whether the boost progress bar is enabled.
	PremiumProgressBarEnabled *bool `json:"premium_progress_bar_enabled,omitempty"`
	// SafetyAlertsChannelID is the id of the channel where safety alerts are sent.
	SafetyAlertsChannelID *Snowflake `json:"safety_alerts_channel_id,omitempty"`
}

// EditGuild modifies a guild's settings. Returns the updated guild object.
// Requires MANAGE_GUILD permission.
//
// Usage example:
//
//	guild, err := client.EditGuild(guildID, GuildEditOptions{
//	    Name: "New Server Name",
//	}, "Renaming server")
func (r *restApi) EditGuild(guildID Snowflake, opts GuildEditOptions, reason string) (Guild, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("PATCH", "/guilds/"+guildID.String(), reqBody, true, reason)
	if err != nil {
		return Guild{}, err
	}

	var guild Guild
	if err := json.Unmarshal(body, &guild); err != nil {
		r.logger.Error("Failed parsing response for PATCH /guilds/{id}: " + err.Error())
		return Guild{}, err
	}
	return guild, nil
}

// LeaveGuild makes the bot leave a guild.
//
// Usage example:
//
//	err := client.LeaveGuild(guildID)
func (r *restApi) LeaveGuild(guildID Snowflake) error {
	_, err := r.doRequest("DELETE", "/users/@me/guilds/"+guildID.String(), nil, true, "")
	return err
}

// CreateGuildChannel creates a new channel in a guild. Returns the created channel.
// Requires MANAGE_CHANNELS permission.
//
// Usage example:
//
//	channel, err := client.CreateGuildChannel(guildID, ChannelCreateOptions{
//	    Name: "new-channel",
//	    Type: ChannelTypeGuildText,
//	}, "Creating new channel")
func (r *restApi) CreateGuildChannel(guildID Snowflake, opts ChannelCreateOptions, reason string) (Channel, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("POST", "/guilds/"+guildID.String()+"/channels", reqBody, true, reason)
	if err != nil {
		return nil, err
	}
	return UnmarshalChannel(body)
}

// GetGuildChannels retrieves all channels in a guild.
//
// Usage example:
//
//	channels, err := client.GetGuildChannels(guildID)
func (r *restApi) GetGuildChannels(guildID Snowflake) ([]Channel, error) {
	body, err := r.doRequest("GET", "/guilds/"+guildID.String()+"/channels", nil, true, "")
	if err != nil {
		return nil, err
	}

	var rawChannels []json.RawMessage
	if err := json.Unmarshal(body, &rawChannels); err != nil {
		r.logger.Error("Failed parsing response for GET /guilds/{id}/channels: " + err.Error())
		return nil, err
	}

	channels := make([]Channel, 0, len(rawChannels))
	for _, raw := range rawChannels {
		ch, err := UnmarshalChannel(raw)
		if err != nil {
			continue // Skip unknown channel types
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

// ModifyChannelPositionsEntry represents a channel position modification.
type ModifyChannelPositionsEntry struct {
	// ID is the channel id.
	ID Snowflake `json:"id"`
	// Position is the sorting position of the channel.
	Position *int `json:"position,omitempty"`
	// LockPermissions syncs the permission overwrites with the parent category.
	LockPermissions *bool `json:"lock_permissions,omitempty"`
	// ParentID is the new parent ID for the channel.
	ParentID *Snowflake `json:"parent_id,omitempty"`
}

// ModifyGuildChannelPositions modifies the positions of guild channels.
// Requires MANAGE_CHANNELS permission.
//
// Usage example:
//
//	err := client.ModifyGuildChannelPositions(guildID, []ModifyChannelPositionsEntry{
//	    {ID: channelID1, Position: intPtr(0)},
//	    {ID: channelID2, Position: intPtr(1)},
//	})
func (r *restApi) ModifyGuildChannelPositions(guildID Snowflake, positions []ModifyChannelPositionsEntry) error {
	reqBody, _ := json.Marshal(positions)
	_, err := r.doRequest("PATCH", "/guilds/"+guildID.String()+"/channels", reqBody, true, "")
	return err
}

// GetGuildPreview retrieves a guild preview by its ID.
// This is available for all guilds that the bot has MANAGE_GUILD in
// or guilds that are discoverable.
//
// Usage example:
//
//	preview, err := client.GetGuildPreview(guildID)
func (r *restApi) GetGuildPreview(guildID Snowflake) (GuildPreview, error) {
	body, err := r.doRequest("GET", "/guilds/"+guildID.String()+"/preview", nil, true, "")
	if err != nil {
		return GuildPreview{}, err
	}

	var preview GuildPreview
	if err := json.Unmarshal(body, &preview); err != nil {
		r.logger.Error("Failed parsing response for GET /guilds/{id}/preview: " + err.Error())
		return GuildPreview{}, err
	}
	return preview, nil
}

// GuildPreview represents a preview of a guild.
type GuildPreview struct {
	ID                       Snowflake `json:"id"`
	Name                     string    `json:"name"`
	Icon                     string    `json:"icon"`
	Splash                   string    `json:"splash"`
	DiscoverySplash          string    `json:"discovery_splash"`
	Emojis                   []Emoji   `json:"emojis"`
	Features                 []string  `json:"features"`
	ApproximateMemberCount   int       `json:"approximate_member_count"`
	ApproximatePresenceCount int       `json:"approximate_presence_count"`
	Description              string    `json:"description"`
	Stickers                 []Sticker `json:"stickers"`
}
