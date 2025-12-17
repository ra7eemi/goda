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
	"time"

	"encoding/json"
)

/***********************
 *  Member Endpoints   *
 ***********************/

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
