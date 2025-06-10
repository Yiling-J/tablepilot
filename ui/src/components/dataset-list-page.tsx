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
  return (
    <div className="grow overflow-auto h-full flex flex-col">
      <ModeToggle hide={true} />
      <TablepilotHeader title="Tablepilot" currentTab="datasets" />
      <ScrollArea className="h-[calc(100vh-120px)]">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8 py-12">
          <div className="tab-content-container">
            <DatasetList />
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}

function DatasetList() {
  const [loading, setLoading] = useState(true);
  const [datasets, setDatasets] = useState<DatasetInfo[]>([]);
  const [editDataset, setEditDataset] = useState<DatasetInfo | undefined>(
    undefined,
  );
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isInfoDialogOpen, setIsInfoDialogOpen] = useState(false);
  const [isPreviewDialogOpen, setIsPreviewDialogOpen] = useState(false);
  const [selectedDatasetId, setSelectedDatasetId] = useState<string | null>(
    null,
  ); // State for selected dataset ID

  const fetchDatasets = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await getDatasets();
      setDatasets(resp.datasets ?? []);
    } catch (error) {
      console.error("Failed to fetch datasets:", error);
      toast({
        title: "Error",
        description: "Failed to fetch datasets. Please try again later.",
        variant: "destructive",
      });
      setDatasets([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDatasets();
  }, [fetchDatasets]);

  const handleCreateDataset = async (payload: {
    name: string;
    description: string;
    type: "list" | "csv" | "image";
    data?: string[];
    files?: File[];
  }) => {
    try {
      const requestPayload = {
        name: payload.name,
        description: payload.description,
        type: payload.type,
        data: payload.data ?? [],
        files: payload.files ?? [],
      };

      await createDataset(requestPayload);

      toast({
        title: "Success",
        description: `Dataset "${payload.name}" created successfully.`,
      });
      setIsCreateDialogOpen(false);
      fetchDatasets();
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
      type: "list" | "csv" | "image";
      data?: string[];
      files?: File[];
    },
  ) => {
    try {
      const requestPayload = {
        name: data.name,
        description: data.description,
        type: data.type,
        data: data.data ?? [],
        files: data.files ?? [],
      };

      await updateDataset(id, requestPayload);

      toast({
        title: "Success",
        description: `Dataset "${data.name}" updated successfully.`,
      });
      setIsCreateDialogOpen(false);
      fetchDatasets();
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

  const handleOpenInfoDialog = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsInfoDialogOpen(true);
  };

  return (
    <div className="grow overflow-auto h-full flex flex-col">
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
                onEdit={async () => {
                  setEditDataset(dataset);
                  setIsCreateDialogOpen(true);
                }}
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
        <Card className="relative flex flex-col cursor-pointer h-60 min-w-72 border-dashed overflow-hidden">
          <div
            className="flex flex-col items-center justify-center hover:bg-muted-foreground/5 transition-all w-full h-full flex-1 hover:h-[70%] peer"
            onClick={() => {
              setEditDataset(undefined);
              setIsCreateDialogOpen(true);
            }}
          >
            <PlusIcon className="w-5 h-5 mr-2 mb-2" />
            <span>Add New Dataset</span>
          </div>
          <div className="absolute bottom-2 right-2">
            <Button variant="ghost" size="icon" onClick={handleOpenInfoDialog}>
              <QuestionMarkCircledIcon className="h-5 w-5" />
            </Button>
          </div>
        </Card>
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
      <DatasetPreviewDialog
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
