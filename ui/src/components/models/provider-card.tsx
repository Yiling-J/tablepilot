import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Edit3,
    PenOff,
    Power,
    PowerOff,
    Settings2,
    Trash2,
} from "lucide-react";
import { AddModelCard } from "./add-model-card";
import { ModelCard } from "./model-card";

import { Model, Provider } from "@/actions";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

interface ProviderCardProps {
  provider: Provider;
  onAddModel: (providerId: string) => void;
  onEditModel: (providerId: string, modelId: string) => void;
  onDeleteModel: (providerId: string, modelId: string) => void;
  onEditProvider: (providerId: string) => void;
  onDeleteProvider: (providerId: string) => void;
  onToggleEnabled: (providerId: string) => void;
}

export function ProviderCard({
  provider,
  onAddModel,
  onEditModel,
  onDeleteModel,
  onEditProvider,
  onDeleteProvider,
  onToggleEnabled,
}: ProviderCardProps) {
  const interactionsDisabled = provider.editable && !provider.enabled;

  return (
    <div
      className={`w-full py-6 border-b border-border/50 ${interactionsDisabled && provider.editable ? "opacity-60" : ""}`}
    >
      {/* Header Section */}
      <div className="flex justify-between items-center gap-2 mb-4 px-2">
        <div className="flex-grow min-w-0">
          <h2 className="text-xl font-bold text-foreground border-l-4 border-blue-500 pl-3">
            {provider.name}
          </h2>
          <div className="flex flex-row flex-wrap items-baseline gap-x-1.5 gap-y-0 mt-2">
            <span className="text-sm text-muted-foreground whitespace-nowrap">
              type: {provider.type}
            </span>
            {provider.type === "OpenAI-compatible" && provider.base_url && (
              <span className="text-sm text-muted-foreground break-all ml-1">
                base_url: {provider.base_url}
              </span>
            )}
          </div>
        </div>
        {/* Right part: Read-only badge and Enable/Disable Switch */}
        <div className="flex flex-col items-end gap-2 flex-shrink-0">
          {!provider.editable && (
            <Badge
              variant="outline"
              className="text-xs flex items-center gap-1"
            >
              <PenOff className="h-3 w-3" /> Read-only
            </Badge>
          )}
          {provider.editable && (
            <div className="flex items-center space-x-2">
              {provider.enabled ? (
                <Power className="h-4 w-4 text-green-500" />
              ) : (
                <PowerOff className="h-4 w-4 text-red-500" />
              )}
              <Label
                htmlFor={`enable-provider-${provider.id}`}
                className="text-xs sr-only"
              >
                {provider.enabled ? "Disable" : "Enable"} Provider
              </Label>
              <Switch
                id={`enable-provider-${provider.id}`}
                checked={provider.enabled}
                onCheckedChange={() => onToggleEnabled(provider.id.toString())}
                aria-label={`${provider.enabled ? "Disable" : "Enable"} provider ${provider.name}`}
              />
            </div>
          )}
        </div>
      </div>

      <div className="px-2 mb-4">
        {provider.models.length > 0 || provider.editable ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {provider.models.map((model: Model) => (
              <ModelCard
                key={model.model}
                model={model}
                onEdit={() => onEditModel(provider.id.toString(), model.model)}
                onDelete={() =>
                  onDeleteModel(provider.id.toString(), model.model)
                }
                isProviderEditable={provider.editable}
                isProviderEnabled={provider.enabled}
              />
            ))}
            {provider.editable && (
              <AddModelCard
                onClick={() => onAddModel(provider.id.toString())}
                disabled={!provider.enabled}
              />
            )}
          </div>
        ) : (
          <div className="text-center py-8 text-muted-foreground">
            <Settings2 className="mx-auto h-12 w-12 mb-2" />
            <p className="font-semibold">No models yet.</p>
            {!provider.editable && (
              <p className="text-sm">
                This provider is read-only and has no models.
              </p>
            )}
          </div>
        )}
      </div>

      {/* Footer Section (Buttons) */}
      {provider.editable && (
        <div className="px-2 flex flex-col sm:flex-row justify-end items-center gap-2">
          <div className="flex gap-2">
            <Button
              onClick={() => onEditProvider(provider.id.toString())}
              variant="outline"
              size="sm"
              disabled={interactionsDisabled}
            >
              <Edit3 className="mr-2 h-4 w-4" /> Edit Provider
            </Button>
            <Button
              onClick={() => onDeleteProvider(provider.id.toString())}
              variant="destructive"
              size="sm"
              disabled={interactionsDisabled}
            >
              <Trash2 className="mr-2 h-4 w-4" /> Delete Provider
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
