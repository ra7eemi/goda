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

type CacheFlags int

const (
	CacheFlagUsers CacheFlags = 1 << iota
	CacheFlagGuilds
	CacheFlagMembers
	CacheFlagThreadMembers
	CacheFlagMessages
	CacheFlagChannels
	CacheFlagRoles
	CacheFlagVoiceStates

	CacheFlagsNone CacheFlags = 0

	CacheFlagsAll = CacheFlagUsers | CacheFlagGuilds | CacheFlagMembers | CacheFlagThreadMembers |
		CacheFlagMessages | CacheFlagChannels | CacheFlagRoles | CacheFlagVoiceStates
)

func (f CacheFlags) Has(bits ...CacheFlags) bool {
	return BitFieldHas(f, bits...)
}

type SnowflakePairKey struct {
	A Snowflake
	B Snowflake
}

type CacheManager interface {
	Flags() CacheFlags
	SetFlags(flags ...CacheFlags)

	GetUser(userID Snowflake) (User, bool)
	GetGuild(guildID Snowflake) (Guild, bool)
	GetMember(guildID, userID Snowflake) (Member, bool)
	GetChannel(channelID Snowflake) (Channel, bool)
	GetMessage(messageID Snowflake) (Message, bool)
	GetVoiceState(guildID, userID Snowflake) (VoiceState, bool)
	GetGuildChannels(guildID Snowflake) (map[Snowflake]GuildChannel, bool)
	GetGuildMembers(guildID Snowflake) (map[Snowflake]Member, bool)
	GetGuildVoiceStates(guildID Snowflake) (map[Snowflake]VoiceState, bool)
	GetGuildRoles(guildID Snowflake) (map[Snowflake]Role, bool)

	HasUser(userID Snowflake) bool
	HasGuild(guildID Snowflake) bool
	HasMember(guildID, userID Snowflake) bool
	HasChannel(channelID Snowflake) bool
	HasMessage(messageID Snowflake) bool
	HasVoiceState(guildID, userID Snowflake) bool
	HasGuildChannels(guildID Snowflake) bool
	HasGuildMembers(guildID Snowflake) bool
	HasGuildVoiceStates(guildID Snowflake) bool
	HasGuildRoles(guildID Snowflake) bool

	CountUsers() int
	CountGuilds() int
	CountMembers() int
	CountChannels() int
	CountMessages() int
	CountVoiceStates() int
	CountRoles() int
	CountGuildChannels(guildID Snowflake) int
	CountGuildMembers(guildID Snowflake) int
	CountGuildRoles(guildID Snowflake) int

	PutUser(user User)
	PutGuild(guild Guild)
	PutMember(member Member)
	PutChannel(channel Channel)
	PutMessage(message Message)
	PutVoiceState(voiceState VoiceState)
	PutRole(role Role)

	DelUser(userID Snowflake) bool
	DelGuild(guildID Snowflake) bool
	DelMember(guildID, userID Snowflake) bool
	DelChannel(channelID Snowflake) bool
	DelMessage(messageID Snowflake) bool
	DelVoiceState(guildID, userID Snowflake) bool
	DelGuildChannels(guildID Snowflake) bool
	DelGuildMembers(guildID Snowflake) bool
	DelRole(guildID, roleID Snowflake) bool
}

// DefaultCache is a high-performance cache implementation using 256-way sharding.
// Lock contention is reduced by ~99.6% compared to single-mutex implementations,
// making it suitable for bots with 10,000+ guilds.
type DefaultCache struct {
	flags CacheFlags

	// Primary entity caches using ShardMap for reduced lock contention
	usersCache    *ShardMap[Snowflake, User]
	guildsCache   *ShardMap[Snowflake, Guild]
	membersCache  *ShardMap[SnowflakePairKey, Member]
	channelsCache *ShardMap[Snowflake, Channel]
	messagesCache *ShardMap[Snowflake, Message]
	voiceStates   *ShardMap[SnowflakePairKey, VoiceState]
	rolesCache    *ShardMap[Snowflake, Role]

	// Sharded indexes for guild-to-entity relationships
	guildToMemberIDs         *shardedIndex // guildID -> set[userID]
	guildToChannelIDs        *shardedIndex // guildID -> set[channelID]
	guildToVoiceStateUserIDs *shardedIndex // guildID -> set[userID]
	guildToRoleIDs           *shardedIndex // guildID -> set[roleID]
}

