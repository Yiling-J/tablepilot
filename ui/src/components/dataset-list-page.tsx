import { DatasetInfo, getDatasets } from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardFooter,
    CardHeader,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { FileIcon, PlusIcon } from "@radix-ui/react-icons";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ModeToggle } from "./darkmode";
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

  const fetchDatasets = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await getDatasets();
      setDatasets(resp.datasets ?? []);
    } catch (error) {
      console.error("Failed to fetch tables:", error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDatasets();
  }, []);

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
                  <p className="line-clamp-4">{datasets.description}</p>
                </div>

                <div className="self-end">
                  <Button
                    variant="destructive"
                    onClick={async (e) => {
                      e.stopPropagation();
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
            onClick={() => {}}
          >
            <PlusIcon className="w-5 h-5 mr-2 mb-2" />
            <span>Add New Dataset</span>
          </div>
        </Card>
      </div>
    </div>
  );
}
