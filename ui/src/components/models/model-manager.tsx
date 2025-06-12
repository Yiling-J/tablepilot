import {
    createProvider,
    deleteProvider,
    getProviders,
    Model,
    Provider,
    updateProvider,
} from "@/actions";
import { ConfirmationDialog } from "@/components/models/confirmation-dialog";
import { ModelFormDialog } from "@/components/models/model-form-dialog";
import { ProviderCard } from "@/components/models/provider-card";
import { ProviderFormDialog } from "@/components/models/provider-form-dialog";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useToast } from "@/hooks/use-toast";
import { PlusCircle, PlusIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Button } from "../ui/button";

export function ModelManager() {
  const { toast } = useToast();
  const [providers, setProviders] = useState<Provider[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Dialog states
  const [isProviderFormOpen, setIsProviderFormOpen] = useState(false);
  const [editingProvider, setEditingProvider] = useState<Provider | null>(null);

  const [isModelFormOpen, setIsModelFormOpen] = useState(false);
  const [editingModel, setEditingModel] = useState<Model | null>(null);
  const [currentProviderForModel, setCurrentProviderForModel] =
    useState<Provider | null>(null);

  // State for Provider Delete Confirmation Dialog
  const [isProviderDeleteConfirmOpen, setIsProviderDeleteConfirmOpen] =
    useState(false);
  const [providerToDeleteId, setProviderToDeleteId] = useState<string | null>(
    null,
  );

  const currentModelIndex = useRef<number | null>(null);

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const providers = await getProviders();
      setProviders(providers);
    } catch (error) {
      console.error("Failed to fetch providers:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleProviderSubmit = async (providerData: Provider) => {
    try {
      if (editingProvider && editingProvider.id !== 0) {
        // Update existing provider
        await updateProvider(editingProvider.id.toString(), providerData);
        toast({
          title: "Provider Updated",
          description: `${providerData.name} has been successfully updated.`,
        });
      } else {
        await createProvider(providerData);
        toast({
          title: "Provider Created",
          description: `${providerData.name} has been successfully created.`,
        });
      }
      setEditingProvider(null);
      await fetchData();
    } catch (error) {
      console.error("Failed to save provider:", error);
    }
  };

  const handleDeleteProvider = async (providerId: string) => {
    try {
      await deleteProvider(providerId);
      setProviders((prev) =>
        prev.filter((p) => p.id.toString() !== providerId),
      );
      toast({
        title: "Provider Deleted",
        description: "The provider has been removed.",
      });
      await fetchData();
    } catch (error) {
      console.error("Failed to delete provider:", error);
      toast({
        variant: "destructive",
        title: "Error Deleting Provider",
        description:
          (error as Error).message ||
          "Could not delete provider. Please try again.",
      });
    }
  };

  const handleToggleProviderEnabled = async (providerId: string) => {
    const updatedProviders = await Promise.all(
      providers.map(async (p) => {
        if (p.id.toString() === providerId) {
          const updated = { ...p, enabled: !p.enabled };
          await updateProvider(updated.id.toString(), updated);
          return updated;
        }
        return p;
      }),
    );

    setProviders(updatedProviders);

    const updatedProvider = updatedProviders.find(
      (p) => p.id.toString() === providerId,
    );
    if (updatedProvider) {
      toast({
        title: `Provider ${updatedProvider.enabled ? "Enabled" : "Disabled"}`,
        description: `${updatedProvider.name} has been ${updatedProvider.enabled ? "enabled" : "disabled"}.`,
      });
    }
  };

  const handleModelSubmit = async (model: Model) => {
    if (!currentProviderForModel) return;

    const updatedProviders = await Promise.all(
      providers.map(async (p) => {
        if (p.id === currentProviderForModel.id) {
          let newModels;
          if (currentModelIndex.current !== null) {
            newModels = [...p.models];
            newModels[currentModelIndex.current] = model;
          } else {
            newModels = [...p.models, model];
          }
          currentModelIndex.current = null;
          const updated = { ...p, models: newModels };
          await updateProvider(updated.id.toString(), updated);
          return updated;
        }
        return p;
      }),
    );

    setProviders(updatedProviders);
    setEditingModel(null);
    setCurrentProviderForModel(null);
  };

  const handleDeleteModel = async (providerId: string, modelId: string) => {
    const updatedProviders = await Promise.all(
      providers.map(async (p) => {
        if (p.id.toString() === providerId) {
          const updated = {
            ...p,
            models: p.models.filter((m: Model) => m.model !== modelId),
          };
          await updateProvider(updated.id.toString(), updated);
          return updated;
        }
        return p;
      }),
    );

    setProviders(updatedProviders);

    toast({
      title: "Model Deleted",
      description: "The model has been removed from the provider.",
    });
  };

  const openEditProviderDialog = (providerId: string) => {
    const provider = providers.find((p) => p.id.toString() === providerId);
    if (provider) {
      setEditingProvider(provider);
      setIsProviderFormOpen(true);
    }
  };

  const openAddModelDialog = (providerId: string) => {
    const provider = providers.find((p) => p.id.toString() === providerId);
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
    const provider = providers.find((p) => p.id.toString() === providerId);
    const model = provider?.models.find((m, i) => {
      if (m.model === modelId) {
        currentModelIndex.current = i;
      }
      return m.model === modelId;
    });
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
      ? providers.find((p) => p.id.toString() === providerId)
      : providers.find((p) => p.id.toString() === id);
    if (
      type === "provider" &&
      provider &&
      !provider.enabled &&
      provider.editable
    ) {
      // Allow deleting disabled providers
    } else if (type === "model" && providerId) {
      const targetProvider = providers.find(
        (p) => p.id.toString() === providerId,
      );
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

    // Only handle provider type here. Model deletion will use its own dialog.
    if (type === "provider") {
      setProviderToDeleteId(id);
      setIsProviderDeleteConfirmOpen(true);
    }
    // No longer setting deleteAction for models here
    // setIsConfirmDeleteDialogOpen(true); // This was for the shared dialog
  };

  const executeProviderDelete = () => {
    if (providerToDeleteId) {
      handleDeleteProvider(providerToDeleteId);
    }
    setProviderToDeleteId(null);
    setIsProviderDeleteConfirmOpen(false);
  };

  // Function to open the "Add Provider" dialog internally, used by parent via prop
  const openAddProviderDialogInternal = () => {
    setEditingProvider(null);
    setIsProviderFormOpen(true);
  };

  // Skeleton Component for ProviderCard
  const ProviderCardSkeleton = () => (
    <Card className="mb-6">
      <CardHeader className="flex flex-row items-center justify-between py-4 px-6">
        <div>
          <Skeleton className="h-6 w-32 mb-2" />{" "}
          <Skeleton className="h-4 w-24" />
        </div>
        <div className="flex space-x-2">
          <Skeleton className="h-9 w-20 rounded-md" />
          <Skeleton className="h-9 w-9 rounded-md" />
        </div>
      </CardHeader>
      <CardContent className="px-6 pb-6">
        <div className="flex justify-between items-center mb-3">
          <Skeleton className="h-5 w-28" />
          <Skeleton className="h-9 w-32 rounded-md" />
        </div>
        {[1, 2].map((i) => (
          <div key={i} className="p-3 border rounded-md mb-3 bg-background">
            <div className="flex justify-between items-center mb-2">
              <Skeleton className="h-5 w-4/12" />
              <Skeleton className="h-8 w-8 rounded-md" />{" "}
            </div>
            <div className="space-y-1.5">
              <Skeleton className="h-3 w-10/12" />
              <Skeleton className="h-3 w-8/12" />
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

  if (providers.length === 0) {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-200px)]">
        <ProviderFormDialog
          isOpen={isProviderFormOpen}
          onOpenChange={setIsProviderFormOpen}
          onSubmit={handleProviderSubmit}
          initialData={editingProvider}
        />
        <div
          className="border border-muted rounded-2xl p-12 bg-transparent text-muted-foreground flex flex-col items-center gap-3 hover:bg-muted-foreground/5 cursor-pointer"
          onClick={() => {
            openAddProviderDialogInternal();
          }}
          aria-label="Add provider"
        >
          <PlusCircle className="w-8 h-8 mb-4" />
          <p>Create a new provider</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="bg-background sticky top-0 z-10 py-4 border-b">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 flex justify-end">
          <Button
            variant="outline"
            onClick={() => {
              openAddProviderDialogInternal();
            }}
            // aria-label="Add new provider" // Removed to allow name to be derived from text content
          >
            <PlusIcon className="h-4 w-4 mr-2" />
            Add New Provider
          </Button>
        </div>
      </div>
      <div className="flex-grow overflow-auto">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
          {providers.map((provider, i) => (
            <ProviderCard
            key={provider.id === 0 ? `pv_${i.toString()}` : provider.id}
            provider={provider}
            onAddModel={openAddModelDialog}
            onEditModel={openEditModelDialog}
            onDeleteModel={handleDeleteModel}
            onEditProvider={openEditProviderDialog}
            onDeleteProvider={(providerId) =>
              openConfirmDeleteDialog("provider", providerId)
            }
            onToggleEnabled={handleToggleProviderEnabled}
          />
        ))}
        </div> {/* Closes max-w-6xl div */}
      </div> {/* Closes flex-grow overflow-auto div */}

      {/* Dialogs should be outside the scrollable content div, but inside the main component div */}
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
      {/* Provider Deletion Confirmation Dialog (existing) */}
      <ConfirmationDialog
        isOpen={isProviderDeleteConfirmOpen}
        onOpenChange={setIsProviderDeleteConfirmOpen}
        onConfirm={executeProviderDelete}
        title={`Confirm Provider Deletion`}
        description={`Are you sure you want to delete this provider? This action cannot be undone.`}
      />
      {/* Removed the two stray closing divs from here, dialogs are now correctly placed */}
    </div>
  );
}
