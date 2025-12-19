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
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

/*******************************************************************************
 *                              REST API CORE
 *******************************************************************************/

// restApi provides methods for Discord REST API endpoints.
type restApi struct {
	req    *requester
	logger Logger
}

// newRestApi creates a new RestAPI instance with optional custom requester and logger.
func newRestApi(req *requester, logger Logger) *restApi {
	return &restApi{
		req:    req,
		logger: logger,
	}
}

// Shutdown gracefully shuts down the REST API client.
func (r *restApi) Shutdown() {
	r.logger.Info("RestAPI shutting down")
	r.req.Shutdown()
	r.logger = nil
	r.req = nil
}

func (r *restApi) doRequest(method, endpoint string, body []byte, authWithToken bool, reason string) ([]byte, error) {
	r.logger.Debug("Calling endpoint: " + method + endpoint)

	res, err := r.req.do(method, endpoint, body, authWithToken, reason)
	if err != nil {
		r.logger.Error("Request failed for endpoint " + method + endpoint + ": " + err.Error())
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		r.logger.Error("Request failed for endpoint " + method + endpoint + ": Invalid Token")
		return nil, errors.New("invalid token")
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		r.logger.Error("Failed reading response body for endpoint " + method + endpoint + ": " + err.Error())
		return nil, err
	}

	r.logger.Debug("Successfully called endpoint: " + method + endpoint)
	return bodyBytes, nil
}

/*******************************************************************************
 *                              GATEWAY METHODS
 *******************************************************************************/

// FetchGatewayBot retrieves bot gateway information including recommended shard count and session limits.
//
// Usage example:
//
//	gateway, err := api.FetchGatewayBot()
//	if err != nil {
//	    // handle error
//	}
//	fmt.Println("Recommended shards:", gateway.Shards)
//
// Returns:
//   - GatewayBot: the bot gateway information.
//   - error: if the request failed or decoding failed.
func (r *restApi) FetchGatewayBot() (GatewayBot, error) {
	body, err := r.doRequest("GET", "/gateway/bot", nil, true, "")
	if err != nil {
		return GatewayBot{}, err
	}

	var obj GatewayBot
	if err := json.Unmarshal(body, &obj); err != nil {
		r.logger.Error("Failed parsing response for /gateway/bot: " + err.Error())
		return GatewayBot{}, err
	}
	return obj, nil
}

/*******************************************************************************
 *                              USER METHODS
 *******************************************************************************/

// FetchSelfUser retrieves the current bot user's data including username, ID, avatar, and flags.
//
// Usage example:
//
//	user, err := api.FetchSelfUser()
//	if err != nil {
//	    // handle error
//	}
//	fmt.Println("Bot username:", user.Username)
//
// Returns:
//   - User: the current user data.
//   - error: if the request failed or decoding failed.
func (r *restApi) FetchSelfUser() (User, error) {
	body, err := r.doRequest("GET", "/users/@me", nil, true, "")
	if err != nil {
		return User{}, err
	}

	var obj User
	if err := json.Unmarshal(body, &obj); err != nil {
		r.logger.Error("Failed parsing response for /users/@me: " + err.Error())
		return User{}, err
	}
	return obj, nil
}

// UpdateSelfUser updates the current bot user's username, avatar, or banner.
//
// Usage example:
//
//	newAvatar, _ := goda.NewImageFile("path/to/avatar.png")
//	err := api.UpdateSelfUser(UpdateSelfUserOptions{
//	    Username: "new_username",
//	    Avatar:   newAvatar,
//	})
//	if err != nil {
//	    // handle error
//	}
//	fmt.Println("User updated successfully")
//
// Returns:
//   - error: if the request failed.
func (r *restApi) UpdateSelfUser(opts UpdateSelfUserOptions) error {
	body, _ := json.Marshal(opts)
	_, err := r.doRequest("PATCH", "/users/@me", body, true, "")
	return err
}

// FetchUser retrieves a user by their Snowflake ID including username, avatar, and flags.
//
// Usage example:
//
//	user, err := api.FetchUser(123456789012345678)
//	if err != nil {
//	    // handle error
//	}
//	fmt.Println("Username:", user.Username)
//
// Returns:
//   - User: the user data.
//   - error: if the request failed or decoding failed.
func (r *restApi) FetchUser(userID Snowflake) (User, error) {
	body, err := r.doRequest("GET", "/users/"+userID.String(), nil, true, "")
	if err != nil {
		return User{}, err
	}

	var obj User
	if err := json.Unmarshal(body, &obj); err != nil {
		r.logger.Error("Failed parsing response for /users/{id}: " + err.Error())
		return User{}, err
	}
	return obj, nil
}

// CreateDM creates a DM channel with a user.
// Returns a DMChannel object.
//
// Usage example:
//
//	dm, err := client.CreateDM(userID)
//	if err == nil {
//	    client.SendMessage(dm.ID, MessageCreateOptions{Content: "Hello!"})
//	}
func (r *restApi) CreateDM(recipientID Snowflake) (DMChannel, error) {
	reqBody, _ := json.Marshal(map[string]Snowflake{"recipient_id": recipientID})
	body, err := r.doRequest("POST", "/users/@me/channels", reqBody, true, "")
	if err != nil {
		return DMChannel{}, err
	}

	var dm DMChannel
	if err := json.Unmarshal(body, &dm); err != nil {
		r.logger.Error("Failed parsing response for POST /users/@me/channels: " + err.Error())
		return DMChannel{}, err
	}
	return dm, nil
}