func NewDefaultCache(flags CacheFlags) CacheManager {
	return &DefaultCache{
		flags:                    flags,
		usersCache:               NewSnowflakeShardMap[User](),
		guildsCache:              NewSnowflakeShardMap[Guild](),
		membersCache:             NewSnowflakePairShardMap[Member](),
		channelsCache:            NewSnowflakeShardMap[Channel](),
		messagesCache:            NewSnowflakeShardMap[Message](),
		voiceStates:              NewSnowflakePairShardMap[VoiceState](),
		rolesCache:               NewSnowflakeShardMap[Role](),
		guildToMemberIDs:         newShardedIndex(),
		guildToChannelIDs:        newShardedIndex(),
		guildToVoiceStateUserIDs: newShardedIndex(),
		guildToRoleIDs:           newShardedIndex(),
	}
}

func (c *DefaultCache) Flags() CacheFlags {
	return c.flags
}

func (c *DefaultCache) SetFlags(flags ...CacheFlags) {
	c.flags = CacheFlagsNone
	for _, f := range flags {
		c.flags |= f
	}
}

func (c *DefaultCache) GetUser(userID Snowflake) (User, bool) {
	return c.usersCache.Get(userID)
}

func (c *DefaultCache) GetGuild(guildID Snowflake) (Guild, bool) {
	return c.guildsCache.Get(guildID)
}

func (c *DefaultCache) GetMember(guildID, userID Snowflake) (Member, bool) {
	return c.membersCache.Get(SnowflakePairKey{A: guildID, B: userID})
}

func (c *DefaultCache) GetChannel(channelID Snowflake) (Channel, bool) {
	return c.channelsCache.Get(channelID)
}

func (c *DefaultCache) GetMessage(messageID Snowflake) (Message, bool) {
	return c.messagesCache.Get(messageID)
}

func (c *DefaultCache) GetVoiceState(guildID, userID Snowflake) (VoiceState, bool) {
	return c.voiceStates.Get(SnowflakePairKey{A: guildID, B: userID})
}

func (c *DefaultCache) GetGuildChannels(guildID Snowflake) (map[Snowflake]GuildChannel, bool) {
	set, ok := c.guildToChannelIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]GuildChannel, len(set))
	for channelID := range set {
		if channel, exists := c.channelsCache.Get(channelID); exists {
			if gc, ok := channel.(GuildChannel); ok {
				res[channelID] = gc
			}
		}
	}
	return res, true
}

func (c *DefaultCache) GetGuildMembers(guildID Snowflake) (map[Snowflake]Member, bool) {
	set, ok := c.guildToMemberIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]Member, len(set))
	for userID := range set {
		key := SnowflakePairKey{A: guildID, B: userID}
		if member, exists := c.membersCache.Get(key); exists {
			res[userID] = member
		}
	}
	return res, true
}

func (c *DefaultCache) GetGuildVoiceStates(guildID Snowflake) (map[Snowflake]VoiceState, bool) {
	set, ok := c.guildToVoiceStateUserIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]VoiceState, len(set))
	for userID := range set {
		key := SnowflakePairKey{A: guildID, B: userID}
		if voiceState, exists := c.voiceStates.Get(key); exists {
			res[userID] = voiceState
		}
	}
	return res, true
}

