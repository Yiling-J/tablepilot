"use client";

import {
    getProviders,
    createProvider,
    updateProvider,
    deleteProvider,
    Model as ModelDataFromActionModel,
    Provider as ProviderDataFromAction,
} from "@/actions";
import { Skeleton } from "@/components/ui/skeleton"; // Added
import { Card, CardHeader, CardContent } from "@/components/ui/card"; // Added for Skeleton
import { ConfirmationDialog } from "@/components/models/confirmation-dialog";
import { ModelFormDialog } from "@/components/models/model-form-dialog";
import { ProviderCard } from "@/components/models/provider-card";
import { ProviderFormDialog } from "@/components/models/provider-form-dialog";
import { useToast } from "@/hooks/use-toast";
import type { ModelData, ProviderData } from "@/types.ts";
import { Search } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

interface ModelManagerProps {
  searchTerm: string;
  shouldOpenAddProviderDialog?: boolean;
  onAddProviderDialogDismiss?: () => void;
}

export function ModelManager({
  searchTerm,
  shouldOpenAddProviderDialog,
  onAddProviderDialogDismiss,
}: ModelManagerProps) {
  const { toast } = useToast();
  const [providers, setProviders] = useState<ProviderData[]>([]);
  const [isLoading, setIsLoading] = useState(true); // Added loading state

  // Dialog states
  const [isProviderFormOpen, setIsProviderFormOpen] = useState(false);
  const [editingProvider, setEditingProvider] = useState<ProviderData | null>(
    null,
  );

  const [isModelFormOpen, setIsModelFormOpen] = useState(false);
  const [editingModel, setEditingModel] = useState<ModelData | null>(null);
  const [currentProviderForModel, setCurrentProviderForModel] =
    useState<ProviderData | null>(null);
  const [isConfirmDeleteDialogOpen, setIsConfirmDeleteDialogOpen] =
    useState(false);
  const [deleteAction, setDeleteAction] = useState<{
    type: "provider" | "model";
    id: string;
    providerId?: string;
  } | null>(null);

  const fetchData = async () => {
    setIsLoading(true); // Set loading true at the start
    try {
      const fetchedProviders: ProviderDataFromAction[] = await getProviders();
      const mappedProviders: ProviderData[] = fetchedProviders.map((p) => ({
        id: p.id.toString(),
        name: p.name,
        type: p.type,
        apiKey: p.key,
        baseUrl: p.base_url,
        models: p.models.map((m: ModelDataFromActionModel) => ({ // Removed index i
          id: m.id.toString(), // Use actual model ID
          model: m.model,
          alias: m.alias,
          max_tokens: m.max_tokens,
          rpm: m.rpm,
          imageSupport: m.image,
          isDefault: false, // Default status might need to be fetched or managed
          client: p.name,
        })),
        enabled: true, // Assuming enabled by default, or fetch this status
        editable: p.editable,
      }));
      setProviders(mappedProviders);
    } catch (error) {
      console.error("Failed to fetch providers:", error);
      toast({
        variant: "destructive",
        title: "Error Fetching Providers",
        description:
          "Could not load provider configurations. Please try again later.",
      });
    } finally {
      setIsLoading(false); // Set loading false in finally block
    }
  };

  useEffect(() => {
    fetchData();
  }, []); // Removed toast from dependency array, fetchData will call it.

  // Handle externally triggered dialogs
  useEffect(() => {
    if (shouldOpenAddProviderDialog) {
      setEditingProvider(null); // Reset editing state for new provider
      setIsProviderFormOpen(true);
      onAddProviderDialogDismiss?.();
    }
  }, [shouldOpenAddProviderDialog, onAddProviderDialogDismiss]);

  // CRUD Operations for Providers
  const mapProviderDataToApiParams = (providerData: ProviderData, isUpdate: boolean = false): ProviderDataFromAction => {
    // Ensure this returns a full Provider object as expected by actions.ts
    // ProviderDataFromAction is an alias for Provider from actions.ts
    return {
      id: isUpdate && providerData.id !== "new" ? parseInt(providerData.id, 10) : 0, // API Provider.id is number. Default to 0 for create.
      name: providerData.name,
      type: providerData.type,
      key: providerData.apiKey ?? "", // Ensure non-nullable string
      base_url: providerData.baseUrl ?? "", // Ensure non-nullable string
      models: providerData.models.map(model => ({ // Map UI ModelData to API Model
        id: parseInt(model.id, 10), // API Model.id is number
        model: model.model,
        alias: model.alias,
        max_tokens: model.max_tokens ?? 0, // Default if undefined
        rpm: model.rpm ?? 0,             // Default if undefined
        image: model.imageSupport ?? false, // Default if undefined
      })),
      editable: providerData.editable ?? true, // Default 'editable' if not present in UI data
    };
  };

  // mapApiProviderToActionProvider was removed as it's no longer used after changes to handleProviderSubmit

  const handleProviderSubmit = async (providerData: ProviderData) => {
    try {
      if (editingProvider && editingProvider.id !== "new") { // Update existing provider
        const providerIdToUpdate = editingProvider.id; 
        const apiParams = mapProviderDataToApiParams(providerData, true);
        // updateProvider returns TableInfo, not Provider. Rely on fetchData to update state.
        await updateProvider(providerIdToUpdate, apiParams); 
        toast({
          title: "Provider Updated",
          description: `${providerData.name} has been successfully updated.`, // Use providerData for name
        });
      } else { // Create new provider
        const apiParams = mapProviderDataToApiParams(providerData, false);
        // createProvider returns TableInfo, not Provider. Rely on fetchData to update state.
        await createProvider(apiParams);
        toast({
          title: "Provider Created",
          description: `${providerData.name} has been successfully created.`, // Use providerData for name
        });
      }
      setEditingProvider(null);
      await fetchData(); 
    } catch (error) {
      console.error("Failed to save provider:", error);
      // It's good practice to also call fetchData() in catch blocks if a refresh might resolve/clarify UI state
      // await fetchData(); // Or not, depending on desired UX for errors.
      toast({
        variant: "destructive",
        title: `Error ${editingProvider ? "Updating" : "Creating"} Provider`,
        description: (error as Error).message || "Could not save provider. Please try again.",
      });
    }
  };

  const handleDeleteProvider = async (providerId: string) => {
    try {
      await deleteProvider(providerId); // Pass ID as string
      // Optimistically update UI, then refresh from server
      setProviders((prev) => prev.filter((p) => p.id !== providerId));
      toast({
        title: "Provider Deleted",
        description: "The provider has been removed.", // Simplified message
      });
      await fetchData(); 
    } catch (error) {
      console.error("Failed to delete provider:", error);
      // await fetchData(); // Again, consider if refresh is needed on error
      toast({
        variant: "destructive",
        title: "Error Deleting Provider",
        description: (error as Error).message || "Could not delete provider. Please try again.",
      });
    }
  };

  const handleToggleProviderEnabled = (providerId: string) => {
    setProviders((prev) =>
      prev.map((p) =>
        p.id === providerId ? { ...p, enabled: !p.enabled } : p,
      ),
    );
    const provider = providers.find((p) => p.id === providerId); // Find from current state before update
    if (provider) {
      toast({
        title: `Provider ${!provider.enabled ? "Enabled" : "Disabled"}`,
        description: `${provider.name} has been ${!provider.enabled ? "enabled" : "disabled"}.`,
      });
    }
  };

  const handleModelSubmit = (model: ModelData) => {
    if (!currentProviderForModel) return;
    setProviders((prev) =>
      prev.map((p) => {
        if (p.id === currentProviderForModel.id) {
          const existingModelIndex = p.models.findIndex(
            (m: ModelData) => m.id === model.id,
          );
          let newModels;
          if (existingModelIndex > -1) {
            newModels = [...p.models];
            newModels[existingModelIndex] = model;
          } else {
            newModels = [...p.models, model];
          }
          if (model.isDefault) {
            newModels = newModels.map((m: ModelData) => ({
              ...m,
              isDefault: m.id === model.id,
            }));
          }
          return { ...p, models: newModels };
        }
        return p;
      }),
    );
    setEditingModel(null);
    setCurrentProviderForModel(null);
  };

  const handleDeleteModel = (providerId: string, modelId: string) => {
    setProviders((prev) =>
      prev.map((p) => {
        if (p.id === providerId) {
          return {
            ...p,
            models: p.models.filter((m: ModelData) => m.id !== modelId),
          };
        }
        return p;
      }),
    );
    toast({
      title: "Model Deleted",
      description: "The model has been removed from the provider.",
    });
  };

  // Dialog Triggers for internal use (e.g., clicking edit on a card)
  const openEditProviderDialog = (providerId: string) => {
    const provider = providers.find((p) => p.id === providerId);
    if (provider) {
      setEditingProvider(provider);
      setIsProviderFormOpen(true);
    }
  };

  const openAddModelDialog = (providerId: string) => {
    const provider = providers.find((p) => p.id === providerId);
    if (provider && provider.enabled) {
      setCurrentProviderForModel(provider);
      setEditingModel(null);
      setIsModelFormOpen(true);
    } else if (provider && !provider.enabled) {
      toast({
        variant: "destructive",
        title: "Provider Disabled",
        description: "Cannot add models to a disabled provider.",
      });
    }
  };

  const openEditModelDialog = (providerId: string, modelId: string) => {
    const provider = providers.find((p) => p.id === providerId);
    const model = provider?.models.find((m) => m.id === modelId);
    if (provider && model && provider.enabled) {
      setCurrentProviderForModel(provider);
      setEditingModel(model);
      setIsModelFormOpen(true);
    } else if (provider && !provider.enabled) {
      toast({
        variant: "destructive",
        title: "Provider Disabled",
        description: "Cannot edit models of a disabled provider.",
      });
    }
  };

  const openConfirmDeleteDialog = (
    type: "provider" | "model",
    id: string,
    providerId?: string,
  ) => {
    const provider = providerId
      ? providers.find((p) => p.id === providerId)
      : providers.find((p) => p.id === id);
    if (
      type === "provider" &&
      provider &&
      !provider.enabled &&
      provider.editable
    ) {
      // Allow deleting disabled providers
    } else if (type === "model" && providerId) {
      const targetProvider = providers.find((p) => p.id === providerId);
      if (
        targetProvider &&
        !targetProvider.enabled &&
        targetProvider.editable
      ) {
        toast({
          variant: "destructive",
          title: "Provider Disabled",
          description: `Cannot delete a model from disabled provider ${targetProvider.name}.`,
        });
        return;
      }
    } else if (
      provider &&
      !provider.enabled &&
      provider.editable &&
      type === "provider"
    ) {
      // This is fine, already covered above.
    } else if (
      provider &&
      !provider.enabled &&
      provider.editable &&
      type === "model"
    ) {
      toast({
        variant: "destructive",
        title: "Provider Disabled",
        description: `Cannot delete a model from a disabled provider. Enable ${provider.name} first.`,
      });
      return;
    }

    setDeleteAction({ type, id, providerId });
    setIsConfirmDeleteDialogOpen(true);
  };

  const executeDelete = () => {
    if (!deleteAction) return;
    if (deleteAction.type === "provider") {
      handleDeleteProvider(deleteAction.id);
    } else if (deleteAction.type === "model" && deleteAction.providerId) {
      handleDeleteModel(deleteAction.providerId, deleteAction.id);
    }
    setDeleteAction(null);
    setIsConfirmDeleteDialogOpen(false);
  };

  const filteredProviders = useMemo(() => {
    if (!searchTerm.trim()) return providers;
    const lowerSearchTerm = searchTerm.toLowerCase();
    return providers
      .filter(
        (provider) =>
          provider.name.toLowerCase().includes(lowerSearchTerm) ||
          provider.type.toLowerCase().includes(lowerSearchTerm) ||
          provider.models.some(
            (model: ModelData) =>
              model.model.toLowerCase().includes(lowerSearchTerm) ||
              model.alias.toLowerCase().includes(lowerSearchTerm),
          ),
      )
      .map((provider) => ({
        ...provider,
        models: provider.models.filter(
          (model: ModelData) =>
            provider.name.toLowerCase().includes(lowerSearchTerm) ||
            provider.type.toLowerCase().includes(lowerSearchTerm) ||
            model.model.toLowerCase().includes(lowerSearchTerm) ||
            model.alias.toLowerCase().includes(lowerSearchTerm),
        ),
      }));
  }, [providers, searchTerm]);

  // Function to open the "Add Provider" dialog internally, used by parent via prop
  const openAddProviderDialogInternal = () => {
    setEditingProvider(null);
    setIsProviderFormOpen(true);
  };

  // Effect to handle prop changes for opening dialogs
  useEffect(() => {
    if (shouldOpenAddProviderDialog) {
      openAddProviderDialogInternal();
      onAddProviderDialogDismiss?.();
    }
  }, [shouldOpenAddProviderDialog, onAddProviderDialogDismiss]);


  // Skeleton Component for ProviderCard
  const ProviderCardSkeleton = () => (
    <Card className="mb-6">
      <CardHeader className="flex flex-row items-center justify-between py-4 px-6"> {/* Adjusted padding */}
        <div>
          <Skeleton className="h-6 w-32 mb-2" /> {/* Provider Name - increased margin */}
          <Skeleton className="h-4 w-24" />      {/* Provider Type */}
        </div>
        <div className="flex space-x-2">
          <Skeleton className="h-9 w-20 rounded-md" /> {/* Edit button */}
          <Skeleton className="h-9 w-9 rounded-md" /> {/* More actions button */}
        </div>
      </CardHeader>
      <CardContent className="px-6 pb-6"> {/* Adjusted padding */}
        <div className="flex justify-between items-center mb-3"> {/* Adjusted margin */}
          <Skeleton className="h-5 w-28" /> {/* "Models" title */}
          <Skeleton className="h-9 w-32 rounded-md" /> {/* Add Model button */}
        </div>
        {/* Placeholder for a few models */}
        {[1, 2].map((i) => (
          <div key={i} className="p-3 border rounded-md mb-3 bg-background"> {/* Use theme background */}
            <div className="flex justify-between items-center mb-2"> {/* Increased margin */}
              <Skeleton className="h-5 w-4/12" /> {/* Model Name/Alias */}
              <Skeleton className="h-8 w-8 rounded-md" /> {/* Edit model button */}
            </div>
            <div className="space-y-1.5"> {/* Adjusted for spacing between lines */}
              <Skeleton className="h-3 w-10/12" />   {/* Model property line 1 */}
              <Skeleton className="h-3 w-8/12" />    {/* Model property line 2 */}
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );

  if (isLoading) {
    return (
      <div>
        <ProviderCardSkeleton />
        <ProviderCardSkeleton />
        <ProviderCardSkeleton />
      </div>
    );
  }

  return (
    <>
      {filteredProviders.length > 0 ? (
        <div>
          {filteredProviders.map((provider) => (
            <ProviderCard
              key={provider.id}
              provider={provider}
              onAddModel={openAddModelDialog}
              onEditModel={openEditModelDialog}
              onDeleteModel={(providerId, modelId) =>
                openConfirmDeleteDialog("model", modelId, providerId)
              }
              onEditProvider={openEditProviderDialog}
              onDeleteProvider={(providerId) =>
                openConfirmDeleteDialog("provider", providerId)
              }
              onToggleEnabled={handleToggleProviderEnabled}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center text-center py-16 text-muted-foreground min-h-[calc(100vh-200px)]">
          <Search className="h-24 w-24 mb-6 text-primary/30" />
          <h2 className="text-3xl font-semibold text-primary-foreground mb-2">
            No Providers Found
          </h2>
          <p className="mb-6 max-w-md">
            {searchTerm
              ? `Your search for "${searchTerm}" did not match any providers or models. Try a different search term.`
              : "You don't have any providers configured yet. If you expect to see some, try refreshing or check your import."}
          </p>
        </div>
      )}

      <ProviderFormDialog
        isOpen={isProviderFormOpen}
        onOpenChange={setIsProviderFormOpen}
        onSubmit={handleProviderSubmit}
        initialData={editingProvider}
      />
      {currentProviderForModel && (
        <ModelFormDialog
          isOpen={isModelFormOpen}
          onOpenChange={setIsModelFormOpen}
          onSubmit={handleModelSubmit}
          providerType={currentProviderForModel.type}
          providerName={currentProviderForModel.name}
          initialData={editingModel}
        />
      )}
      <ConfirmationDialog
        isOpen={isConfirmDeleteDialogOpen}
        onOpenChange={setIsConfirmDeleteDialogOpen}
        onConfirm={executeDelete}
        title={`Confirm Deletion`}
        description={`Are you sure you want to delete this ${deleteAction?.type}? This action cannot be undone.`}
      />
    </>
  );
}
