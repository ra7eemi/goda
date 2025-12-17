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

/*****************************
 *  GuildChannelManager      *
 *****************************/

// GuildChannelManager manages channels within a guild.
// It provides methods to get, fetch, create, and filter channels.
type GuildChannelManager struct {
	client  *Client
	guildID Snowflake
	cache   *Collection[Snowflake, GuildChannel]
}

// NewGuildChannelManager creates a new GuildChannelManager.
func NewGuildChannelManager(client *Client, guildID Snowflake) *GuildChannelManager {
	return &GuildChannelManager{
		client:  client,
		guildID: guildID,
		cache:   NewCollection[Snowflake, GuildChannel](),
	}
}

// Get retrieves a channel from the cache.
// Returns the channel and true if found, or nil and false if not found.
func (m *GuildChannelManager) Get(channelID Snowflake) (GuildChannel, bool) {
	return m.cache.Get(channelID)
}

// Fetch fetches a channel from the API and updates the cache.
// Returns the channel.
func (m *GuildChannelManager) Fetch(channelID Snowflake) (GuildChannel, error) {
	ch, err := m.client.FetchChannel(channelID)
	if err != nil {
		return nil, err
	}
	gc, ok := ch.(GuildChannel)
	if !ok {
		return nil, ErrChannelNotText
	}
	m.cache.Set(channelID, gc)
	return gc, nil
}

// Create creates a new channel in the guild.
// Returns the created channel.
func (m *GuildChannelManager) Create(opts ChannelCreateOptions, reason string) (GuildChannel, error) {
	ch, err := m.client.CreateGuildChannel(m.guildID, opts, reason)
	if err != nil {
		return nil, err
	}
	gc, ok := ch.(GuildChannel)
	if !ok {
		return nil, ErrChannelNotText
	}
	m.cache.Set(gc.GetID(), gc)
	return gc, nil
}

// All returns all cached channels.
func (m *GuildChannelManager) All() []GuildChannel {
	return m.cache.Values()
}

// Size returns the number of cached channels.
func (m *GuildChannelManager) Size() int {
	return m.cache.Size()
}

// Filter returns channels matching the predicate.
func (m *GuildChannelManager) Filter(fn func(GuildChannel) bool) []GuildChannel {
	return m.cache.Filter(fn)
}

// Find returns the first channel matching the predicate.
func (m *GuildChannelManager) Find(fn func(GuildChannel) bool) (GuildChannel, bool) {
	return m.cache.Find(fn)
}

// ByName finds a channel by name.
func (m *GuildChannelManager) ByName(name string) (GuildChannel, bool) {
	return m.cache.Find(func(ch GuildChannel) bool {
		return ch.GetName() == name
	})
}

// ByType returns all channels of a specific type.
func (m *GuildChannelManager) ByType(t ChannelType) []GuildChannel {
	return m.cache.Filter(func(ch GuildChannel) bool {
		return ch.GetType() == t
	})
}

// TextChannels returns all text channels.
func (m *GuildChannelManager) TextChannels() []*TextChannel {
	result := make([]*TextChannel, 0)
	for _, ch := range m.cache.Values() {
		if tc, ok := ch.(*TextChannel); ok {
			result = append(result, tc)
		}
	}
	return result
}

// VoiceChannels returns all voice channels.
func (m *GuildChannelManager) VoiceChannels() []*VoiceChannel {
	result := make([]*VoiceChannel, 0)
	for _, ch := range m.cache.Values() {
		if vc, ok := ch.(*VoiceChannel); ok {
			result = append(result, vc)
		}
	}
	return result
}

// Categories returns all category channels.
func (m *GuildChannelManager) Categories() []*CategoryChannel {
	result := make([]*CategoryChannel, 0)
	for _, ch := range m.cache.Values() {
		if cc, ok := ch.(*CategoryChannel); ok {
			result = append(result, cc)
		}
	}
	return result
}

// set adds a channel to the cache (internal use).
func (m *GuildChannelManager) set(ch GuildChannel) {
	m.cache.Set(ch.GetID(), ch)
}

// delete removes a channel from the cache (internal use).
func (m *GuildChannelManager) delete(channelID Snowflake) {
	m.cache.Delete(channelID)
}

