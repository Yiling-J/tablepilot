import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    CheckCircle2,
    Edit3,
    PenOff,
    Power,
    PowerOff,
    Settings2,
    Trash2,
    XCircle,
} from "lucide-react";
import { AddModelCard } from "./add-model-card";

import { Model, Provider } from "@/actions";
import { CommonCard } from "@/components/ui/common-card";
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

  if (!provider.models) {
    provider.models = [];
  }

  return (
    <div
      className={`provider-card w-full py-6 border-b border-border/50 ${interactionsDisabled && provider.editable ? "opacity-60" : ""}`}
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
              <CommonCard
                key={model.model}
                name={model.alias || model.model}
                onEdit={
                  provider.editable && provider.enabled
                    ? () => onEditModel(provider.id.toString(), model.model)
                    : undefined
                }
                onDelete={
                  provider.editable && provider.enabled
                    ? () => onDeleteModel(provider.id.toString(), model.model)
                    : undefined
                }
              >
                <div className="flex-grow space-y-3 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">Max Tokens:</span>
                    <span className="font-medium text-card-foreground">
                      {(model.max_tokens ?? 6000) > 0
                        ? (model.max_tokens ?? 0)
                        : 6000}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">RPM Limit:</span>
                    <span className="font-medium text-card-foreground">
                      {(model.rpm ?? 0) > 0
                        ? (model.rpm ?? 0).toLocaleString()
                        : "N/A"}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">
                      Image Generation Support:
                    </span>
                    {model.image ? (
                      <CheckCircle2 className="h-5 w-5 text-green-500" />
                    ) : (
                      <XCircle className="h-5 w-5 text-red-500" />
                    )}
                  </div>
                </div>
              </CommonCard>
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
