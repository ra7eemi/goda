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

/*****************************
 * Interaction Endpoints     *
 *****************************/

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

/*****************************
 * Application Command Endpoints *
 *****************************/

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