/*****************************
 *  GuildMemberManager       *
 *****************************/

// GuildMemberManager manages members within a guild.
// It provides methods to get, fetch, search, and moderate members.
type GuildMemberManager struct {
	client  *Client
	guildID Snowflake
	cache   *Collection[Snowflake, *Member]
}

// NewGuildMemberManager creates a new GuildMemberManager.
func NewGuildMemberManager(client *Client, guildID Snowflake) *GuildMemberManager {
	return &GuildMemberManager{
		client:  client,
		guildID: guildID,
		cache:   NewCollection[Snowflake, *Member](),
	}
}

// Get retrieves a member from the cache by user ID.
// Returns the member and true if found, or nil and false if not found.
func (m *GuildMemberManager) Get(userID Snowflake) (*Member, bool) {
	return m.cache.Get(userID)
}

// Fetch fetches a member from the API and updates the cache.
// Returns the member.
func (m *GuildMemberManager) Fetch(userID Snowflake) (*Member, error) {
	member, err := m.client.FetchMember(m.guildID, userID)
	if err != nil {
		return nil, err
	}
	member.SetClient(m.client)
	m.cache.Set(userID, &member)
	return &member, nil
}

// FetchAll fetches members from the API with pagination.
// Note: Requires GUILD_MEMBERS privileged intent.
func (m *GuildMemberManager) FetchAll(opts ListMembersOptions) ([]*Member, error) {
	members, err := m.client.ListMembers(m.guildID, opts)
	if err != nil {
		return nil, err
	}
	result := make([]*Member, len(members))
	for i := range members {
		members[i].SetClient(m.client)
		m.cache.Set(members[i].User.ID, &members[i])
		result[i] = &members[i]
	}
	return result, nil
}

// Search searches for members by username or nickname.
// Returns up to `limit` members.
func (m *GuildMemberManager) Search(query string, limit int) ([]*Member, error) {
	members, err := m.client.SearchMembers(m.guildID, query, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*Member, len(members))
	for i := range members {
		members[i].SetClient(m.client)
		m.cache.Set(members[i].User.ID, &members[i])
		result[i] = &members[i]
	}
	return result, nil
}

// Kick kicks a member from the guild.
// Requires KICK_MEMBERS permission.
func (m *GuildMemberManager) Kick(userID Snowflake, reason string) error {
	return m.client.KickMember(m.guildID, userID, reason)
}

// Ban bans a user from the guild.
// Requires BAN_MEMBERS permission.
func (m *GuildMemberManager) Ban(userID Snowflake, opts BanOptions, reason string) error {
	return m.client.BanMember(m.guildID, userID, opts, reason)
}

// Unban unbans a user from the guild.
// Requires BAN_MEMBERS permission.
func (m *GuildMemberManager) Unban(userID Snowflake, reason string) error {
	return m.client.UnbanMember(m.guildID, userID, reason)
}

// All returns all cached members.
func (m *GuildMemberManager) All() []*Member {
	return m.cache.Values()
}

// Size returns the number of cached members.
func (m *GuildMemberManager) Size() int {
	return m.cache.Size()
}

// Filter returns members matching the predicate.
func (m *GuildMemberManager) Filter(fn func(*Member) bool) []*Member {
	return m.cache.Filter(fn)
}

// Find returns the first member matching the predicate.
func (m *GuildMemberManager) Find(fn func(*Member) bool) (*Member, bool) {
	return m.cache.Find(fn)
}

// ByUsername finds a member by username.
func (m *GuildMemberManager) ByUsername(username string) (*Member, bool) {
	return m.cache.Find(func(member *Member) bool {
		return member.User.Username == username
	})
}

// ByNickname finds a member by nickname.
func (m *GuildMemberManager) ByNickname(nickname string) (*Member, bool) {
	return m.cache.Find(func(member *Member) bool {
		return member.Nickname == nickname
	})
}

// WithRole returns all members that have a specific role.
func (m *GuildMemberManager) WithRole(roleID Snowflake) []*Member {
	return m.cache.Filter(func(member *Member) bool {
		return member.HasRole(roleID)
	})
}

// set adds a member to the cache (internal use).
func (m *GuildMemberManager) set(member *Member) {
	m.cache.Set(member.User.ID, member)
}