// GetCurrentUserGuilds retrieves a list of partial guild objects the current user is a member of.
// Requires the guilds OAuth2 scope.
//
// Usage example:
//
//	guilds, err := client.GetCurrentUserGuilds(GetCurrentUserGuildsOptions{Limit: 100})
func (r *restApi) GetCurrentUserGuilds(opts GetCurrentUserGuildsOptions) ([]PartialGuild, error) {
	endpoint := "/users/@me/guilds"
	query := opts.toQuery()
	if query != "" {
		endpoint += "?" + query
	}

	body, err := r.doRequest("GET", endpoint, nil, true, "")
	if err != nil {
		return nil, err
	}

	var guilds []PartialGuild
	if err := json.Unmarshal(body, &guilds); err != nil {
		r.logger.Error("Failed parsing response for GET /users/@me/guilds: " + err.Error())
		return nil, err
	}
	return guilds, nil
}

// GetCurrentUserGuildsOptions are options for getting current user guilds.
type GetCurrentUserGuildsOptions struct {
	// Before gets guilds before this guild ID.
	Before Snowflake
	// After gets guilds after this guild ID.
	After Snowflake
	// Limit is the max number of guilds to return (1-200). Default is 200.
	Limit int
	// WithCounts includes approximate member and presence counts.
	WithCounts bool
}

func (o GetCurrentUserGuildsOptions) toQuery() string {
	params := make([]string, 0)
	if !o.Before.UnSet() {
		params = append(params, "before="+o.Before.String())
	}
	if !o.After.UnSet() {
		params = append(params, "after="+o.After.String())
	}
	if o.Limit > 0 {
		if o.Limit > 200 {
			o.Limit = 200
		}
		params = append(params, "limit="+string(rune(o.Limit)))
	}
	if o.WithCounts {
		params = append(params, "with_counts=true")
	}
	if len(params) == 0 {
		return ""
	}
	result := params[0]
	for i := 1; i < len(params); i++ {
		result += "&" + params[i]
	}
	return result
}

// GetCurrentUserGuildMember retrieves the current user's member object for a guild.
//
// Usage example:
//
//	member, err := client.GetCurrentUserGuildMember(guildID)
func (r *restApi) GetCurrentUserGuildMember(guildID Snowflake) (Member, error) {
	body, err := r.doRequest("GET", "/users/@me/guilds/"+guildID.String()+"/member", nil, true, "")
	if err != nil {
		return Member{}, err
	}

	var member Member
	if err := json.Unmarshal(body, &member); err != nil {
		r.logger.Error("Failed parsing response for GET /users/@me/guilds/{id}/member: " + err.Error())
		return Member{}, err
	}
	member.GuildID = guildID
	return member, nil
}

// GetUserConnections retrieves the current user's connections.
// Requires the connections OAuth2 scope.
//
// Usage example:
//
//	connections, err := client.GetUserConnections()
func (r *restApi) GetUserConnections() ([]Connection, error) {
	body, err := r.doRequest("GET", "/users/@me/connections", nil, true, "")
	if err != nil {
		return nil, err
	}

	var connections []Connection
	if err := json.Unmarshal(body, &connections); err != nil {
		r.logger.Error("Failed parsing response for GET /users/@me/connections: " + err.Error())
		return nil, err
	}
	return connections, nil
}

// Connection represents a user's connected account.
type Connection struct {
	// ID is the id of the connection account.
	ID string `json:"id"`
	// Name is the username of the connection account.
	Name string `json:"name"`
	// Type is the service of the connection (twitch, youtube, etc.).
	Type string `json:"type"`
	// Revoked indicates whether the connection is revoked.
	Revoked bool `json:"revoked"`
	// Integrations is an array of partial server integrations.
	Integrations []Integration `json:"integrations"`
	// Verified indicates whether the connection is verified.
	Verified bool `json:"verified"`
	// FriendSync indicates whether friend sync is enabled.
	FriendSync bool `json:"friend_sync"`
	// ShowActivity indicates whether activities related to this connection are shown.
	ShowActivity bool `json:"show_activity"`
	// TwoWayLink indicates whether this connection has a corresponding third party OAuth2 token.
	TwoWayLink bool `json:"two_way_link"`
	// Visibility is the visibility of this connection.
	Visibility int `json:"visibility"`
}

// Integration represents a guild integration.
type Integration struct {
	ID                Snowflake    `json:"id"`
	Name              string       `json:"name"`
	Type              string       `json:"type"`
	Enabled           bool         `json:"enabled"`
	Syncing           bool         `json:"syncing"`
	RoleID            Snowflake    `json:"role_id"`
	EnableEmoticons   bool         `json:"enable_emoticons"`
	ExpireBehavior    int          `json:"expire_behavior"`
	ExpireGracePeriod int          `json:"expire_grace_period"`
	User              *User        `json:"user"`
	Account           Account      `json:"account"`
	SyncedAt          string       `json:"synced_at"`
	SubscriberCount   int          `json:"subscriber_count"`
	Revoked           bool         `json:"revoked"`
	Application       *Application `json:"application"`
}

// Account represents an integration account.
type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

/*******************************************************************************
 *                              GUILD METHODS
 *******************************************************************************/

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

/*******************************************************************************
 *                              CHANNEL METHODS
 *******************************************************************************/

// FetchChannel retrieves a channel by its Snowflake ID and decodes it into its concrete type
// (e.g. TextChannel, VoiceChannel, CategoryChannel).
//
// Usage example:
//
//	channel, err := api.FetchChannel(123456789012345678)
//	if err != nil {
//	    // handle error
//	}
//	fmt.Println("Channel ID:", channel.GetID())
//
// Returns:
//   - Channel: the decoded channel object.
//   - error: if the request failed or the type is unknown or decoding failed.
func (r *restApi) FetchChannel(channelID Snowflake) (Channel, error) {
	body, err := r.doRequest("GET", "/channels/"+channelID.String(), nil, true, "")
	if err != nil {
		return nil, err
	}
	return UnmarshalChannel(body)
}

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

/*******************************************************************************
 *                              MESSAGE METHODS
 *******************************************************************************/