func (c *DefaultCache) GetGuildRoles(guildID Snowflake) (map[Snowflake]Role, bool) {
	set, ok := c.guildToRoleIDs.Get(guildID)
	if !ok {
		return nil, false
	}
	res := make(map[Snowflake]Role, len(set))
	for roleID := range set {
		if role, exists := c.rolesCache.Get(roleID); exists {
			res[roleID] = role
		}
	}
	return res, true
}

func (c *DefaultCache) HasUser(userID Snowflake) bool {
	if !c.flags.Has(CacheFlagUsers) {
		return false
	}
	return c.usersCache.Has(userID)
}

func (c *DefaultCache) HasGuild(guildID Snowflake) bool {
	if !c.flags.Has(CacheFlagGuilds) {
		return false
	}
	return c.guildsCache.Has(guildID)
}

func (c *DefaultCache) HasMember(guildID, userID Snowflake) bool {
	if !c.flags.Has(CacheFlagMembers) {
		return false
	}
	return c.membersCache.Has(SnowflakePairKey{A: guildID, B: userID})
}

func (c *DefaultCache) HasChannel(channelID Snowflake) bool {
	if !c.flags.Has(CacheFlagChannels) {
		return false
	}
	return c.channelsCache.Has(channelID)
}

func (c *DefaultCache) HasMessage(messageID Snowflake) bool {
	if !c.flags.Has(CacheFlagMessages) {
		return false
	}
	return c.messagesCache.Has(messageID)
}

func (c *DefaultCache) HasVoiceState(guildID, userID Snowflake) bool {
	if !c.flags.Has(CacheFlagVoiceStates) {
		return false
	}
	return c.voiceStates.Has(SnowflakePairKey{A: guildID, B: userID})
}

func (c *DefaultCache) HasGuildChannels(guildID Snowflake) bool {
	if !c.flags.Has(CacheFlagChannels) {
		return false
	}
	return c.guildToChannelIDs.Has(guildID)
}

func (c *DefaultCache) HasGuildMembers(guildID Snowflake) bool {
	if !c.flags.Has(CacheFlagMembers) {
		return false
	}
	return c.guildToMemberIDs.Has(guildID)
}

func (c *DefaultCache) HasGuildVoiceStates(guildID Snowflake) bool {
	if !c.flags.Has(CacheFlagVoiceStates) {
		return false
	}
	return c.guildToVoiceStateUserIDs.Has(guildID)
}

func (c *DefaultCache) HasGuildRoles(guildID Snowflake) bool {
	if !c.flags.Has(CacheFlagRoles) {
		return false
	}
	return c.guildToRoleIDs.Has(guildID)
}

func (c *DefaultCache) CountUsers() int {
	return c.usersCache.Len()
}

func (c *DefaultCache) CountGuilds() int {
	return c.guildsCache.Len()
}

func (c *DefaultCache) CountMembers() int {
	return c.membersCache.Len()
}

func (c *DefaultCache) CountChannels() int {
	return c.channelsCache.Len()
}

func (c *DefaultCache) CountMessages() int {
	return c.messagesCache.Len()
}

func (c *DefaultCache) CountVoiceStates() int {
	return c.voiceStates.Len()
}

func (c *DefaultCache) CountRoles() int {
	return c.rolesCache.Len()
}

func (c *DefaultCache) CountGuildChannels(guildID Snowflake) int {
	return c.guildToChannelIDs.Count(guildID)
}

func (c *DefaultCache) CountGuildMembers(guildID Snowflake) int {
	return c.guildToMemberIDs.Count(guildID)
}

func (c *DefaultCache) CountGuildRoles(guildID Snowflake) int {
	return c.guildToRoleIDs.Count(guildID)
}

func (c *DefaultCache) PutUser(user User) {
	if !c.flags.Has(CacheFlagUsers) {
		return
	}
	c.usersCache.Set(user.ID, user)
}

func (c *DefaultCache) PutGuild(guild Guild) {
	if !c.flags.Has(CacheFlagGuilds) {
		return
	}
	c.guildsCache.Set(guild.ID, guild)
}