// delete removes a member from the cache (internal use).
func (m *GuildMemberManager) delete(userID Snowflake) {
	m.cache.Delete(userID)
}

/*****************************
 *   GuildRoleManager        *
 *****************************/

// GuildRoleManager manages roles within a guild.
// It provides methods to get, create, and filter roles.
type GuildRoleManager struct {
	client  *Client
	guildID Snowflake
	cache   *Collection[Snowflake, *Role]
}

// NewGuildRoleManager creates a new GuildRoleManager.
func NewGuildRoleManager(client *Client, guildID Snowflake) *GuildRoleManager {
	return &GuildRoleManager{
		client:  client,
		guildID: guildID,
		cache:   NewCollection[Snowflake, *Role](),
	}
}

// Get retrieves a role from the cache.
// Returns the role and true if found, or nil and false if not found.
func (m *GuildRoleManager) Get(roleID Snowflake) (*Role, bool) {
	return m.cache.Get(roleID)
}

// Fetch fetches all roles from the API and updates the cache.
// Returns all roles.
func (m *GuildRoleManager) Fetch() ([]*Role, error) {
	roles, err := m.client.FetchRoles(m.guildID)
	if err != nil {
		return nil, err
	}
	result := make([]*Role, len(roles))
	for i := range roles {
		roles[i].SetClient(m.client)
		m.cache.Set(roles[i].ID, &roles[i])
		result[i] = &roles[i]
	}
	return result, nil
}

// Create creates a new role in the guild.
// Requires MANAGE_ROLES permission.
func (m *GuildRoleManager) Create(opts RoleCreateOptions, reason string) (*Role, error) {
	role, err := m.client.CreateRole(m.guildID, opts, reason)
	if err != nil {
		return nil, err
	}
	role.SetClient(m.client)
	m.cache.Set(role.ID, &role)
	return &role, nil
}

// Delete deletes a role from the guild.
// Requires MANAGE_ROLES permission.
func (m *GuildRoleManager) Delete(roleID Snowflake, reason string) error {
	err := m.client.DeleteRole(m.guildID, roleID, reason)
	if err == nil {
		m.cache.Delete(roleID)
	}
	return err
}

// All returns all cached roles.
func (m *GuildRoleManager) All() []*Role {
	return m.cache.Values()
}

// Size returns the number of cached roles.
func (m *GuildRoleManager) Size() int {
	return m.cache.Size()
}

// Filter returns roles matching the predicate.
func (m *GuildRoleManager) Filter(fn func(*Role) bool) []*Role {
	return m.cache.Filter(fn)
}

// Find returns the first role matching the predicate.
func (m *GuildRoleManager) Find(fn func(*Role) bool) (*Role, bool) {
	return m.cache.Find(fn)
}

// ByName finds a role by name.
func (m *GuildRoleManager) ByName(name string) (*Role, bool) {
	return m.cache.Find(func(role *Role) bool {
		return role.Name == name
	})
}

// Everyone returns the @everyone role.
// The @everyone role ID is the same as the guild ID.
func (m *GuildRoleManager) Everyone() (*Role, bool) {
	return m.cache.Get(m.guildID)
}

// Highest returns the role with the highest position.
func (m *GuildRoleManager) Highest() (*Role, bool) {
	var highest *Role
	for _, role := range m.cache.Values() {
		if highest == nil || role.Position > highest.Position {
			highest = role
		}
	}
	if highest == nil {
		return nil, false
	}
	return highest, true
}

// Hoisted returns all roles that are displayed separately.
func (m *GuildRoleManager) Hoisted() []*Role {
	return m.cache.Filter(func(role *Role) bool {
		return role.Hoist
	})
}

// Mentionable returns all mentionable roles.
func (m *GuildRoleManager) Mentionable() []*Role {
	return m.cache.Filter(func(role *Role) bool {
		return role.Mentionable
	})
}

// set adds a role to the cache (internal use).
func (m *GuildRoleManager) set(role *Role) {
	m.cache.Set(role.ID, role)
}

// delete removes a role from the cache (internal use).
func (m *GuildRoleManager) delete(roleID Snowflake) {
	m.cache.Delete(roleID)
}
