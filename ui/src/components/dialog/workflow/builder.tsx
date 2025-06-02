import { ChevronDown, ChevronUp, Pencil, Plus, Trash2 } from "lucide-react";

import { useEffect, useRef, useState } from "react";

import {
    ColumnType,
    CreateColumnStepPayload,
    CreateTableStepPayload,
    ImportDataStepPayload,
    TableInfo,
    TypedWorkflowStep,
    UserInputStepPayload,
    Workflow,
    WorkflowStepType,
    WorkflowVariable,
    createWorkflow,
    getTables,
    tableCreateRequestToTableInfo,
    updateWorkflow,
} from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    // DialogDescription, // Removed as it's no longer used
    DialogFooter,
    DialogHeader,
} from "@/components/ui/dialog";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ContextVariable } from "@/components/ui/var-input";
import { DialogTitle } from "@radix-ui/react-dialog";
import { AutofillStep } from "./steps/autofillStep";
import { CreateColumnStep } from "./steps/createColumnStep";
import { CreateTableStep } from "./steps/createTableStep";
import { DeleteColumnStep } from "./steps/deleteColumnStep";
import { DeleteTableStep } from "./steps/deleteTableStep";
import { ExportTableStep } from "./steps/exportTableStep";
import { GenerateStep } from "./steps/generateStep";
import { ImportStep } from "./steps/importStep";
import { UserInputStep } from "./steps/userInputStep";

export interface StepContext {
  variables: ContextVariable[];
  tables: TableInfo[];
}

