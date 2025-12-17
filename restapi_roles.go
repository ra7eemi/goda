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
 *   Role Endpoints    *
 ***********************/

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
