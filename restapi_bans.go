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
 *    Ban Endpoints    *
 ***********************/

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