// SendMessage send's a to the spesified channel.
//
// Usage example:
//
//	message, err := .SendMessage(123456789012345678, MessageCreateOptions{
//	       Content: "Hello, World!",
//	})
//	if err != nil {
//	    // handle error
//	}
//	fmt.Println("Message ID:", message.ID)
//
// Returns:
//   - Message: the message object.
//   - error: if the request or decoding failed.
func (r *restApi) SendMessage(channelID Snowflake, opts MessageCreateOptions) (Message, error) {
	reqBody, err := json.Marshal(opts)
	body, err := r.doRequest("POST", "/channels/"+channelID.String()+"/messages", reqBody, true, "")

	var message Message

	if err != nil {
		return message, err
	}

	err = json.Unmarshal(body, &message)
	if err != nil {
		return message, err
	}
	return message, nil
}

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

/*******************************************************************************
 *                              REACTION METHODS
 *******************************************************************************/

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

/*******************************************************************************
 *                              PIN METHODS
 *******************************************************************************/

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

/*******************************************************************************
 *                              MEMBER METHODS
 *******************************************************************************/

// FetchMember retrieves a guild member by their user ID.
//
// Usage example:
//
//	member, err := client.FetchMember(guildID, userID)
func (r *restApi) FetchMember(guildID, userID Snowflake) (Member, error) {
	body, err := r.doRequest("GET", "/guilds/"+guildID.String()+"/members/"+userID.String(), nil, true, "")
	if err != nil {
		return Member{}, err
	}

	var member Member
	if err := json.Unmarshal(body, &member); err != nil {
		r.logger.Error("Failed parsing response for GET /guilds/{id}/members/{id}: " + err.Error())
		return Member{}, err
	}
	member.GuildID = guildID
	return member, nil
}

// ListMembersOptions are options for listing guild members.
type ListMembersOptions struct {
	// Limit is the max number of members to return (1-1000). Default is 1.
	Limit int
	// After is the highest user id in the previous page.
	After Snowflake
}

