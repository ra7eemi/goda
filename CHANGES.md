GODA Single Package Architecture Refactoring Plan
Overview
Refactor GODA (Golang Optimized Discord API) to a "Single Package Architecture" mimicking Discord.js ease of use with discord-interactions-js webhook utilities.
/*******************************************************************************
 *                              USER METHODS
 *******************************************************************************/
Order of sections in merged file:
REST API CORE (existing)
GATEWAY METHODS
USER METHODS
GUILD METHODS
CHANNEL METHODS
MESSAGE METHODS
REACTION METHODS
PIN METHODS
MEMBER METHODS
ROLE METHODS
BAN METHODS
INTERACTION METHODS
APPLICATION COMMAND METHODS
2.1 Create interaction_http.go (NEW FILE)
Contents:
VerifyInteraction function - ed25519 signature verification:

func VerifyInteraction(publicKey, signature, timestamp string, body []byte) bool
Decode hex public key (32 bytes)
Decode hex signature (64 bytes)
Verify: ed25519.Verify(pubKey, timestamp+body, signature)
InteractionServerConfig struct:

type InteractionServerConfig struct {
    PublicKey string  // Hex-encoded public key
    Addr      string  // e.g., ":8080"
    Path      string  // e.g., "/interactions"
}
InteractionServer struct:

type InteractionServer struct {
    config InteractionServerConfig
    client *Client
    server *http.Server
    mux    *http.ServeMux
}
Methods:
NewInteractionServer(client *Client, config InteractionServerConfig) *InteractionServer
Start() error - HTTP server
StartTLS(certFile, keyFile string) error - HTTPS server
Shutdown(ctx context.Context) error
handleInteraction(w http.ResponseWriter, r *http.Request) - Internal handler
Handler logic:
Verify signature with VerifyInteraction()
Auto-respond to PING with PONG
Set client on interactions via SetClient()
Dispatch to handlers
Default deferred response for async handling
2.2 Complete interaction data structures in interaction.go
Add these types (around line 364):
ComponentInteractionData:

type ComponentInteractionData struct {
    CustomID      string                             `json:"custom_id"`
    ComponentType ComponentType                      `json:"component_type"`
    Values        []string                           `json:"values,omitempty"`
    Resolved      *ComponentInteractionResolvedData  `json:"resolved,omitempty"`
}
ComponentInteraction:

type ComponentInteraction struct {
    EntityBase
    ApplicationCommandInteractionFields
    Data    ComponentInteractionData `json:"data"`
    Message *Message                 `json:"message,omitempty"`
}
AutoCompleteInteractionData:

type AutoCompleteInteractionData struct {
    ApplicationCommandInteractionDataFields
    Options []AutoCompleteOption `json:"options"`
}
AutoCompleteInteraction:

type AutoCompleteInteraction struct {
    EntityBase
    ApplicationCommandInteractionFields
    Data AutoCompleteInteractionData `json:"data"`
}
ModalSubmitInteractionData:

type ModalSubmitInteractionData struct {
    CustomID   string                   `json:"custom_id"`
    Components []ModalSubmitComponent   `json:"components"`
}
ModalSubmitInteraction:

type ModalSubmitInteraction struct {
    EntityBase
    ApplicationCommandInteractionFields
    Data ModalSubmitInteractionData `json:"data"`
}
2.3 Add EntityBase to existing interaction types
Modify in interaction.go:
ChatInputCommandInteraction - add EntityBase
UserCommandInteraction - add EntityBase
MessageCommandInteraction - add EntityBase
2.4 Add interaction action methods
Add methods to ChatInputCommandInteraction (and similar for other types):
Reply(content string) error
ReplyWith(data InteractionResponseData) error
ReplyEphemeral(content string) error
ReplyEmbed(embed Embed) error
DeferReply() error
DeferReplyEphemeral(ephemeral bool) error
EditReply(content string) (Message, error)
EditReplyWith(data InteractionResponseData) (Message, error)
DeleteReply() error
Followup(content string) (Message, error)
FollowupWith(data InteractionResponseData) (Message, error)
GetOptionString(name string) (string, bool)
GetOptionInt(name string) (int, bool)
GetOptionBool(name string) (bool, bool)
GetOptionUser(name string) (User, bool)
Phase 3: Top-Level Managers
3.1 Create client_managers.go (NEW FILE)
UserManager:

