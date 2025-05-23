import {
  deleteWorkflow,
  getWorkflow,
  getWorkflows,
  Workflow,
  WorkflowInfo,
} from "@/actions";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PlusIcon } from "@radix-ui/react-icons";
import { SettingsIcon } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import WorkflowBuilderDialog from "@/components/dialog/workflow/builder";
import WorkflowExecutionDialog from "@/components/dialog/workflow/workflow";

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
    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-2 lg:grid-cols-3 gap-6">
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
            <div
              key={wf.id}
              className="h-60 flex flex-col rounded-lg bg-background p-4 border border-gray-400/30 hover:bg-muted-foreground/5 cursor-pointer"
              onClick={async () => {
                const w = await getWorkflow(wf.id);
                setWorkflow(w);
                setRunWorkflowOpen(true);
              }}
            >
              <div className="text-xl font-bold truncate">{wf.name}</div>
              <div className="grow mt-2">
                <p className="line-clamp-4">{wf.description}</p>
              </div>

              <div className="flex justify-between">
                <Button
                  variant="outline"
                  size="icon"
                  onClick={async (e) => {
                    e.stopPropagation();
                    const w = await getWorkflow(wf.id);
                    setWorkflow(w);
                    setRunWorkflowBuilderOpen(true);
                  }}
                >
                  <SettingsIcon />
                </Button>
                <Button
                  variant="destructive"
                  onClick={async (e) => {
                    e.stopPropagation();
                    await deleteWorkflow(wf.id);
                    await refreshWorkflows(); 
                  }}
                >
                  Delete
                </Button>
              </div>
            </div>
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
  );
}