// ListMembers retrieves a list of guild members.
// Requires GUILD_MEMBERS privileged intent.
//
// Usage example:
//
//	members, err := client.ListMembers(guildID, ListMembersOptions{Limit: 100})
func (r *restApi) ListMembers(guildID Snowflake, opts ListMembersOptions) ([]Member, error) {
	query := url.Values{}
	if opts.Limit > 0 {
		if opts.Limit > 1000 {
			opts.Limit = 1000
		}
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if !opts.After.UnSet() {
		query.Set("after", opts.After.String())
	}

	endpoint := "/guilds/" + guildID.String() + "/members"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	body, err := r.doRequest("GET", endpoint, nil, true, "")
	if err != nil {
		return nil, err
	}

	var members []Member
	if err := json.Unmarshal(body, &members); err != nil {
		r.logger.Error("Failed parsing response for GET /guilds/{id}/members: " + err.Error())
		return nil, err
	}

	// Set guild ID on all members
	for i := range members {
		members[i].GuildID = guildID
	}
	return members, nil
}

// SearchMembers searches for guild members whose username or nickname starts with the query.
// Returns a max of 1000 members.
//
// Usage example:
//
//	members, err := client.SearchMembers(guildID, "john", 10)
func (r *restApi) SearchMembers(guildID Snowflake, query string, limit int) ([]Member, error) {
	params := url.Values{}
	params.Set("query", query)
	if limit > 0 {
		if limit > 1000 {
			limit = 1000
		}
		params.Set("limit", strconv.Itoa(limit))
	}

	body, err := r.doRequest("GET", "/guilds/"+guildID.String()+"/members/search?"+params.Encode(), nil, true, "")
	if err != nil {
		return nil, err
	}

	var members []Member
	if err := json.Unmarshal(body, &members); err != nil {
		r.logger.Error("Failed parsing response for GET /guilds/{id}/members/search: " + err.Error())
		return nil, err
	}

	// Set guild ID on all members
	for i := range members {
		members[i].GuildID = guildID
	}
	return members, nil
}

// MemberEditOptions are options for editing a guild member.
type MemberEditOptions struct {
	// Nick is the value to set the user's nickname to. Requires MANAGE_NICKNAMES permission.
	Nick *string `json:"nick,omitempty"`
	// Roles is an array of role ids the member is assigned. Requires MANAGE_ROLES permission.
	Roles []Snowflake `json:"roles,omitempty"`
	// Mute indicates whether the user is muted in voice channels. Requires MUTE_MEMBERS permission.
	Mute *bool `json:"mute,omitempty"`
	// Deaf indicates whether the user is deafened in voice channels. Requires DEAFEN_MEMBERS permission.
	Deaf *bool `json:"deaf,omitempty"`
	// ChannelID is the id of channel to move user to (if they are in voice). Requires MOVE_MEMBERS permission.
	ChannelID *Snowflake `json:"channel_id,omitempty"`
	// CommunicationDisabledUntil is when the user's timeout will expire (up to 28 days). Requires MODERATE_MEMBERS permission.
	CommunicationDisabledUntil *time.Time `json:"communication_disabled_until,omitempty"`
	// Flags are guild member flags.
	Flags *MemberFlags `json:"flags,omitempty"`
}

// EditMember modifies attributes of a guild member.
// Returns the updated member object.
//
// Usage example:
//
//	nick := "New Nickname"
//	member, err := client.EditMember(guildID, userID, MemberEditOptions{
//	    Nick: &nick,
//	}, "Nickname change")
func (r *restApi) EditMember(guildID, userID Snowflake, opts MemberEditOptions, reason string) (Member, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("PATCH", "/guilds/"+guildID.String()+"/members/"+userID.String(), reqBody, true, reason)
	if err != nil {
		return Member{}, err
	}

	var member Member
	if err := json.Unmarshal(body, &member); err != nil {
		r.logger.Error("Failed parsing response for PATCH /guilds/{id}/members/{id}: " + err.Error())
		return Member{}, err
	}
	member.GuildID = guildID
	return member, nil
}

// KickMember removes a member from a guild.
// Requires KICK_MEMBERS permission.
//
// Usage example:
//
//	err := client.KickMember(guildID, userID, "Rule violation")
func (r *restApi) KickMember(guildID, userID Snowflake, reason string) error {
	_, err := r.doRequest("DELETE", "/guilds/"+guildID.String()+"/members/"+userID.String(), nil, true, reason)
	return err
}

// AddMemberRole adds a role to a guild member.
// Requires MANAGE_ROLES permission.
//
// Usage example:
//
//	err := client.AddMemberRole(guildID, userID, roleID, "Assigning role")
func (r *restApi) AddMemberRole(guildID, userID, roleID Snowflake, reason string) error {
	_, err := r.doRequest("PUT", "/guilds/"+guildID.String()+"/members/"+userID.String()+"/roles/"+roleID.String(), nil, true, reason)
	return err
}

// RemoveMemberRole removes a role from a guild member.
// Requires MANAGE_ROLES permission.
//
// Usage example:
//
//	err := client.RemoveMemberRole(guildID, userID, roleID, "Removing role")
func (r *restApi) RemoveMemberRole(guildID, userID, roleID Snowflake, reason string) error {
	_, err := r.doRequest("DELETE", "/guilds/"+guildID.String()+"/members/"+userID.String()+"/roles/"+roleID.String(), nil, true, reason)
	return err
}

// ModifyCurrentMemberOptions are options for modifying the current member (bot).
type ModifyCurrentMemberOptions struct {
	// Nick is the value to set the bot's nickname to. Requires CHANGE_NICKNAME permission.
	Nick *string `json:"nick,omitempty"`
}

// ModifyCurrentMember modifies the bot's own nickname in a guild.
// Requires CHANGE_NICKNAME permission.
//
// Usage example:
//
//	nick := "Bot Nickname"
//	member, err := client.ModifyCurrentMember(guildID, ModifyCurrentMemberOptions{
//	    Nick: &nick,
//	}, "Changing bot nickname")
func (r *restApi) ModifyCurrentMember(guildID Snowflake, opts ModifyCurrentMemberOptions, reason string) (Member, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("PATCH", "/guilds/"+guildID.String()+"/members/@me", reqBody, true, reason)
	if err != nil {
		return Member{}, err
	}

	var member Member
	if err := json.Unmarshal(body, &member); err != nil {
		r.logger.Error("Failed parsing response for PATCH /guilds/{id}/members/@me: " + err.Error())
		return Member{}, err
	}
	member.GuildID = guildID
	return member, nil
}

// TimeoutMember times out (mutes) a member for a specified duration.
// This is a convenience method that wraps EditMember.
// Requires MODERATE_MEMBERS permission.
//
// Usage example:
//
//	err := client.TimeoutMember(guildID, userID, 10*time.Minute, "Spam")
func (r *restApi) TimeoutMember(guildID, userID Snowflake, duration time.Duration, reason string) error {
	until := time.Now().Add(duration)
	_, err := r.EditMember(guildID, userID, MemberEditOptions{
		CommunicationDisabledUntil: &until,
	}, reason)
	return err
}

// RemoveTimeout removes a timeout from a member.
// This is a convenience method that wraps EditMember.
// Requires MODERATE_MEMBERS permission.
//
// Usage example:
//
//	err := client.RemoveTimeout(guildID, userID, "Timeout lifted")
func (r *restApi) RemoveTimeout(guildID, userID Snowflake, reason string) error {
	_, err := r.EditMember(guildID, userID, MemberEditOptions{
		CommunicationDisabledUntil: nil,
	}, reason)
	return err
}

/*******************************************************************************
 *                              ROLE METHODS
 *******************************************************************************/

// FetchRoles retrieves all roles for a guild.
//
// Usage example:
//
//	roles, err := client.FetchRoles(guildID)
func (r *restApi) FetchRoles(guildID Snowflake) ([]Role, error) {
	body, err := r.doRequest("GET", "/guilds/"+guildID.String()+"/roles", nil, true, "")
	if err != nil {
		return nil, err
	}

	var roles []Role
	if err := json.Unmarshal(body, &roles); err != nil {
		r.logger.Error("Failed parsing response for GET /guilds/{id}/roles: " + err.Error())
		return nil, err
	}

	// Set guild ID on all roles
	for i := range roles {
		roles[i].GuildID = guildID
	}
	return roles, nil
}

// RoleCreateOptions are options for creating a role.
type RoleCreateOptions struct {
	// Name is the name of the role (max 100 characters). Default is "new role".
	Name string `json:"name,omitempty"`
	// Permissions is the bitwise value of the enabled/disabled permissions.
	Permissions *Permissions `json:"permissions,omitempty,string"`
	// Color is the RGB color value. Default is 0 (no color).
	Color Color `json:"color,omitempty"`
	// Hoist indicates whether the role should be displayed separately in the sidebar.
	Hoist bool `json:"hoist,omitempty"`
	// Icon is the role's icon image (if the guild has the feature).
	Icon *ImageFile `json:"icon,omitempty"`
	// UnicodeEmoji is the role's unicode emoji as a standard emoji.
	UnicodeEmoji string `json:"unicode_emoji,omitempty"`
	// Mentionable indicates whether the role should be mentionable.
	Mentionable bool `json:"mentionable,omitempty"`
}

// CreateRole creates a new role for a guild.
// Requires MANAGE_ROLES permission.
//
// Usage example:
//
//	role, err := client.CreateRole(guildID, RoleCreateOptions{
//	    Name: "Moderator",
//	    Color: 0x3498db,
//	    Hoist: true,
//	    Mentionable: true,
//	}, "Creating moderator role")
func (r *restApi) CreateRole(guildID Snowflake, opts RoleCreateOptions, reason string) (Role, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("POST", "/guilds/"+guildID.String()+"/roles", reqBody, true, reason)
	if err != nil {
		return Role{}, err
	}

	var role Role
	if err := json.Unmarshal(body, &role); err != nil {
		r.logger.Error("Failed parsing response for POST /guilds/{id}/roles: " + err.Error())
		return Role{}, err
	}
	role.GuildID = guildID
	return role, nil
}

// RoleEditOptions are options for editing a role.
type RoleEditOptions struct {
	// Name is the name of the role (max 100 characters).
	Name string `json:"name,omitempty"`
	// Permissions is the bitwise value of the enabled/disabled permissions.
	Permissions *Permissions `json:"permissions,omitempty,string"`
	// Color is the RGB color value.
	Color *Color `json:"color,omitempty"`
	// Hoist indicates whether the role should be displayed separately in the sidebar.
	Hoist *bool `json:"hoist,omitempty"`
	// Icon is the role's icon image (if the guild has the feature).
	Icon *ImageFile `json:"icon,omitempty"`
	// UnicodeEmoji is the role's unicode emoji as a standard emoji.
	UnicodeEmoji *string `json:"unicode_emoji,omitempty"`
	// Mentionable indicates whether the role should be mentionable.
	Mentionable *bool `json:"mentionable,omitempty"`
}

// EditRole modifies a guild role.
// Requires MANAGE_ROLES permission.
//
// Usage example:
//
//	role, err := client.EditRole(guildID, roleID, RoleEditOptions{
//	    Name: "Senior Moderator",
//	}, "Promoting role")
func (r *restApi) EditRole(guildID, roleID Snowflake, opts RoleEditOptions, reason string) (Role, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("PATCH", "/guilds/"+guildID.String()+"/roles/"+roleID.String(), reqBody, true, reason)
	if err != nil {
		return Role{}, err
	}

	var role Role
	if err := json.Unmarshal(body, &role); err != nil {
		r.logger.Error("Failed parsing response for PATCH /guilds/{id}/roles/{id}: " + err.Error())
		return Role{}, err
	}
	role.GuildID = guildID
	return role, nil
}

// DeleteRole deletes a guild role.
// Requires MANAGE_ROLES permission.
//
// Usage example:
//
//	err := client.DeleteRole(guildID, roleID, "Role no longer needed")
func (r *restApi) DeleteRole(guildID, roleID Snowflake, reason string) error {
	_, err := r.doRequest("DELETE", "/guilds/"+guildID.String()+"/roles/"+roleID.String(), nil, true, reason)
	return err
}

// ModifyRolePositionsEntry represents a role position modification.
type ModifyRolePositionsEntry struct {
	// ID is the role id.
	ID Snowflake `json:"id"`
	// Position is the sorting position of the role.
	Position *int `json:"position,omitempty"`
}

// ModifyRolePositions modifies the positions of roles in a guild.
// Requires MANAGE_ROLES permission.
//
// Usage example:
//
//	roles, err := client.ModifyRolePositions(guildID, []ModifyRolePositionsEntry{
//	    {ID: roleID1, Position: intPtr(1)},
//	    {ID: roleID2, Position: intPtr(2)},
//	}, "Reordering roles")
func (r *restApi) ModifyRolePositions(guildID Snowflake, positions []ModifyRolePositionsEntry, reason string) ([]Role, error) {
	reqBody, _ := json.Marshal(positions)
	body, err := r.doRequest("PATCH", "/guilds/"+guildID.String()+"/roles", reqBody, true, reason)
	if err != nil {
		return nil, err
	}

	var roles []Role
	if err := json.Unmarshal(body, &roles); err != nil {
		r.logger.Error("Failed parsing response for PATCH /guilds/{id}/roles: " + err.Error())
		return nil, err
	}

	// Set guild ID on all roles
	for i := range roles {
		roles[i].GuildID = guildID
	}
	return roles, nil
}

/*******************************************************************************
 *                              BAN METHODS
 *******************************************************************************/

// Ban represents a guild ban.
type Ban struct {
	// Reason is the reason for the ban.
	Reason string `json:"reason"`
	// User is the banned user.
	User User `json:"user"`
}

// BanOptions are options for banning a guild member.
type BanOptions struct {
	// DeleteMessageSeconds is the number of seconds to delete messages for (0-604800).
	// 0 deletes no messages, 604800 (7 days) is the maximum.
	DeleteMessageSeconds int `json:"delete_message_seconds,omitempty"`
}

// BanMember bans a user from a guild, and optionally deletes previous messages sent by them.
// Requires BAN_MEMBERS permission.
//
// Usage example:
//
//	err := client.BanMember(guildID, userID, BanOptions{
//	    DeleteMessageSeconds: 86400, // Delete 1 day of messages
//	}, "Rule violation")
func (r *restApi) BanMember(guildID, userID Snowflake, opts BanOptions, reason string) error {
	reqBody, _ := json.Marshal(opts)
	_, err := r.doRequest("PUT", "/guilds/"+guildID.String()+"/bans/"+userID.String(), reqBody, true, reason)
	return err
}

// UnbanMember removes the ban for a user.
// Requires BAN_MEMBERS permission.
//
// Usage example:
//
//	err := client.UnbanMember(guildID, userID, "Appeal accepted")
func (r *restApi) UnbanMember(guildID, userID Snowflake, reason string) error {
	_, err := r.doRequest("DELETE", "/guilds/"+guildID.String()+"/bans/"+userID.String(), nil, true, reason)
	return err
}

// GetBan retrieves the ban for a specific user.
// Requires BAN_MEMBERS permission.
//
// Usage example:
//
//	ban, err := client.GetBan(guildID, userID)
func (r *restApi) GetBan(guildID, userID Snowflake) (Ban, error) {
	body, err := r.doRequest("GET", "/guilds/"+guildID.String()+"/bans/"+userID.String(), nil, true, "")
	if err != nil {
		return Ban{}, err
	}

	var ban Ban
	if err := json.Unmarshal(body, &ban); err != nil {
		r.logger.Error("Failed parsing response for GET /guilds/{id}/bans/{id}: " + err.Error())
		return Ban{}, err
	}
	return ban, nil
}

// ListBansOptions are options for listing guild bans.
type ListBansOptions struct {
	// Limit is the number of users to return (1-1000). Default is 1000.
	Limit int
	// Before is the user id to get users before.
	Before Snowflake
	// After is the user id to get users after.
	After Snowflake
}

// ListBans retrieves a list of banned users for a guild.
// Requires BAN_MEMBERS permission.
//
// Usage example:
//
//	bans, err := client.ListBans(guildID, ListBansOptions{Limit: 100})
func (r *restApi) ListBans(guildID Snowflake, opts ListBansOptions) ([]Ban, error) {
	query := url.Values{}
	if opts.Limit > 0 {
		if opts.Limit > 1000 {
			opts.Limit = 1000
		}
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if !opts.Before.UnSet() {
		query.Set("before", opts.Before.String())
	}
	if !opts.After.UnSet() {
		query.Set("after", opts.After.String())
	}

	endpoint := "/guilds/" + guildID.String() + "/bans"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	body, err := r.doRequest("GET", endpoint, nil, true, "")
	if err != nil {
		return nil, err
	}

	var bans []Ban
	if err := json.Unmarshal(body, &bans); err != nil {
		r.logger.Error("Failed parsing response for GET /guilds/{id}/bans: " + err.Error())
		return nil, err
	}
	return bans, nil
}

// BulkBanOptions are options for bulk banning users.
type BulkBanOptions struct {
	// UserIDs is a list of user ids to ban (max 200).
	UserIDs []Snowflake `json:"user_ids"`
	// DeleteMessageSeconds is the number of seconds to delete messages for (0-604800).
	DeleteMessageSeconds int `json:"delete_message_seconds,omitempty"`
}

// BulkBanResponse is the response from a bulk ban request.
type BulkBanResponse struct {
	// BannedUsers is a list of user ids that were banned.
	BannedUsers []Snowflake `json:"banned_users"`
	// FailedUsers is a list of user ids that could not be banned.
	FailedUsers []Snowflake `json:"failed_users"`
}

// BulkBanMembers bans up to 200 users from a guild.
// Requires BAN_MEMBERS and MANAGE_GUILD permissions.
//
// Usage example:
//
//	response, err := client.BulkBanMembers(guildID, BulkBanOptions{
//	    UserIDs: []Snowflake{userID1, userID2, userID3},
//	    DeleteMessageSeconds: 86400,
//	}, "Mass rule violation")
func (r *restApi) BulkBanMembers(guildID Snowflake, opts BulkBanOptions, reason string) (BulkBanResponse, error) {
	reqBody, _ := json.Marshal(opts)
	body, err := r.doRequest("POST", "/guilds/"+guildID.String()+"/bulk-ban", reqBody, true, reason)
	if err != nil {
		return BulkBanResponse{}, err
	}

	var response BulkBanResponse
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("Failed parsing response for POST /guilds/{id}/bulk-ban: " + err.Error())
		return BulkBanResponse{}, err
	}
	return response, nil
}

/*******************************************************************************
 *                          INTERACTION METHODS
 *******************************************************************************/

// InteractionResponseType is the type of response to an interaction.
type InteractionResponseType int

const (
	// InteractionResponseTypePong acknowledges a ping.
	InteractionResponseTypePong InteractionResponseType = 1
	// InteractionResponseTypeChannelMessageWithSource responds with a message, showing the user's input.
	InteractionResponseTypeChannelMessageWithSource InteractionResponseType = 4
	// InteractionResponseTypeDeferredChannelMessageWithSource acknowledges, showing a loading state.
	InteractionResponseTypeDeferredChannelMessageWithSource InteractionResponseType = 5
	// InteractionResponseTypeDeferredUpdateMessage acknowledges without updating.
	InteractionResponseTypeDeferredUpdateMessage InteractionResponseType = 6
	// InteractionResponseTypeUpdateMessage edits the message the component was attached to.
	InteractionResponseTypeUpdateMessage InteractionResponseType = 7
	// InteractionResponseTypeApplicationCommandAutocompleteResult responds to an autocomplete interaction.
	InteractionResponseTypeApplicationCommandAutocompleteResult InteractionResponseType = 8
	// InteractionResponseTypeModal responds with a popup modal.
	InteractionResponseTypeModal InteractionResponseType = 9
	// InteractionResponseTypePremiumRequired responds to an interaction with an upgrade button.
	InteractionResponseTypePremiumRequired InteractionResponseType = 10
	// InteractionResponseTypeLaunchActivity launches an activity.
	InteractionResponseTypeLaunchActivity InteractionResponseType = 12
)

// InteractionResponseData is the data payload for an interaction response.
type InteractionResponseData struct {
	// TTS indicates if the message is text-to-speech.
	TTS bool `json:"tts,omitempty"`
	// Content is the message content (up to 2000 characters).
	Content string `json:"content,omitempty"`
	// Embeds are the embeds for the message (up to 10).
	Embeds []Embed `json:"embeds,omitempty"`
	// AllowedMentions are allowed mentions for the message.
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`
	// Flags are message flags (only SUPPRESS_EMBEDS and EPHEMERAL can be set).
	Flags MessageFlags `json:"flags,omitempty"`
	// Components are message components.
	Components []LayoutComponent `json:"components,omitempty"`
	// Attachments are attachment objects with filename and description.
	Attachments []Attachment `json:"attachments,omitempty"`
	// Poll is a poll for the message.
	Poll *PollCreateOptions `json:"poll,omitempty"`
	// Choices are autocomplete choices (max 25).
	Choices []ApplicationCommandOptionChoice `json:"choices,omitempty"`
	// CustomID is the custom id for a modal.
	CustomID string `json:"custom_id,omitempty"`
	// Title is the title for a modal (max 45 characters).
	Title string `json:"title,omitempty"`
}

// InteractionResponse is the response structure for an interaction.
type InteractionResponse struct {
	// Type is the type of response.
	Type InteractionResponseType `json:"type"`
	// Data is an optional response message.
	Data *InteractionResponseData `json:"data,omitempty"`
}

// CreateInteractionResponse responds to an interaction.
// This must be called within 3 seconds of receiving the interaction.
//
// Usage example:
//
//	err := client.CreateInteractionResponse(interactionID, interactionToken, InteractionResponse{
//	    Type: InteractionResponseTypeChannelMessageWithSource,
//	    Data: &InteractionResponseData{
//	        Content: "Hello!",
//	    },
//	})
func (r *restApi) CreateInteractionResponse(interactionID Snowflake, token string, response InteractionResponse) error {
	reqBody, _ := json.Marshal(response)
	// Note: Interaction responses don't use bot token auth
	_, err := r.doRequest("POST", "/interactions/"+interactionID.String()+"/"+token+"/callback", reqBody, false, "")
	return err
}

// GetOriginalInteractionResponse retrieves the initial response to an interaction.
//
// Usage example:
//
//	message, err := client.GetOriginalInteractionResponse(applicationID, interactionToken)
func (r *restApi) GetOriginalInteractionResponse(applicationID Snowflake, token string) (Message, error) {
	body, err := r.doRequest("GET", "/webhooks/"+applicationID.String()+"/"+token+"/messages/@original", nil, false, "")
	if err != nil {
		return Message{}, err
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		r.logger.Error("Failed parsing response for GET original interaction response: " + err.Error())
		return Message{}, err
	}
	return message, nil
}

// EditOriginalInteractionResponse edits the initial response to an interaction.
//
// Usage example:
//
//	message, err := client.EditOriginalInteractionResponse(applicationID, interactionToken, InteractionResponseData{
//	    Content: "Updated content!",
//	})
func (r *restApi) EditOriginalInteractionResponse(applicationID Snowflake, token string, data InteractionResponseData) (Message, error) {
	reqBody, _ := json.Marshal(data)
	body, err := r.doRequest("PATCH", "/webhooks/"+applicationID.String()+"/"+token+"/messages/@original", reqBody, false, "")
	if err != nil {
		return Message{}, err
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		r.logger.Error("Failed parsing response for PATCH original interaction response: " + err.Error())
		return Message{}, err
	}
	return message, nil
}

// DeleteOriginalInteractionResponse deletes the initial response to an interaction.
//
// Usage example:
//
//	err := client.DeleteOriginalInteractionResponse(applicationID, interactionToken)
func (r *restApi) DeleteOriginalInteractionResponse(applicationID Snowflake, token string) error {
	_, err := r.doRequest("DELETE", "/webhooks/"+applicationID.String()+"/"+token+"/messages/@original", nil, false, "")
	return err
}

// CreateFollowupMessage creates a followup message for an interaction.
//
// Usage example:
//
//	message, err := client.CreateFollowupMessage(applicationID, interactionToken, InteractionResponseData{
//	    Content: "Followup message!",
//	})
func (r *restApi) CreateFollowupMessage(applicationID Snowflake, token string, data InteractionResponseData) (Message, error) {
	reqBody, _ := json.Marshal(data)
	body, err := r.doRequest("POST", "/webhooks/"+applicationID.String()+"/"+token, reqBody, false, "")
	if err != nil {
		return Message{}, err
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		r.logger.Error("Failed parsing response for POST followup message: " + err.Error())
		return Message{}, err
	}
	return message, nil
}

// GetFollowupMessage retrieves a followup message for an interaction.
//
// Usage example:
//
//	message, err := client.GetFollowupMessage(applicationID, interactionToken, messageID)
func (r *restApi) GetFollowupMessage(applicationID Snowflake, token string, messageID Snowflake) (Message, error) {
	body, err := r.doRequest("GET", "/webhooks/"+applicationID.String()+"/"+token+"/messages/"+messageID.String(), nil, false, "")
	if err != nil {
		return Message{}, err
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		r.logger.Error("Failed parsing response for GET followup message: " + err.Error())
		return Message{}, err
	}
	return message, nil
}

// EditFollowupMessage edits a followup message for an interaction.
//
// Usage example:
//
//	message, err := client.EditFollowupMessage(applicationID, interactionToken, messageID, InteractionResponseData{
//	    Content: "Edited followup!",
//	})
func (r *restApi) EditFollowupMessage(applicationID Snowflake, token string, messageID Snowflake, data InteractionResponseData) (Message, error) {
	reqBody, _ := json.Marshal(data)
	body, err := r.doRequest("PATCH", "/webhooks/"+applicationID.String()+"/"+token+"/messages/"+messageID.String(), reqBody, false, "")
	if err != nil {
		return Message{}, err
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		r.logger.Error("Failed parsing response for PATCH followup message: " + err.Error())
		return Message{}, err
	}
	return message, nil
}

// DeleteFollowupMessage deletes a followup message for an interaction.
//
// Usage example:
//
//	err := client.DeleteFollowupMessage(applicationID, interactionToken, messageID)
func (r *restApi) DeleteFollowupMessage(applicationID Snowflake, token string, messageID Snowflake) error {
	_, err := r.doRequest("DELETE", "/webhooks/"+applicationID.String()+"/"+token+"/messages/"+messageID.String(), nil, false, "")
	return err
}

/*******************************************************************************
 *                      APPLICATION COMMAND METHODS
 *******************************************************************************/

// GetGlobalApplicationCommands retrieves all global application commands.
//
// Usage example:
//
//	commands, err := client.GetGlobalApplicationCommands(applicationID)
func (r *restApi) GetGlobalApplicationCommands(applicationID Snowflake) ([]ApplicationCommand, error) {
	body, err := r.doRequest("GET", "/applications/"+applicationID.String()+"/commands", nil, true, "")
	if err != nil {
		return nil, err
	}

	var commands []ApplicationCommand
	if err := json.Unmarshal(body, &commands); err != nil {
		r.logger.Error("Failed parsing response for GET global commands: " + err.Error())
		return nil, err
	}
	return commands, nil
}

// CreateGlobalApplicationCommand creates a new global application command.
//
// Usage example:
//
//	command, err := client.CreateGlobalApplicationCommand(applicationID, ApplicationCommand{
//	    Name: "ping",
//	    Description: "Replies with pong",
//	})
func (r *restApi) CreateGlobalApplicationCommand(applicationID Snowflake, command ApplicationCommand) (ApplicationCommand, error) {
	reqBody, _ := json.Marshal(command)
	body, err := r.doRequest("POST", "/applications/"+applicationID.String()+"/commands", reqBody, true, "")
	if err != nil {
		return nil, err
	}

	result, err := UnmarshalApplicationCommand(body)
	if err != nil {
		r.logger.Error("Failed parsing response for POST global command: " + err.Error())
		return nil, err
	}
	return result, nil
}

// BulkOverwriteGlobalCommands overwrites all global application commands.
// This will replace all existing global commands.
//
// Usage example:
//
//	commands, err := client.BulkOverwriteGlobalCommands(applicationID, []ApplicationCommand{
//	    {Name: "ping", Description: "Pong!"},
//	    {Name: "help", Description: "Get help"},
//	})
func (r *restApi) BulkOverwriteGlobalCommands(applicationID Snowflake, commands []ApplicationCommand) ([]ApplicationCommand, error) {
	reqBody, _ := json.Marshal(commands)
	body, err := r.doRequest("PUT", "/applications/"+applicationID.String()+"/commands", reqBody, true, "")
	if err != nil {
		return nil, err
	}

	var result []ApplicationCommand
	if err := json.Unmarshal(body, &result); err != nil {
		r.logger.Error("Failed parsing response for PUT global commands: " + err.Error())
		return nil, err
	}
	return result, nil
}

// DeleteGlobalApplicationCommand deletes a global application command.
//
// Usage example:
//
//	err := client.DeleteGlobalApplicationCommand(applicationID, commandID)
func (r *restApi) DeleteGlobalApplicationCommand(applicationID, commandID Snowflake) error {
	_, err := r.doRequest("DELETE", "/applications/"+applicationID.String()+"/commands/"+commandID.String(), nil, true, "")
	return err
}

// GetGuildApplicationCommands retrieves all guild-specific application commands.
//
// Usage example:
//
//	commands, err := client.GetGuildApplicationCommands(applicationID, guildID)
func (r *restApi) GetGuildApplicationCommands(applicationID, guildID Snowflake) ([]ApplicationCommand, error) {
	body, err := r.doRequest("GET", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands", nil, true, "")
	if err != nil {
		return nil, err
	}

	var commands []ApplicationCommand
	if err := json.Unmarshal(body, &commands); err != nil {
		r.logger.Error("Failed parsing response for GET guild commands: " + err.Error())
		return nil, err
	}
	return commands, nil
}

// CreateGuildApplicationCommand creates a new guild-specific application command.
//
// Usage example:
//
//	command, err := client.CreateGuildApplicationCommand(applicationID, guildID, ApplicationCommand{
//	    Name: "test",
//	    Description: "A test command",
//	})
func (r *restApi) CreateGuildApplicationCommand(applicationID, guildID Snowflake, command ApplicationCommand) (ApplicationCommand, error) {
	reqBody, _ := json.Marshal(command)
	body, err := r.doRequest("POST", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands", reqBody, true, "")
	if err != nil {
		return nil, err
	}

	result, err := UnmarshalApplicationCommand(body)
	if err != nil {
		r.logger.Error("Failed parsing response for POST guild command: " + err.Error())
		return nil, err
	}
	return result, nil
}

// BulkOverwriteGuildCommands overwrites all guild-specific application commands.
//
// Usage example:
//
//	commands, err := client.BulkOverwriteGuildCommands(applicationID, guildID, []ApplicationCommand{
//	    {Name: "admin", Description: "Admin command"},
//	})
func (r *restApi) BulkOverwriteGuildCommands(applicationID, guildID Snowflake, commands []ApplicationCommand) ([]ApplicationCommand, error) {
	reqBody, _ := json.Marshal(commands)
	body, err := r.doRequest("PUT", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands", reqBody, true, "")
	if err != nil {
		return nil, err
	}

	var result []ApplicationCommand
	if err := json.Unmarshal(body, &result); err != nil {
		r.logger.Error("Failed parsing response for PUT guild commands: " + err.Error())
		return nil, err
	}
	return result, nil
}

// DeleteGuildApplicationCommand deletes a guild-specific application command.
//
// Usage example:
//
//	err := client.DeleteGuildApplicationCommand(applicationID, guildID, commandID)
func (r *restApi) DeleteGuildApplicationCommand(applicationID, guildID, commandID Snowflake) error {
	_, err := r.doRequest("DELETE", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands/"+commandID.String(), nil, true, "")
	return err
}
