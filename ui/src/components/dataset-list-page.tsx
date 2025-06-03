import { createDataset, DatasetInfo, getDatasets } from "@/actions"; // Added createDataset
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardFooter,
    CardHeader,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "@/hooks/use-toast"; // Import toast
import { PlusIcon } from "@radix-ui/react-icons";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ModeToggle } from "./darkmode";
import { CreateDatasetDialog } from "./dialog/dataset/dataset"; // Corrected Import path
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
  const navigate = useNavigate();
  const [datasets, setDatasets] = useState<DatasetInfo[]>([]);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false); // State for dialog

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
  }, []); // Removed toast from dependency array as it's a stable function

  useEffect(() => {
    fetchDatasets();
  }, [fetchDatasets]); // Added fetchDatasets to dependency array

  const handleCreateDataset = async (data: {
    name: string;
    description: string;
    type: "list" | "csv";
    options?: string[];
    files?: File[];
  }) => {
    try {
      // The createDataset action might expect FormData for file uploads,
      // or it might handle a plain object and construct FormData internally.
      // Assuming it handles a plain object for now, as per its likely signature from actions.ts
      // If it strictly requires FormData, this part needs adjustment.

      // Prepare the request object for createDataset, aligning with CreateDatasetRequest
      const requestPayload = {
        name: data.name,
        description: data.description,
        type: data.type,
        data: data.type === "list" ? data.options || [] : [],
        files: data.type === "csv" ? data.files || [] : [],
      };

      await createDataset(requestPayload); // createDataset returns Promise<string> (the ID)

      toast({
        title: "Success",
        description: `Dataset "${data.name}" created successfully.`, // Use data.name from input
      });
      setIsCreateDialogOpen(false); // Close dialog on success
      fetchDatasets(); // Refresh the list, which will include the new dataset
    } catch (error: any) {
      console.error("Failed to create dataset:", error);
      toast({
        title: "Error Creating Dataset",
        description:
          error.message || "Failed to create dataset. Please try again.",
        variant: "destructive",
      });
      // Dialog remains open for user to correct or retry if needed, or close manually.
    }
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
              <div
                key={dataset.id}
                className="h-60 flex flex-col rounded-lg bg-background p-4 border border-gray-400/30 hover:bg-muted-foreground/5 cursor-pointer"
                onClick={() => navigate(`/datasets/${dataset.id}`)}
              >
                <div className="text-xl font-bold truncate">{dataset.name}</div>
                <div className="grow mt-2">
                  <p className="line-clamp-4">{dataset.description}</p>
                </div>

                <div className="self-end">
                  <Button
                    variant="destructive"
                    onClick={async (e) => {
                      e.stopPropagation();
                      toast({
                        title: "Delete clicked (not implemented)",
                        description: `Dataset: ${dataset.name}`,
                      });
                    }}
                  >
                    Delete
                  </Button>
                </div>
              </div>
            ))}
        <Card className="flex flex-col cursor-pointer h-60 min-w-72 border-dashed overflow-hidden">
          <div
            className="flex flex-col items-center justify-center hover:bg-muted-foreground/5 transition-all w-full h-full flex-1 hover:h-[70%] peer"
            onClick={() => setIsCreateDialogOpen(true)} // Corrected onClick
          >
            <PlusIcon className="w-5 h-5 mr-2 mb-2" />
            <span>Add New Dataset</span>
          </div>
        </Card>
      </div>
      <CreateDatasetDialog
        isOpen={isCreateDialogOpen}
        onClose={() => setIsCreateDialogOpen(false)}
        onCreate={handleCreateDataset}
      />
    </div>
  );
}