func (c *DefaultCache) PutMember(member Member) {
	if !c.flags.Has(CacheFlagMembers) {
		return
	}
	userID := member.User.ID
	guildID := member.GuildID
	key := SnowflakePairKey{A: guildID, B: userID}
	c.membersCache.Set(key, member)
	c.guildToMemberIDs.Add(guildID, userID)
}

func (c *DefaultCache) PutChannel(channel Channel) {
	if !c.flags.Has(CacheFlagChannels) {
		return
	}
	channelID := channel.GetID()
	c.channelsCache.Set(channelID, channel)
	if guildChannel, ok := channel.(GuildChannel); ok {
		guildID := guildChannel.GetGuildID()
		c.guildToChannelIDs.Add(guildID, channelID)
	}
}

func (c *DefaultCache) PutMessage(message Message) {
	if !c.flags.Has(CacheFlagMessages) {
		return
	}
	c.messagesCache.Set(message.ID, message)
}

func (c *DefaultCache) PutVoiceState(voiceState VoiceState) {
	if !c.flags.Has(CacheFlagVoiceStates) {
		return
	}
	guildID := voiceState.GuildID
	userID := voiceState.UserID
	key := SnowflakePairKey{A: guildID, B: userID}
	c.voiceStates.Set(key, voiceState)
	c.guildToVoiceStateUserIDs.Add(guildID, userID)
}

func (c *DefaultCache) PutRole(role Role) {
	if !c.flags.Has(CacheFlagRoles) {
		return
	}
	guildID := role.GuildID
	roleID := role.ID
	c.rolesCache.Set(roleID, role)
	c.guildToRoleIDs.Add(guildID, roleID)
}

func (c *DefaultCache) DelUser(userID Snowflake) bool {
	return c.usersCache.Delete(userID)
}

func (c *DefaultCache) DelGuild(guildID Snowflake) bool {
	return c.guildsCache.Delete(guildID)
}

func (c *DefaultCache) DelMember(guildID, userID Snowflake) bool {
	key := SnowflakePairKey{A: guildID, B: userID}
	ok := c.membersCache.Delete(key)
	if ok {
		c.guildToMemberIDs.Remove(guildID, userID)
	}
	return ok
}

func (c *DefaultCache) DelChannel(channelID Snowflake) bool {
	channel, ok := c.channelsCache.Get(channelID)
	if !ok {
		return false
	}
	c.channelsCache.Delete(channelID)
	if guildChannel, ok := channel.(GuildChannel); ok {
		c.guildToChannelIDs.Remove(guildChannel.GetGuildID(), channelID)
	}
	return true
}

func (c *DefaultCache) DelMessage(messageID Snowflake) bool {
	return c.messagesCache.Delete(messageID)
}

func (c *DefaultCache) DelVoiceState(guildID, userID Snowflake) bool {
	key := SnowflakePairKey{A: guildID, B: userID}
	ok := c.voiceStates.Delete(key)
	if ok {
		c.guildToVoiceStateUserIDs.Remove(guildID, userID)
	}
	return ok
}

func (c *DefaultCache) DelRole(guildID, roleID Snowflake) bool {
	ok := c.rolesCache.Delete(roleID)
	if ok {
		c.guildToRoleIDs.Remove(guildID, roleID)
	}
	return ok
}

func (c *DefaultCache) DelGuildChannels(guildID Snowflake) bool {
	set, ok := c.guildToChannelIDs.Delete(guildID)
	if !ok {
		return false
	}
	for channelID := range set {
		c.channelsCache.Delete(channelID)
	}
	return true
}

func (c *DefaultCache) DelGuildMembers(guildID Snowflake) bool {
	set, ok := c.guildToMemberIDs.Delete(guildID)
	if !ok {
		return false
	}
	for userID := range set {
		key := SnowflakePairKey{A: guildID, B: userID}
		c.membersCache.Delete(key)
	}
	return true
}
