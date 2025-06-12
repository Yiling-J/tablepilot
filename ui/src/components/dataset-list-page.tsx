import {
    createDataset,
    DatasetInfo,
    deleteDataset,
    getDatasets,
    updateDataset,
} from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardFooter,
    CardHeader,
} from "@/components/ui/card";
import { CommonCard } from "@/components/ui/common-card";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "@/hooks/use-toast"; // Import toast
import { PlusIcon, QuestionMarkCircledIcon } from "@radix-ui/react-icons";
import { useCallback, useEffect, useState } from "react";
import { ModeToggle } from "./darkmode";
import { CreateDatasetDialog } from "./dialog/dataset/dataset";
import { DatasetInfoDialog } from "./dialog/dataset/info";
import { DatasetPreviewDialog } from "./dialog/dataset/preview";
import { TablepilotHeader } from "./header.tsx";
import { ScrollArea } from "./ui/scroll-area.tsx";

export function DatasetListPage() {
  const [editDataset, setEditDataset] = useState<DatasetInfo | undefined>(
    undefined,
  );
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isInfoDialogOpen, setIsInfoDialogOpen] = useState(false);
  // Retain fetchDatasets and datasets state here if CreateDatasetDialog needs to refresh list,
  // or pass refresh function down from a higher context if available.
  // For now, let's assume DatasetList's fetchDatasets is sufficient or dialogs handle their own data.
  // const [datasets, setDatasets] = useState<DatasetInfo[]>([]); // Keep datasets for dialogs if needed // Removed as 'datasets' is unused

  const fetchDatasetsCallback = useCallback(async () => {
    // This function will be passed to DatasetList to refresh its data
    // setLoading(true); // setLoading will be handled by DatasetList
    try {
      // const resp = await getDatasets(); // Call to getDatasets might be redundant if child list also fetches
      // setDatasets(resp.datasets ?? []); // Update datasets in parent as well for dialogs // Removed as 'setDatasets' is no longer defined
      // For now, just ensure the callback can be called.
      // A more thorough review might be needed if the parent truly needs to manage this state
      // or if this fetch is redundant with the child component's fetching.
      // Triggering a re-fetch in the child is handled by the prop.
      // If additional actions are needed in parent after child re-fetches, that's a different pattern.
      await getDatasets(); // Still calling this to ensure it behaves as before, minus setting state
    } catch (error) {
      console.error("Failed to fetch datasets in parent (callback):", error); // Clarified error source
      toast({
        title: "Error",
        description:
          "Failed to fetch datasets after operation. Please try again later.",
        variant: "destructive",
      });
    } finally {
      // setLoading(false); // setLoading will be handled by DatasetList
    }
  }, []);

  const handleOpenEditDialog = (dataset: DatasetInfo) => {
    setEditDataset(dataset);
    setIsCreateDialogOpen(true);
  };

  const handleCreateDataset = async (data: {
    name: string;
    description: string;
    type: "list" | "csv";
    options?: string[];
    files?: File[];
  }) => {
    try {
      const requestPayload = {
        name: data.name,
        description: data.description,
        type: data.type,
        data: data.type === "list" ? data.options || [] : [],
        files: data.type === "csv" ? data.files || [] : [],
        private: false,
      };
      await createDataset(requestPayload);
      toast({
        title: "Success",
        description: `Dataset "${data.name}" created successfully.`,
      });
      setIsCreateDialogOpen(false);
      fetchDatasetsCallback(); // Refresh datasets
    } catch (error: unknown) {
      toast({
        title: "Error Creating Dataset",
        description:
          (error instanceof Error ? error.message : undefined) ||
          "Failed to create dataset. Please try again.",
        variant: "destructive",
      });
    }
  };

  const handleUpdateDataset = async (
    id: string,
    data: {
      name: string;
      description: string;
      type: "list" | "csv";
      options?: string[];
      files?: File[];
    },
  ) => {
    try {
      const requestPayload = {
        name: data.name,
        description: data.description,
        type: data.type,
        data: data.type === "list" ? data.options || [] : [],
        files: data.type === "csv" ? data.files || [] : [],
        private: false,
      };
      await updateDataset(id, requestPayload);
      toast({
        title: "Success",
        description: `Dataset "${data.name}" updated successfully.`,
      });
      setIsCreateDialogOpen(false);
      fetchDatasetsCallback(); // Refresh datasets
    } catch (error: unknown) {
      toast({
        title: "Error Updating Dataset",
        description:
          (error instanceof Error ? error.message : undefined) ||
          "Failed to update dataset. Please try again.",
        variant: "destructive",
      });
    }
  };

  return (
    <div className="grow h-full flex flex-col">
      <ModeToggle hide={true} />
      <TablepilotHeader title="Tablepilot" currentTab="datasets" />
      <div className="bg-background sticky top-0 z-10 pt-4 pb-2 border-b">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 flex justify-end space-x-2">
          <Button
            variant="outline"
            onClick={() => {
              setEditDataset(undefined);
              setIsCreateDialogOpen(true);
            }}
          >
            <PlusIcon className="w-4 h-4 mr-2" />
            Add New Dataset
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={() => setIsInfoDialogOpen(true)}
          >
            <QuestionMarkCircledIcon className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <CreateDatasetDialog
        isOpen={isCreateDialogOpen}
        onClose={() => setIsCreateDialogOpen(false)}
        onCreate={handleCreateDataset}
        onUpdate={handleUpdateDataset}
        dataset={editDataset}
      />
      <DatasetInfoDialog
        isOpen={isInfoDialogOpen}
        onClose={() => setIsInfoDialogOpen(false)}
      />
      <ScrollArea className="flex-grow">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
          <div className="tab-content-container">
            <DatasetList
              fetchDatasets={fetchDatasetsCallback}
              onEditDataset={handleOpenEditDialog}
            />
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}

