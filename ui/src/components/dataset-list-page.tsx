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
import { Input } from "./ui/input.tsx";
import { TablepilotHeader } from "./header.tsx";
import { ScrollArea } from "./ui/scroll-area.tsx";

export function DatasetListPage() {
  const [editDataset, setEditDataset] = useState<DatasetInfo | undefined>(
    undefined,
  );
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isInfoDialogOpen, setIsInfoDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const fetchDatasetsCallback = useCallback(async () => {
    // This callback is passed to DatasetList to trigger a data refresh in the parent
    // (e.g., after a create/update operation initiated from a dialog owned by DatasetListPage).
    // The actual state update for the list displayed happens in DatasetList itself.
    // setLoading for DatasetListPage itself is not managed here, as DatasetList handles its own loading state.
    try {
      await getDatasets(); // Calling this primarily to ensure any broader app state reliant on this fetch is updated.
    } catch (error) {
      console.error("Failed to fetch datasets in parent callback:", error);
      toast({
        title: "Error",
        description:
          "Failed to refresh datasets after operation. Please try again later.",
        variant: "destructive",
      });
    }
    // setLoading(false); // setLoading for DatasetListPage is not managed by this callback.
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
      fetchDatasetsCallback();
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
      fetchDatasetsCallback();
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
      <div className="bg-background sticky top-0 z-10 pt-4 pb-1">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 flex justify-between items-center space-x-4">
          <Input
            type="text"
            placeholder="Search datasets..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="max-w-sm h-9"
          />
          <div className="flex space-x-2">
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
              searchQuery={searchQuery}
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
  searchQuery: string;
}

function DatasetList({
  fetchDatasets,
  onEditDataset,
  searchQuery,
}: DatasetListProps) {
  const [loading, setLoading] = useState(true);
  const [datasets, setDatasets] = useState<DatasetInfo[]>([]);
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
      setDatasets([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDatasetsInternal();
  }, [fetchDatasetsInternal]);

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
          : datasets
              .filter((dataset) =>
                dataset.name.toLowerCase().includes(searchQuery.toLowerCase()),
              )
              .map((dataset) => (
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
                  await fetchDatasets();
                  setLoading(false);
                }}
                badgeText={dataset.type}
              >
                <p className="line-clamp-4">{dataset.description}</p>
              </CommonCard>
            ))}
      </div>
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
