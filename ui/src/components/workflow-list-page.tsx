import {
    deleteWorkflow,
    getWorkflow,
    getWorkflows,
    Workflow,
    WorkflowInfo,
} from "@/actions";
import WorkflowBuilderDialog from "@/components/dialog/workflow/builder";
import WorkflowExecutionDialog from "@/components/dialog/workflow/workflow";
import {
    Card,
    CardContent,
    CardFooter,
    CardHeader,
} from "@/components/ui/card";
import { CommonCard } from "@/components/ui/common-card";
import { Skeleton } from "@/components/ui/skeleton";
import { PlusIcon } from "@radix-ui/react-icons";
import { useCallback, useEffect, useState } from "react";
import { ModeToggle } from "./darkmode";
import { TablepilotHeader } from "./header";
import { ScrollArea } from "./ui/scroll-area";
import { Button } from "./ui/button"; // Import Button

export function WorkflowListPage() {
  const [workflows, setWorkflows] = useState<WorkflowInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [workflow, setWorkflow] = useState<undefined | Workflow>(undefined);
  const [runWorkflowOpen, setRunWorkflowOpen] = useState(false);
  const [WorkflowBuilderOpen, setRunWorkflowBuilderOpen] = useState(false);

  const refreshWorkflows = useCallback(async () => {
    setLoading(true);
    try {
      const wf = await getWorkflows();
      setWorkflows(wf.workflows ?? []); // Ensure workflows is an array
    } catch (error) {
      console.error("Failed to fetch workflows:", error);
      setWorkflows([]); // Set to empty array on error
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refreshWorkflows();
  }, [refreshWorkflows]);

  return (
    <div className="grow h-full flex flex-col">
      <ModeToggle hide={true} />
      <TablepilotHeader title="Tablepilot" currentTab="workflows" />
      <div className="bg-background sticky top-0 z-10 py-4 border-b">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 flex justify-end">
          <Button
            variant="outline"
            onClick={() => {
              setWorkflow(undefined);
              setRunWorkflowBuilderOpen(true);
            }}
          >
            <PlusIcon className="w-4 h-4 mr-2" />
            Add New Workflow
          </Button>
        </div>
      </div>
      <WorkflowExecutionDialog
        workflow={workflow}
        open={runWorkflowOpen}
        onOpenChange={setRunWorkflowOpen}
      />
      <WorkflowBuilderDialog
        id={workflow?.id}
        workflow={workflow}
        open={WorkflowBuilderOpen}
        onOpenChange={setRunWorkflowBuilderOpen}
        onSave={refreshWorkflows}
      />
      <ScrollArea className="flex-grow">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
          <div className="tab-content-container">
            <div className="max-w-6xl grid grid-cols-1 sm:grid-cols-2 md:grid-cols-2 lg:grid-cols-3 gap-6 pt-6">
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
                : workflows.map((wf) => (
                    <CommonCard
                      key={wf.id}
                      name={wf.name}
                      onClick={async () => {
                        const w = await getWorkflow(wf.id);
                        setWorkflow(w);
                        setRunWorkflowOpen(true);
                      }}
                      onEdit={async () => {
                        const w = await getWorkflow(wf.id);
                        setWorkflow(w);
                        setRunWorkflowBuilderOpen(true);
                      }}
                      onDelete={async () => {
                        await deleteWorkflow(wf.id);
                        await refreshWorkflows();
                      }}
                    >
                      <p className="line-clamp-4">{wf.description}</p>
                    </CommonCard>
                  ))}
              {/* The Add New Workflow Card has been removed */}
            </div>
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}
