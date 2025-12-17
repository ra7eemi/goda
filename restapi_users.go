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
 *   User Endpoints    *
 ***********************/

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
	ID                Snowflake `json:"id"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	Enabled           bool      `json:"enabled"`
	Syncing           bool      `json:"syncing"`
	RoleID            Snowflake `json:"role_id"`
	EnableEmoticons   bool      `json:"enable_emoticons"`
	ExpireBehavior    int       `json:"expire_behavior"`
	ExpireGracePeriod int       `json:"expire_grace_period"`
	User              *User     `json:"user"`
	Account           Account   `json:"account"`
	SyncedAt          string    `json:"synced_at"`
	SubscriberCount   int       `json:"subscriber_count"`
	Revoked           bool      `json:"revoked"`
	Application       *Application `json:"application"`
}

// Account represents an integration account.
type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