interface DatasetListProps {
  fetchDatasets: () => Promise<void>;
  onEditDataset: (dataset: DatasetInfo) => void;
}

function DatasetList({ fetchDatasets, onEditDataset }: DatasetListProps) {
  const [loading, setLoading] = useState(true);
  const [datasets, setDatasets] = useState<DatasetInfo[]>([]);
  // const [editDataset, setEditDataset] = useState<DatasetInfo | undefined>( // Moved to parent
  // undefined,
  // );
  // const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false); // Moved to parent
  // const [isInfoDialogOpen, setIsInfoDialogOpen] = useState(false); // Moved to parent
  const [isPreviewDialogOpen, setIsPreviewDialogOpen] = useState(false);
  const [selectedDatasetId, setSelectedDatasetId] = useState<string | null>(
    null,
  );

  const fetchDatasetsInternal = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await getDatasets();
      setDatasets(resp.datasets ?? []);
    } catch (error) {
      console.error("Failed to fetch datasets in DatasetList:", error);
      // Toast is handled by parent or not needed if parent refreshes
      setDatasets([]);
    } finally {
      setLoading(false);
    }
  }, []); // Removed fetchDatasets from dependency array as it's now an external prop

  useEffect(() => {
    fetchDatasetsInternal();
  }, [fetchDatasetsInternal]); // Depends on the internal fetcher

  // Removed handleCreateDataset, handleUpdateDataset, handleOpenInfoDialog as they are in parent

  return (
    <div className="grow overflow-auto h-full flex flex-col pt-6">
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {loading
          ? Array.from({ length: 4 }).map((_, index) => (
              <Card key={index} className="w-80">
                <CardHeader>
                  <Skeleton className="h-6 w-3/4 mb-2" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-4 w-full mb-2" />
                  <Skeleton className="h-4 w-full" />
                </CardContent>
                <CardFooter>
                  <Skeleton className="h-4 w-1/4" />
                </CardFooter>
              </Card>
            ))
          : datasets.map((dataset) => (
              <CommonCard
                key={dataset.id}
                name={dataset.name}
                onClick={() => {
                  setSelectedDatasetId(dataset.id);
                  setIsPreviewDialogOpen(true);
                }}
                onEdit={() => onEditDataset(dataset)}
                onDelete={async () => {
                  setLoading(true);
                  await deleteDataset(dataset.id);
                  await fetchDatasets(); // This will call the prop to refresh data in parent and here
                  setLoading(false);
                }}
                badgeText={dataset.type}
              >
                <p className="line-clamp-4">{dataset.description}</p>
              </CommonCard>
            ))}
        {/* Removed the Add New Dataset Card and Help button from here */}
      </div>
      {/* Dialogs are now in DatasetListPage */}
      <DatasetPreviewDialog // Keep preview dialog here as it's specific to this list's interaction
        isOpen={isPreviewDialogOpen}
        onClose={() => {
          setIsPreviewDialogOpen(false);
          setSelectedDatasetId(null);
        }}
        datasetId={selectedDatasetId ?? undefined}
      />
    </div>
  );
}
