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
    <div className="grow overflow-auto h-full flex flex-col">
      <ModeToggle hide={true} />
      <TablepilotHeader title="Tablepilot" currentTab="workflows" />
      <ScrollArea className="h-[calc(100vh-120px)]">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8 py-12">
          <div className="tab-content-container">
            <div className="max-w-6xl grid grid-cols-1 sm:grid-cols-2 md:grid-cols-2 lg:grid-cols-3 gap-6">
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
              <Card className="flex flex-col cursor-pointer h-60 min-w-72 border-dashed overflow-hidden">
                <div
                  className="flex flex-col items-center justify-center hover:bg-muted-foreground/5 transition-all w-full h-full flex-1"
                  onClick={() => {
                    setWorkflow(undefined);
                    setRunWorkflowBuilderOpen(true);
                  }}
                >
                  <PlusIcon className="w-5 h-5 mr-2 mb-2" />
                  <span>Add New Workflow</span>
                </div>
              </Card>
            </div>
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}
