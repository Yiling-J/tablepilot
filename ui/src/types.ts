// ui/src/types.ts
export interface ModelData {
  id: string; // Typically UI-generated unique ID (e.g., uuid)
  model: string; // The actual model name/identifier from API
  alias: string; // User-friendly alias, defaults to model name
  max_tokens?: number;
  rpm?: number;
  imageSupport?: boolean; // Whether the model supports images
  isDefault?: boolean; // Whether this is the default model for the provider
  client?: string; // Name of the provider/client
  // Add any other fields that model-card.tsx or model-form-dialog.tsx might expect
}

export interface ProviderData {
  id: string; // UI or API provider ID (stringified if from API)
  name: string; // Provider name
  type: string; // Provider type (e.g., "OpenAI", "Anthropic", "Generic")
  apiKey?: string; // API key, optional if not always needed by UI directly
  baseUrl?: string; // Base URL for generic providers
  models: ModelData[]; // Array of models offered by this provider
  enabled: boolean; // Whether the provider is currently enabled in the UI
  editable: boolean; // Whether the provider configuration is editable by the user
  // Add any other fields that provider-card.tsx or provider-form-dialog.tsx might expect
}

// Define other shared types if they were intended to be in @/types
// For example, if ProviderType was a specific enum/union:
export type ProviderType = "OpenAI" | "Anthropic" | "VertexAI" | "Generic" | string;

// Added based on model-form-dialog.tsx and provider-form-dialog.tsx needs
export interface ModelFormData {
  model: string;
  alias?: string;
  max_tokens: number;
  rpm: number;
  isDefault?: boolean;
  imageSupport: boolean;
}

export interface ProviderFormData {
  name: string;
  type: ProviderType;
  apiKey: string;
  baseUrl?: string;
}

// Added based on optimize-config-dialog.tsx needs
export interface AISuggestionInput {
  providerType: ProviderType;
  usagePatterns: string;
  currentMaxTokens?: number;
  currentRpm?: number;
}

export interface AISuggestionOutput {
  suggestedMaxTokens: number;
  suggestedRpm: number;
  reasoning: string;
}

// Added based on provider-form-dialog.tsx for ProviderTypeOptions
export const ProviderTypeOptions: ProviderType[] = ["OpenAI", "Anthropic", "VertexAI", "Generic"];
