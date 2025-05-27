
'use client';

import { useState, useEffect, useMemo } from 'react';
import { getProviders, Provider as ProviderDataFromAction, Model as ModelDataFromActionModel } from '@/actions'; // Added Model
import type { ProviderData, ModelData } from '@/types.ts';
import { ProviderCard } from '@/components/models/provider-card';
import { ProviderFormDialog } from '@/components/models/provider-form-dialog';
import { ModelFormDialog } from '@/components/models/model-form-dialog';
import { OptimizeConfigDialog } from '@/components/models/optimize-config-dialog';
import { ImportExportDialog } from '@/components/models/import-export-dialog';
import { ConfirmationDialog } from '@/components/models/confirmation-dialog';
// Button removed as it's unused
import { Search } from 'lucide-react'; // PlusCircle removed as it's unused
import { useToast } from '@/hooks/use-toast';
import { v4 as uuidv4 } from 'uuid';

interface ModelManagerProps {
  searchTerm: string;
  // onProvidersChange?: (providers: ProviderData[]) => void; // Removed
  shouldOpenAddProviderDialog?: boolean;
  onAddProviderDialogDismiss?: () => void;
  shouldOpenImportExportDialog?: boolean;
  onImportExportDialogDismiss?: () => void;
}

export function ModelManager({
  searchTerm,
  // onProvidersChange, // Removed
  shouldOpenAddProviderDialog,
  onAddProviderDialogDismiss,
  shouldOpenImportExportDialog,
  onImportExportDialogDismiss,
}: ModelManagerProps) {
  const { toast } = useToast();
  const [providers, setProviders] = useState<ProviderData[]>([]);
  // const [isInitialLoadComplete, setIsInitialLoadComplete] = useState(false); // Removed

  // Dialog states
  const [isProviderFormOpen, setIsProviderFormOpen] = useState(false);
  const [editingProvider, setEditingProvider] = useState<ProviderData | null>(null);

  const [isModelFormOpen, setIsModelFormOpen] = useState(false);
  const [editingModel, setEditingModel] = useState<ModelData | null>(null);
  const [currentProviderForModel, setCurrentProviderForModel] = useState<ProviderData | null>(null);

  const [isOptimizeDialogOpen, setIsOptimizeDialogOpen] = useState(false);
  const [optimizingModelInfo, /* setOptimizingModelInfo */] = useState<{ provider: ProviderData, model: ModelData } | null>(null); // setOptimizingModelInfo removed
  
  const [isImportExportOpen, setIsImportExportOpen] = useState(false);

  const [isConfirmDeleteDialogOpen, setIsConfirmDeleteDialogOpen] = useState(false);
  const [deleteAction, setDeleteAction] = useState<{ type: 'provider' | 'model', id: string, providerId?: string } | null>(null);
  
  // Fetch providers from API
  useEffect(() => {
    const fetchData = async () => {
      try {
        const fetchedProviders: ProviderDataFromAction[] = await getProviders();
        const mappedProviders: ProviderData[] = fetchedProviders.map((p) => ({
          id: String(p.id), // Convert number to string
          name: p.name,
          type: p.type, // Assuming ProviderType and string are compatible
          apiKey: p.key, // Store API key if needed, or omit if not used directly by UI
          baseUrl: p.base_url, // Store baseUrl if needed
          models: p.models.map((m: ModelDataFromActionModel) => ({
            id: uuidv4(), // Generate unique ID for UI model
            model: m.model,
            alias: m.alias || m.model, // Use model name as alias if alias is not provided
            max_tokens: m.max_tokens,
            rpm: m.rpm,
            imageSupport: m.image, // Map 'image' to 'imageSupport'
            isDefault: false, // Default to false, adjust if API provides this
            client: p.name, // Set client to provider name
          })),
          enabled: true, // Default to true as API doesn't provide this
          editable: p.editable,
        }));
        setProviders(mappedProviders);
      } catch (error) {
        console.error("Failed to fetch providers:", error);
        toast({
          variant: "destructive",
          title: "Error Fetching Providers",
          description: "Could not load provider configurations. Please try again later.",
        });
        // Optionally, set providers to an empty array or mock data on error
        // setProviders(mockProviders.map(p => ({...p, enabled: p.enabled === undefined ? true : p.enabled })));
      }
    };
    fetchData();
  }, [toast]); // Added toast to dependency array

  // Handle externally triggered dialogs
  useEffect(() => {
    if (shouldOpenAddProviderDialog) {
      setEditingProvider(null); // Reset editing state for new provider
      setIsProviderFormOpen(true);
      onAddProviderDialogDismiss?.();
    }
  }, [shouldOpenAddProviderDialog, onAddProviderDialogDismiss]);

  useEffect(() => {
    if (shouldOpenImportExportDialog) {
      setIsImportExportOpen(true);
      onImportExportDialogDismiss?.();
    }
  }, [shouldOpenImportExportDialog, onImportExportDialogDismiss]);


  // CRUD Operations for Providers
  const handleProviderSubmit = (provider: ProviderData) => {
    let updatedProvidersList: ProviderData[];
    setProviders(prev => {
      const existingIndex = prev.findIndex(p => p.id === provider.id);
      if (existingIndex > -1) {
        const updated = [...prev];
        updated[existingIndex] = provider;
        updatedProvidersList = updated;
        return updated;
      }
      updatedProvidersList = [...prev, { ...provider, enabled: true, editable: true }]; 
      return updatedProvidersList;
    });
    setEditingProvider(null);
  };

  const handleDeleteProvider = (providerId: string) => {
    setProviders(prev => prev.filter(p => p.id !== providerId));
    toast({ title: "Provider Deleted", description: "The provider and its models have been removed." });
  };

  const handleToggleProviderEnabled = (providerId: string) => {
    setProviders(prev => 
      prev.map(p => 
        p.id === providerId ? { ...p, enabled: !p.enabled } : p
      )
    );
    const provider = providers.find(p => p.id === providerId); // Find from current state before update
    if (provider) { // Check if provider exists before accessing !provider.enabled
        toast({ title: `Provider ${!provider.enabled ? 'Enabled' : 'Disabled'}`, description: `${provider.name} has been ${!provider.enabled ? 'enabled' : 'disabled'}.`});
    }
  };

  // CRUD Operations for Models
  const handleModelSubmit = (model: ModelData) => {
    if (!currentProviderForModel) return;
    setProviders(prev => prev.map(p => {
      if (p.id === currentProviderForModel.id) {
        const existingModelIndex = p.models.findIndex((m: ModelData) => m.id === model.id);
        let newModels;
        if (existingModelIndex > -1) {
          newModels = [...p.models];
          newModels[existingModelIndex] = model;
        } else {
          newModels = [...p.models, model];
        }
        if (model.isDefault) {
          newModels = newModels.map((m: ModelData) => ({ ...m, isDefault: m.id === model.id }));
        }
        return { ...p, models: newModels };
      }
      return p;
    }));
    setEditingModel(null);
    setCurrentProviderForModel(null);
  };

  const handleDeleteModel = (providerId: string, modelId: string) => {
    setProviders(prev => prev.map(p => {
      if (p.id === providerId) {
        return { ...p, models: p.models.filter((m: ModelData) => m.id !== modelId) };
      }
      return p;
    }));
    toast({ title: "Model Deleted", description: "The model has been removed from the provider." });
  };

  // Dialog Triggers for internal use (e.g., clicking edit on a card)
  const openEditProviderDialog = (providerId: string) => {
    const provider = providers.find(p => p.id === providerId);
    if (provider) {
      setEditingProvider(provider);
      setIsProviderFormOpen(true);
    }
  };
  
  const openAddModelDialog = (providerId: string) => {
    const provider = providers.find(p => p.id === providerId);
    if (provider && provider.enabled) {
      setCurrentProviderForModel(provider);
      setEditingModel(null);
      setIsModelFormOpen(true);
    } else if (provider && !provider.enabled) {
        toast({ variant: "destructive", title: "Provider Disabled", description: "Cannot add models to a disabled provider." });
    }
  };

  const openEditModelDialog = (providerId: string, modelId: string) => {
    const provider = providers.find(p => p.id === providerId);
    const model = provider?.models.find(m => m.id === modelId);
    if (provider && model && provider.enabled) {
      setCurrentProviderForModel(provider);
      setEditingModel(model);
      setIsModelFormOpen(true);
    } else if (provider && !provider.enabled) {
        toast({ variant: "destructive", title: "Provider Disabled", description: "Cannot edit models of a disabled provider." });
    }
  };
  
  // openOptimizeDialog function removed as it was unused

  // const handleApplyOptimization = (optimizedValues: { max_tokens: number; rpm: number }) => { // Unused function removed
  //   if (!optimizingModelInfo) return;
  //   const { provider, model } = optimizingModelInfo;
    
  //   const updatedModel = { ...model, ...optimizedValues };
    
  //   setProviders(prev => prev.map(p => {
  //     if (p.id === provider.id) {
  //       return {
  //         ...p,
  //         models: p.models.map((m: ModelData) => m.id === model.id ? updatedModel : m)
  //       };
  //     }
  //     return p;
  //   }));
  //   setOptimizingModelInfo(null);
  // };

  const openConfirmDeleteDialog = (type: 'provider' | 'model', id: string, providerId?: string) => {
    const provider = providerId ? providers.find(p => p.id === providerId) : providers.find(p => p.id === id);
    if (type === 'provider' && provider && !provider.enabled && provider.editable) {
        // Allow deleting disabled providers
    } else if (type === 'model' && providerId) {
        const targetProvider = providers.find(p => p.id === providerId);
        if (targetProvider && !targetProvider.enabled && targetProvider.editable) {
            toast({ variant: "destructive", title: "Provider Disabled", description: `Cannot delete a model from disabled provider ${targetProvider.name}.` });
            return;
        }
    } else if (provider && !provider.enabled && provider.editable && type === 'provider') {
        // This is fine, already covered above.
    } else if (provider && !provider.enabled && provider.editable && type === 'model') {
       toast({ variant: "destructive", title: "Provider Disabled", description: `Cannot delete a model from a disabled provider. Enable ${provider.name} first.` });
       return;
    }


    setDeleteAction({ type, id, providerId });
    setIsConfirmDeleteDialogOpen(true);
  };

  const executeDelete = () => {
    if (!deleteAction) return;
    if (deleteAction.type === 'provider') {
      handleDeleteProvider(deleteAction.id);
    } else if (deleteAction.type === 'model' && deleteAction.providerId) {
      handleDeleteModel(deleteAction.providerId, deleteAction.id);
    }
    setDeleteAction(null);
    setIsConfirmDeleteDialogOpen(false);
  };

  const handleImportConfig = (config: ProviderData[]) => {
    const validatedConfig = config.map(p => ({
      ...p,
      enabled: p.enabled === undefined ? true : p.enabled,
      editable: p.editable === undefined ? true : p.editable,
      models: p.models || [],
    }));
    setProviders(validatedConfig); 
    toast({ title: "Configuration Imported", description: "Configuration has been successfully imported." });
  };

  const filteredProviders = useMemo(() => {
    if (!searchTerm.trim()) return providers;
    const lowerSearchTerm = searchTerm.toLowerCase();
    return providers.filter(provider =>
      provider.name.toLowerCase().includes(lowerSearchTerm) ||
      provider.type.toLowerCase().includes(lowerSearchTerm) ||
      provider.models.some((model: ModelData) =>
        model.model.toLowerCase().includes(lowerSearchTerm) ||
        model.alias.toLowerCase().includes(lowerSearchTerm)
      )
    ).map(provider => ({
        ...provider,
        models: provider.models.filter((model: ModelData) =>
            provider.name.toLowerCase().includes(lowerSearchTerm) || 
            provider.type.toLowerCase().includes(lowerSearchTerm) || 
            model.model.toLowerCase().includes(lowerSearchTerm) ||
            model.alias.toLowerCase().includes(lowerSearchTerm)
        )
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

  useEffect(() => {
    if (shouldOpenImportExportDialog) {
      setIsImportExportOpen(true);
      onImportExportDialogDismiss?.();
    }
  }, [shouldOpenImportExportDialog, onImportExportDialogDismiss]);


  return (
    <>
      {filteredProviders.length > 0 ? (
        <div> {/* Removed space-y-8, ProviderCard now handles its own spacing/separation */}
          {filteredProviders.map((provider) => (
            <ProviderCard
              key={provider.id}
              provider={provider}
              onAddModel={openAddModelDialog}
              onEditModel={openEditModelDialog}
              onDeleteModel={(providerId, modelId) => openConfirmDeleteDialog('model', modelId, providerId)}
              onEditProvider={openEditProviderDialog}
              onDeleteProvider={(providerId) => openConfirmDeleteDialog('provider', providerId)}
              onToggleEnabled={handleToggleProviderEnabled}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center text-center py-16 text-muted-foreground min-h-[calc(100vh-200px)]">
           <Search className="h-24 w-24 mb-6 text-primary/30" />
          <h2 className="text-3xl font-semibold text-primary-foreground mb-2">No Providers Found</h2>
          <p className="mb-6 max-w-md">
            {searchTerm ? `Your search for "${searchTerm}" did not match any providers or models. Try a different search term.` : "You don't have any providers configured yet. If you expect to see some, try refreshing or check your import."}
          </p>
          { /* Button to add provider is now in AppHeader, controlled by parent */ }
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
      {optimizingModelInfo && (
        <OptimizeConfigDialog
            isOpen={isOptimizeDialogOpen}
            onOpenChange={setIsOptimizeDialogOpen}
            // onApplyOptimization={handleApplyOptimization} // Prop removed
            providerType={optimizingModelInfo.provider.type}
            model={optimizingModelInfo.model}
        />
      )}
      <ImportExportDialog
        isOpen={isImportExportOpen}
        onOpenChange={setIsImportExportOpen}
        currentConfig={providers}
        onImport={handleImportConfig}
      />
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

