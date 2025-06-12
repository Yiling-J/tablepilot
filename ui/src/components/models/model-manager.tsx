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
import { useDelayedLoading } from "@/hooks/use-delayed-loading";
import { useToast } from "@/hooks/use-toast";
import { PlusCircle, PlusIcon } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { ModeToggle } from "../darkmode";
import { TablepilotHeader } from "../header";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { ScrollArea } from "../ui/scroll-area";

export function ModelManager() {
  const { toast } = useToast();
  const [providers, setProviders] = useState<Provider[]>([]);
  const [actualIsLoading, setActualIsLoading] = useState(true);
  const isLoading = useDelayedLoading(actualIsLoading, 500); // Using 500ms delay
  const [searchQuery, setSearchQuery] = useState("");

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

  useEffect(() => {
    const fetchData = async () => {
      setActualIsLoading(true);
      try {
        const fetchedProviders = await getProviders();
        setProviders(fetchedProviders);
      } catch (error) {
        console.error("Failed to fetch providers:", error);
      } finally {
        setActualIsLoading(false);
      }
    };
    fetchData();
  }, []);

  const filteredProviders = useMemo(() => {
    if (!searchQuery) {
      return providers.map((p) => ({ ...p, displayModels: p.models }));
    }
    const lowerSearchQuery = searchQuery.toLowerCase();
    return providers
      .map((provider) => {
        const providerNameMatches = provider.name
          .toLowerCase()
          .includes(lowerSearchQuery);
        const matchingModels = provider.models.filter(
          (model) =>
            (model.alias &&
              model.alias.toLowerCase().includes(lowerSearchQuery)) ||
            model.model.toLowerCase().includes(lowerSearchQuery),
        );

        if (providerNameMatches || matchingModels.length > 0) {
          return {
            ...provider,
            displayModels: providerNameMatches
              ? provider.models
              : matchingModels,
          };
        }
        return null;
      })
      .filter((p): p is Provider & { displayModels: Model[] } => p !== null);
  }, [providers, searchQuery]);

  const handleProviderSubmit = async (providerData: Provider) => {
    try {
      if (editingProvider && editingProvider.id !== 0) {
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
      setActualIsLoading(true);
      getProviders()
        .then(setProviders)
        .finally(() => setActualIsLoading(false));
    } catch (error) {
      console.error("Failed to save provider:", error);
    }
  };

  const handleDeleteProvider = async (providerId: string) => {
    try {
      await deleteProvider(providerId);
      toast({
        title: "Provider Deleted",
        description: "The provider has been removed.",
      });
      setActualIsLoading(true);
      getProviders()
        .then(setProviders)
        .finally(() => setActualIsLoading(false));
    } catch (error) {
      console.error("Failed to delete provider:", error);
      toast({
        variant: "destructive",
        title: "Error Deleting Provider",
        description:
          (error instanceof Error ? error.message : "Unknown error") ||
          "Could not delete provider. Please try again.",
      });
    }
  };

  const handleToggleProviderEnabled = async (providerId: string) => {
    const providerToUpdate = providers.find(
      (p) => p.id.toString() === providerId,
    );
    if (!providerToUpdate) return;

    const updatedProviderData = {
      ...providerToUpdate,
      enabled: !providerToUpdate.enabled,
    };

    try {
      await updateProvider(providerId, updatedProviderData);
      setProviders((prev) =>
        prev.map((p) =>
          p.id.toString() === providerId ? updatedProviderData : p,
        ),
      );
      toast({
        title: `Provider ${updatedProviderData.enabled ? "Enabled" : "Disabled"}`,
        description: `${updatedProviderData.name} has been ${updatedProviderData.enabled ? "enabled" : "disabled"}.`,
      });
    } catch (error) {
      console.error("Failed to toggle provider state:", error);
    }
  };

  const handleModelSubmit = async (model: Model) => {
    if (!currentProviderForModel) return;

    let newModels;
    if (currentModelIndex.current !== null) {
      newModels = [...currentProviderForModel.models];
      newModels[currentModelIndex.current] = model;
    } else {
      newModels = [...currentProviderForModel.models, model];
    }

    const updatedProviderData = {
      ...currentProviderForModel,
      models: newModels,
    };

    try {
      await updateProvider(
        currentProviderForModel.id.toString(),
        updatedProviderData,
      );
      setProviders((prev) =>
        prev.map((p) =>
          p.id === currentProviderForModel.id ? updatedProviderData : p,
        ),
      );
      toast({ title: editingModel ? "Model Updated" : "Model Added" });
      setEditingModel(null);
      setCurrentProviderForModel(null);
      currentModelIndex.current = null;
    } catch (error) {
      console.error("Failed to save model:", error);
    }
  };

  const handleDeleteModel = async (providerId: string, modelId: string) => {
    const providerToUpdate = providers.find(
      (p) => p.id.toString() === providerId,
    );
    if (!providerToUpdate) return;

    const newModels = providerToUpdate.models.filter(
      (m: Model) => m.model !== modelId,
    );
    const updatedProviderData = { ...providerToUpdate, models: newModels };

    try {
      await updateProvider(providerId, updatedProviderData);
      setProviders((prev) =>
        prev.map((p) =>
          p.id.toString() === providerId ? updatedProviderData : p,
        ),
      );
      toast({
        title: "Model Deleted",
        description: "The model has been removed from the provider.",
      });
    } catch (error) {
      console.error("Failed to delete model:", error);
    }
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
      currentModelIndex.current = null; // Ensure this is reset for add
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
    const modelIndex = provider?.models.findIndex((m) => m.model === modelId);

    if (
      provider &&
      modelIndex !== undefined &&
      modelIndex !== -1 &&
      provider.enabled
    ) {
      setCurrentProviderForModel(provider);
      setEditingModel(provider.models[modelIndex]);
      currentModelIndex.current = modelIndex;
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

    if (type === "provider") {
      setProviderToDeleteId(id);
      setIsProviderDeleteConfirmOpen(true);
    }
  };

  const executeProviderDelete = () => {
    if (providerToDeleteId) {
      handleDeleteProvider(providerToDeleteId);
    }
    setProviderToDeleteId(null);
    setIsProviderDeleteConfirmOpen(false);
  };

  const openAddProviderDialogInternal = () => {
    setEditingProvider(null);
    setIsProviderFormOpen(true);
  };

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
              <Skeleton className="h-8 w-8 rounded-md" />
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

  if (providers.length === 0 && !searchQuery) {
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
          onClick={openAddProviderDialogInternal}
          aria-label="Add provider"
        >
          <PlusCircle className="w-8 h-8 mb-4" />
          <p>Create a new provider</p>
        </div>
      </div>
    );
  }

  return (
    <div className="grow h-screen flex flex-col">
      <ModeToggle hide={true} />
      <TablepilotHeader title="Tablepilot" currentTab="models" />
      <div className="bg-background sticky top-0 z-10 pt-4 pb-1">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 flex justify-between items-center space-x-4">
          <Input
            type="text"
            placeholder="Search providers or models..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="max-w-sm h-9 rounded-full"
          />
          <div className="flex space-x-2">
            <Button variant="outline" onClick={openAddProviderDialogInternal}>
              <PlusIcon className="h-4 w-4 mr-2" />
              Add New Provider
            </Button>
          </div>
        </div>
      </div>

      <ScrollArea className="flex-grow">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
          {filteredProviders.length === 0 && searchQuery ? (
            <div className="text-center text-muted-foreground py-10">
              No providers or models found matching your search.
            </div>
          ) : (
            filteredProviders.map((provider, i) => (
              <ProviderCard
                key={provider.id === 0 ? `pv_${i.toString()}` : provider.id}
                provider={{ ...provider, models: provider.displayModels }}
                onAddModel={openAddModelDialog}
                onEditModel={openEditModelDialog}
                onDeleteModel={handleDeleteModel}
                onEditProvider={openEditProviderDialog}
                onDeleteProvider={(providerId) =>
                  openConfirmDeleteDialog("provider", providerId)
                }
                onToggleEnabled={handleToggleProviderEnabled}
              />
            ))
          )}
        </div>
      </ScrollArea>

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
        isOpen={isProviderDeleteConfirmOpen}
        onOpenChange={setIsProviderDeleteConfirmOpen}
        onConfirm={executeProviderDelete}
        title={`Confirm Provider Deletion`}
        description={`Are you sure you want to delete this provider? This action cannot be undone.`}
      />
    </div>
  );
}