type UserManager struct { client *Client }
func (m *UserManager) Fetch(userID Snowflake) (User, error)
func (m *UserManager) Get(userID Snowflake) (User, bool)
func (m *UserManager) Me() (User, error)
func (m *UserManager) CreateDM(userID Snowflake) (DMChannel, error)
GuildManager:

type GuildManager struct { client *Client }
func (m *GuildManager) Fetch(guildID Snowflake) (Guild, error)
func (m *GuildManager) Get(guildID Snowflake) (Guild, bool)
func (m *GuildManager) All() []Guild
func (m *GuildManager) Leave(guildID Snowflake) error
ChannelManager:

type ChannelManager struct { client *Client }
func (m *ChannelManager) Fetch(channelID Snowflake) (Channel, error)
func (m *ChannelManager) Get(channelID Snowflake) (Channel, bool)
func (m *ChannelManager) Delete(channelID Snowflake, reason string) error
CommandManager:

type CommandManager struct { client *Client }
func (m *CommandManager) GetGlobal() ([]ApplicationCommand, error)
func (m *CommandManager) CreateGlobal(command ApplicationCommand) (ApplicationCommand, error)
func (m *CommandManager) BulkOverwriteGlobal(commands []ApplicationCommand) ([]ApplicationCommand, error)
func (m *CommandManager) DeleteGlobal(commandID Snowflake) error
func (m *CommandManager) GetGuild(guildID Snowflake) ([]ApplicationCommand, error)
func (m *CommandManager) CreateGuild(guildID Snowflake, command ApplicationCommand) (ApplicationCommand, error)
func (m *CommandManager) BulkOverwriteGuild(guildID Snowflake, commands []ApplicationCommand) ([]ApplicationCommand, error)
func (m *CommandManager) DeleteGuild(guildID, commandID Snowflake) error
3.2 Modify Client struct in client.go
Add fields:

type Client struct {
    // ... existing fields ...
    applicationID Snowflake      // Store for command operations

    // Top-level managers
    Users    *UserManager
    Guilds   *GuildManager
    Channels *ChannelManager
    Commands *CommandManager
}
3.3 Initialize managers in New() function
After dispatcher initialization:

client.Users = &UserManager{client: client}
client.Guilds = &GuildManager{client: client}
client.Channels = &ChannelManager{client: client}
client.Commands = &CommandManager{client: client}
3.4 Store applicationID in Start() function
After fetching self user, store the ID (bot user ID = application ID).
Phase 4: Dispatcher Updates
4.1 Modify dispatcher.go
Add method for HTTP-based interaction dispatch:

func (d *dispatcher) dispatchInteraction(interaction Interaction)
Routes to appropriate handlers based on interaction type.
Implementation Order
Phase 1.1 - Create section structure in restapi.go
Phase 1.1 - Copy each restapi_*.go content into sections
Phase 1.1 - Remove duplicate imports/package declarations
Phase 1.2 - Delete old restapi_*.go files
Verify - go build ./...
Phase 2.2 - Complete interaction data structures
Phase 2.3 - Add EntityBase to command interactions
Phase 2.4 - Add action methods to interactions
Phase 2.1 - Create interaction_http.go
Phase 3.1 - Create client_managers.go
Phase 3.2 - Update Client struct
Phase 3.3 - Initialize managers in New()
Phase 3.4 - Store applicationID in Start()
Phase 4.1 - Add dispatchInteraction to dispatcher
Verify - go test ./...
Files Summary
Action	File
MODIFY	restapi.go - Merge all REST methods
DELETE	restapi_users.go, restapi_channels.go, restapi_messages.go, restapi_members.go, restapi_roles.go, restapi_bans.go, restapi_guilds.go, restapi_interactions.go
CREATE	interaction_http.go - Ed25519 verification + HTTP server
MODIFY	interaction.go - Complete types, add EntityBase, action methods
CREATE	client_managers.go - UserManager, GuildManager, ChannelManager, CommandManager
MODIFY	client.go - Add managers + applicationID
MODIFY	dispatcher.go - Add dispatchInteraction method
Usage After Refactoring

// Discord.js-like syntax
client.Users.Fetch(userID)
client.Guilds.Get(guildID)
client.Commands.CreateGlobal(command)

// Interaction replies
interaction.Reply("Hello!")
interaction.ReplyEphemeral("Only you can see this")
interaction.DeferReply()
interaction.EditReply("Updated response")

// HTTP webhook server
server := goda.NewInteractionServer(client, goda.InteractionServerConfig{
    PublicKey: "your_public_key",
    Addr:      ":8080",
    Path:      "/interactions",
})
server.Start()