export default function WorkflowBuilderDialog({
  id,
  workflow,
  open,
  onOpenChange,
  onSave,
}: {
  id?: string;
  workflow?: Workflow;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: () => Promise<void>;
}) {
  // Workflow state
  const [workflowName, setWorkflowName] = useState<string>("New Workflow");
  const [isEditingName, setIsEditingName] = useState<boolean>(false);
  const [workflowDescription, setWorkflowDescription] = useState<string>(
    "Enter workflow description...",
  );
  const [isEditingDescription, setIsEditingDescription] =
    useState<boolean>(false);
  const [steps, setSteps] = useState<TypedWorkflowStep[]>([]);
  const [selectedStepIndex, setSelectedStepIndex] = useState<number | null>(
    null,
  );
  const [stepContexts, setStepContexts] = useState<StepContext[]>([]);
  // const [createTableDialogOpen, setCreateTableDialogOpen] = useState(false); // Removed state
  const existingTables = useRef<TableInfo[]>([]);

  const fetchTables = async () => {
    const resp = await getTables();
    existingTables.current = resp.tables ?? [];
  };

  const onOpen = async () => {
    await fetchTables();
    if (workflow) {
      setWorkflowName(workflow.name);
      setWorkflowDescription(
        workflow.description || "Enter workflow description...",
      );
      const wsteps: TypedWorkflowStep[] = [
        {
          type: "UserInput",
          payload: {
            variables: workflow.variables,
          },
        },
        ...workflow.steps,
      ];
      setSteps(wsteps);
      buildStepContext(wsteps);
      setSelectedStepIndex(0);
    } else {
      const s = [
        {
          type: "UserInput",
          payload: { variables: [] } as UserInputStepPayload,
        } as TypedWorkflowStep,
      ];
      setSteps(s);
      buildStepContext(s);
      setSelectedStepIndex(0);
    }
  };

  useEffect(() => {
    if (!open) {
      return;
    }
    onOpen();
  }, [open]);

  const buildStepContext = (steps: TypedWorkflowStep[]) => {
    const contexts: StepContext[] = [
      {
        variables: [
          { path: "date", display: "date", type: "string" },
          { path: "time", display: "time", type: "string" },
          { path: "datetime", display: "datetime", type: "string" },
        ],
        tables: [...existingTables.current],
      },
    ];
    steps.forEach((step, index) => {
      const nv: ContextVariable[] = [...contexts[index].variables];
      const tbs: TableInfo[] = [...contexts[index].tables];
      switch (step.type) {
        case "UserInput":
          nv.push(
            ...(step.payload as UserInputStepPayload).variables.map((v) => {
              return {
                display: v.name,
                path: v.name,
                type: v.type,
              } as ContextVariable;
            }),
          );
          break;
        case "CreateTable":
          const tableName = (step.payload as CreateTableStepPayload).request
            .name;
          const ii = tbs.findIndex((t) => t.name === tableName);
          const tr = (step.payload as CreateTableStepPayload).request;
          const ti = tableCreateRequestToTableInfo(tr);
          if (ii === -1) {
            tbs.push(ti);
          } else {
            tbs[ii] = ti;
          }
          break;
        case "CreateColumn":
          const pd = step.payload as CreateColumnStepPayload;
          const t = tbs.find((t) => t.name === pd.table);
          if (t) {
            t.columns.push({
              id: pd.name,
              name: pd.name,
              description: pd.description,
              type: pd.type as ColumnType,
              fill_mode: "ai",
            });
          }
          break;
        case "DeleteColumn":
          break;
        case "Import":
          if (step.payload.name.length > 0 && step.payload.table.length === 0) {
            tbs.push({
              id: step.payload.name,
              name: step.payload.name,
              description: "",
              columns: [],
              model: "",
            });
          }
          break;
      }
      contexts.push({ variables: nv, tables: tbs });
    });
    setStepContexts(contexts);
  };

  // Helper to create a new action
  const createNewStep = (type: WorkflowStepType): TypedWorkflowStep => {
    buildStepContext(steps);
    switch (type) {
      case "UserInput":
        return { type, payload: { variables: [] } };
      case "CreateTable":
        return {
          type,
          payload: {
            on_exists: "Stop",
            request: {
              name: "",
              description: "",
              sources: [],
              columns: [],
            },
          },
        };
      case "DeleteTable":
        return { type, payload: { table: "" } };
      case "CreateColumn":
        return {
          type,
          payload: {
            table: "",
            name: "",
            description: "",
            type: "string",
          },
        };
      case "DeleteColumn":
        return { type, payload: { table: "", column: "" } };
      case "Import":
        return {
          type,
          payload: {
            file: "",
            prompt: "",
            name: "", // Added: Default to empty string
            table: "", // Added: Default to empty string
            truncate: false, // Added: Default to false
          },
        };
      case "Generate":
        return { type, payload: { count: 20, batch: 5, table: "" } };
      case "Autofill":
        return {
          type,
          payload: {
            count: 20,
            batch: 5,
            table: "",
            columns: [],
            context_columns: [],
            prompt: "",
          },
        };
      case "ExportTable":
        return { type, payload: { table: "" } };
      default:
        throw new Error("unknown step type");
    }
  };

  const addStep = (type: WorkflowStepType) => {
    // Prevent adding another UserInput step
    if (type === "UserInput") return;

    const newStep = createNewStep(type);
    setSteps([...steps, newStep]);
    setSelectedStepIndex(steps.length);
  };

  // Remove an action from the workflow
  const removeAction = (index: number) => {
    // Prevent removing the UserInput step
    if (index === 0 && steps[0].type === "UserInput") return;

    const newWorkflow = [...steps];
    newWorkflow.splice(index, 1);
    setSteps(newWorkflow);
    buildStepContext(newWorkflow);

    if (selectedStepIndex === index) {
      setSelectedStepIndex(null);
    } else if (selectedStepIndex !== null && selectedStepIndex > index) {
      setSelectedStepIndex(selectedStepIndex - 1);
    }
  };

  // Move action up in the workflow
  const moveActionUp = (index: number) => {
    // Prevent moving the UserInput step or moving steps above it
    if (index <= 1) return;

    const newWorkflow = [...steps];
    const temp = newWorkflow[index];
    newWorkflow[index] = newWorkflow[index - 1];
    newWorkflow[index - 1] = temp;
    setSteps(newWorkflow);

    if (selectedStepIndex === index) {
      setSelectedStepIndex(index - 1);
    } else if (selectedStepIndex === index - 1) {
      setSelectedStepIndex(index);
    }
  };

  const moveActionDown = (index: number) => {
    if (index === steps.length - 1) return;

    const newWorkflow = [...steps];
    const temp = newWorkflow[index];
    newWorkflow[index] = newWorkflow[index + 1];
    newWorkflow[index + 1] = temp;
    setSteps(newWorkflow);

    if (selectedStepIndex === index) {
      setSelectedStepIndex(index + 1);
    } else if (selectedStepIndex === index + 1) {
      setSelectedStepIndex(index);
    }
  };

  const updateStep = (updatedStep: TypedWorkflowStep) => {
    if (selectedStepIndex === null) return;

    const newWorkflow = [...steps];
    newWorkflow[selectedStepIndex] = updatedStep;
    setSteps(newWorkflow);
  };

  const addVariable = () => {
    if (selectedStepIndex !== 0 || steps[0].type !== "UserInput") return;

    const p = steps[0].payload as UserInputStepPayload;
    const newVariable: WorkflowVariable = {
      name: "",
      type: "string",
      default_value: "",
      options: [],
    };
    p.variables.push(newVariable);

    const updatedStep: TypedWorkflowStep = {
      type: "UserInput",
      payload: p,
    };
    updateStep(updatedStep);
  };

  const updateVariable = (
    variableIndex: number,
    updatedVariable: WorkflowVariable,
  ) => {
    if (selectedStepIndex !== 0 || steps[0].type !== "UserInput") return;

    const p = steps[0].payload as UserInputStepPayload;
    p.variables[variableIndex] = updatedVariable;

    const updatedStep: TypedWorkflowStep = {
      type: "UserInput",
      payload: p,
    };
    updateStep(updatedStep);
  };

  // Remove a variable from the UserInput step
  const removeVariable = (variableIndex: number) => {
    if (selectedStepIndex !== 0 || steps[0].type !== "UserInput") return;

    const p = steps[0].payload as UserInputStepPayload;
    p.variables.splice(variableIndex, 1);

    const updatedStep: TypedWorkflowStep = {
      type: "UserInput",
      payload: p,
    };

    updateStep(updatedStep);
  };

  // Save the workflow
  const saveWorkflow = async () => {
    const wf = {
      variables: [] as WorkflowVariable[],
      steps: [] as TypedWorkflowStep[],
      name: workflowName,
      description: workflowDescription,
    } as Workflow;
    steps.forEach((s) => {
      switch (s.type) {
        case "UserInput":
          wf.variables = (s.payload as UserInputStepPayload).variables;
          break;
        default:
          wf.steps.push(s);
          break;
      }
    });
    if (id) {
      await updateWorkflow(id, wf);
    } else {
      await createWorkflow(wf);
    }
    onOpenChange(false);
    await onSave();
  };

  // Get the selected action
  const selectedStep =
    selectedStepIndex !== null ? steps[selectedStepIndex] : null;

  // Handle workflow name input blur
  const handleNameBlur = () => {
    setIsEditingName(false);
    // If name is empty, reset to default
    if (!workflowName.trim()) {
      setWorkflowName("New Workflow");
    }
  };

  const clear = () => {
    setWorkflowName("New Workflow");
    setWorkflowDescription("Enter workflow description...");
    setIsEditingDescription(false);
    setSteps([]);
    setSelectedStepIndex(null);
    setStepContexts([]);
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v) {
          clear();
        }
        onOpenChange(v);
      }}
    >
      <DialogContent className="max-w-5xl h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle hidden={true}></DialogTitle>
          <div className="flex items-center gap-2">
            {isEditingName ? (
              <Input
                value={workflowName}
                onChange={(e) => setWorkflowName(e.target.value)}
                onBlur={handleNameBlur}
                autoFocus
                className="text-xl font-semibold h-9 focus-visible:ring-offset-0"
                placeholder="Enter workflow name"
              />
            ) : (
              <div
                className="flex items-center gap-2 cursor-pointer group"
                onClick={() => setIsEditingName(true)}
              >
                <h2 className="text-xl font-semibold">{workflowName}</h2>
                <Pencil className="h-4 w-4" />
                <div className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
                  Click to edit
                </div>
              </div>
            )}
          </div>
          {isEditingDescription ? (
            <Input
              value={workflowDescription}
              onChange={(e) => setWorkflowDescription(e.target.value)}
              onBlur={() => {
                setIsEditingDescription(false);
                if (!workflowDescription.trim()) {
                  setWorkflowDescription("Enter workflow description...");
                }
              }}
              autoFocus
              className="text-sm h-9 focus-visible:ring-offset-0"
              placeholder="Enter workflow description"
            />
          ) : (
            <div
              className="flex items-center gap-2 cursor-pointer group text-sm text-muted-foreground"
              onClick={() => setIsEditingDescription(true)}
            >
              <span>{workflowDescription}</span>
              <Pencil className="h-3 w-3" />
              <div className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
                Click to edit description
              </div>
            </div>
          )}
        </DialogHeader>

        <div className="flex flex-1 gap-4 overflow-hidden">
          {/* Left Column - Workflow Steps */}
          <div className="w-2/5 border rounded-md flex flex-col">
            <div className="p-4 border-b bg-muted/40 flex justify-between items-center">
              <h3 className="font-medium">Workflow Steps</h3>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button size="sm" variant="outline">
                    <Plus className="h-4 w-4 mr-1" /> Add Step
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => addStep("CreateTable")}>
                    Create Table
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => addStep("DeleteTable")}>
                    Delete Table
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => addStep("CreateColumn")}>
                    Create Column
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => addStep("DeleteColumn")}>
                    Delete Column
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => addStep("Import")}>
                    Import Data
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => addStep("ExportTable")}>
                    Export Data
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => addStep("Generate")}>
                    Generate
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => addStep("Autofill")}>
                    Autofill
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            <ScrollArea className="flex-1">
              <div className="p-2">
                {steps.length === 0 ? (
                  <div className="text-center p-8 text-muted-foreground">
                    Loading workflow...
                  </div>
                ) : (
                  <div className="space-y-2">
                    {steps.map((step, index) => (
                      <div
                        key={index}
                        className={`p-3 border rounded-md flex items-center justify-between cursor-pointer hover:bg-muted/50 ${
                          selectedStepIndex === index
                            ? "border-primary bg-primary/5"
                            : ""
                        } ${step.type === "UserInput" ? "bg-muted/20" : ""}`}
                        onClick={() => setSelectedStepIndex(index)}
                      >
                        <div className="flex-1 flex items-center h-[40px]">
                          <div>
                            <div className="font-medium">{step.type}</div>
                          </div>
                        </div>
                        <div className="flex gap-1">
                          {/* Only show move up/down buttons for non-UserInput steps */}
                          {step.type !== "UserInput" && (
                            <>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  moveActionUp(index);
                                }}
                                disabled={index <= 1} // Disable if it's the first non-UserInput step
                              >
                                <ChevronUp className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  moveActionDown(index);
                                }}
                                disabled={index === steps.length - 1}
                              >
                                <ChevronDown className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  removeAction(index);
                                }}
                              >
                                <Trash2 className="h-4 w-4 text-destructive" />
                              </Button>
                            </>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </ScrollArea>
          </div>

          {/* Right Column - Action Properties */}
          <div
            className="w-3/5 border rounded-md flex flex-col"
            key={selectedStepIndex}
          >
            <ScrollArea className="flex-1">
              <div className="p-4">
                {selectedStep ? (
                  <div className="space-y-4">
                    {/* Common properties for all actions */}
                    <div className="space-y-2">
                      <Label>Step Type</Label>
                      <div className="flex items-center gap-2 p-2 bg-muted/30 border rounded-md">
                        <span className="font-medium">{selectedStep.type}</span>
                      </div>
                    </div>

                    {/* UserInput properties */}
                    {selectedStep.type === "UserInput" &&
                      selectedStepIndex !== null && (
                        <UserInputStep
                          step={selectedStep.payload as UserInputStepPayload}
                          context={stepContexts[selectedStepIndex]}
                          onAddVariable={addVariable}
                          onUpdateVariable={updateVariable}
                          onRemoveVariable={removeVariable}
                        />
                      )}

                    {/* CreateTable properties */}
                    {selectedStep.type === "CreateTable" &&
                      selectedStepIndex !== null && (
                        <CreateTableStep
                          step={selectedStep.payload as CreateTableStepPayload}
                          context={stepContexts[selectedStepIndex]}
                          onUpdateStep={(payload) =>
                            updateStep({ type: "CreateTable", payload })
                          }
                        />
                      )}

                    {/* DeleteTable properties */}
                    {selectedStep.type === "DeleteTable" &&
                      selectedStepIndex !== null && (
                        <DeleteTableStep
                          step={selectedStep.payload}
                          context={stepContexts[selectedStepIndex]}
                          onUpdateStep={(payload) =>
                            updateStep({ type: "DeleteTable", payload })
                          }
                        />
                      )}

                    {/* CreateColumn properties */}
                    {selectedStep.type === "CreateColumn" &&
                      selectedStepIndex !== null && (
                        <CreateColumnStep
                          step={selectedStep.payload as CreateColumnStepPayload}
                          context={stepContexts[selectedStepIndex]}
                          onUpdateStep={(payload) =>
                            updateStep({ type: "CreateColumn", payload })
                          }
                        />
                      )}

                    {/* DeleteColumn properties */}
                    {selectedStep.type === "DeleteColumn" &&
                      selectedStepIndex !== null && (
                        <DeleteColumnStep
                          step={selectedStep.payload}
                          context={stepContexts[selectedStepIndex]}
                          onUpdateStep={(payload) =>
                            updateStep({ type: "DeleteColumn", payload })
                          }
                        />
                      )}

                    {/* Generate properties */}
                    {selectedStep.type === "Generate" &&
                      selectedStepIndex !== null && (
                        <GenerateStep
                          step={selectedStep.payload}
                          context={stepContexts[selectedStepIndex]}
                          onUpdateStep={(payload) =>
                            updateStep({ type: "Generate", payload })
                          }
                        />
                      )}

                    {/* Autofill properties */}
                    {selectedStep.type === "Autofill" &&
                      selectedStepIndex !== null && (
                        <AutofillStep
                          step={selectedStep.payload}
                          context={stepContexts[selectedStepIndex]}
                          onUpdateStep={(payload) =>
                            updateStep({ type: "Autofill", payload })
                          }
                        />
                      )}

                    {/* ImportData properties */}
                    {selectedStep.type === "Import" &&
                      selectedStepIndex !== null &&
                      stepContexts[0] && (
                        <ImportStep
                          step={selectedStep.payload as ImportDataStepPayload}
                          context={stepContexts[selectedStepIndex]}
                          allUserInputVariables={(
                            steps[0].payload as UserInputStepPayload
                          ).variables.map((v) => ({
                            display: v.name,
                            path: v.name,
                            type: v.type,
                          }))}
                          onUpdateStep={(payload) =>
                            updateStep({ type: "Import", payload })
                          }
                        />
                      )}

                    {/* Export properties */}
                    {selectedStep.type === "ExportTable" &&
                      selectedStepIndex !== null && (
                        <ExportTableStep
                          step={selectedStep.payload}
                          context={stepContexts[selectedStepIndex]}
                          onUpdateStep={(payload) =>
                            updateStep({ type: "ExportTable", payload })
                          }
                        />
                      )}
                  </div>
                ) : (
                  <div className="text-center p-8 text-muted-foreground">
                    Select a step from the workflow to edit its properties.
                  </div>
                )}
              </div>
            </ScrollArea>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => {
              clear();
              onOpenChange(false);
            }}
          >
            Cancel
          </Button>
          <Button onClick={saveWorkflow}>Save Workflow</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
