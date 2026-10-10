package typesense

import "typesense/api"

// Conversations provides access to conversation model configurations.
type Conversations struct {
	apiClient api.ClientWithResponsesInterface
}

// Models provides conversation model creation and listing.
func (c *Conversations) Models() *ConversationModels {
	return &ConversationModels{apiClient: c.apiClient}
}

// Model selects a conversation model by ID.
func (c *Conversations) Model(id string) *ConversationModel {
	return &ConversationModel{apiClient: c.apiClient, modelID: id}
}
