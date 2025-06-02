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
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardFooter,
    CardHeader,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PlusIcon } from "@radix-ui/react-icons";
import { SettingsIcon, Trash2 } from "lucide-react";
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
                    <Card
                      key={wf.id}
                      className="h-60 flex flex-col rounded-lg bg-background border border-gray-400/30 hover:bg-muted-foreground/5 cursor-pointer"
                    >
                      <div
                        className="flex flex-col flex-grow p-4 cursor-pointer"
                        onClick={async () => {
                          const w = await getWorkflow(wf.id);
                          setWorkflow(w);
                          setRunWorkflowOpen(true);
                        }}
                      >
                        <CardHeader className="p-0 pb-2"> {/* Adjusted padding */}
                          <div className="text-lg font-semibold truncate"> {/* Adjusted style */}
                            {wf.name}
                          </div>
                        </CardHeader>
                        <CardContent className="grow mt-2 p-0">
                          <p className="line-clamp-4">{wf.description}</p>
                        </CardContent>
                      </div>
                      <CardFooter className="px-4 py-3 border-t border-gray-400/30 flex justify-end gap-2"> {/* Adjusted padding */}
                        <Button
                          variant="ghost"
                          size="icon"
                          title="Settings"
                          onClick={async (e) => {
                            e.stopPropagation();
                            const w = await getWorkflow(wf.id);
                            setWorkflow(w);
                            setRunWorkflowBuilderOpen(true);
                          }}
                        >
                          <SettingsIcon className="h-4 w-4" />
                        </Button>
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              title="Delete Workflow"
                              className="text-destructive hover:text-destructive hover:bg-destructive/10"
                              onClick={(e) => e.stopPropagation()} // Prevent navigation
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent onClick={(e) => e.stopPropagation()}>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                              <AlertDialogDescription>
                                This action cannot be undone. This will permanently delete the workflow.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction
                                onClick={async () => {
                                  await deleteWorkflow(wf.id);
                                  await refreshWorkflows();
                                }}
                              >
                                Delete
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </CardFooter>
                    </Card>
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
