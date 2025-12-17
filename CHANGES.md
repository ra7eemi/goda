
  1. Replaced sonic with encoding/json

  - Replaced all github.com/bytedance/sonic imports with encoding/json
  - Replaced all sonic.Marshal → json.Marshal
  - Replaced all sonic.Unmarshal → json.Unmarshal
  - Fixed duplicate import declarations in multiple files

  2. Updated go.mod

  - Removed github.com/bytedance/sonic and all its indirect dependencies
  - Only remaining dependencies are github.com/gobwas/ws (WebSocket library)

  3. Channel Messaging Methods (from original task)

  Added action methods to all channel types in channel.go:
  - TextChannel: Send, SendWith, SendEmbed, FetchMessages, FetchMessage, BulkDelete, Delete, Edit, Guild
  - VoiceChannel: Send, SendWith, SendEmbed, FetchMessages, FetchMessage, Delete, Edit, Guild
  - AnnouncementChannel: All messaging methods + BulkDelete
  - StageVoiceChannel: All messaging methods
  - ThreadChannel: All messaging methods + BulkDelete
  - ForumChannel/MediaChannel: Delete, Edit, Guild (no direct messaging)
  - DMChannel/GroupDMChannel: Send, SendWith, SendEmbed, FetchMessages, FetchMessage

  4. Fixed Pre-existing Issues

  - Fixed Guild() methods return type to match CacheManager.GetGuild signature
  - Added missing error types (ErrChannelNotVoice, ErrChannelNotStage, etc.)
  - Added missing types (ImageFile, ApplicationCommandOptionChoice, StringMap)
  - Fixed ApplicationCommand interface usage in REST API methods
  - Removed duplicate ChannelCreateOptions and PermissionOverwrite definitions
