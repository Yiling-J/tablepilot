import {
    ChevronDown,
    ChevronUp,
    Pencil,
    Plus,
    PlusCircle,
    Trash2,
    X,
} from "lucide-react";

import { useEffect, useState } from "react";

import {
    CreateColumnStepPayload,
    CreateTableStepPayload,
    ImportDataStepPayload,
    TypedWorkflowStep,
    UserInputStepPayload,
    Workflow,
    WorkflowStepType,
    WorkflowVariable,
    WorkflowVariableType,
    createWorkflow,
    updateWorkflow,
} from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
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
import { NumberInput } from "@/components/ui/number-input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { ContextVariable, MentionInput } from "@/components/ui/var-input";
import { CreateTableDialog } from "../create-table";

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
  const [steps, setSteps] = useState<TypedWorkflowStep[]>([]);
  const [selectedStepIndex, setSelectedStepIndex] = useState<number | null>(
    null,
  );
  const [stepVariables, setStepVariables] = useState<ContextVariable[][]>([]);
  const [createTableDialogOpen, setCreateTableDialogOpen] = useState(false);

  useEffect(() => {
    if (workflow) {
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
    }
  }, [workflow]);

  // Initialize workflow with UserInput step
  useEffect(() => {
    if (steps.length === 0) {
      setSteps([
        {
          type: "UserInput",
          payload: { variables: [] } as UserInputStepPayload,
        },
      ]);
      setSelectedStepIndex(0);
    }
  }, [steps.length]);

  // workflow validation and variables reset
  useEffect(() => {
    const vars: ContextVariable[][] = [
      [
        { path: "date", display: "date", type: "string" },
        { path: "time", display: "time", type: "string" },
        { path: "datetime", display: "datetime", type: "string" },
      ],
    ];
    steps.forEach((step, index) => {
      const nv: ContextVariable[] = [...vars[index]];
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
          nv.push({
            display: `CreateTable[${(step.payload as CreateTableStepPayload).request?.name ?? ""}].table`,
            path: `step${index}.table`,
            type: "string",
          });
          break;
        case "CreateColumn":
          nv.push({
            display: `CreateColumn[${(step.payload as CreateColumnStepPayload).name}].column`,
            path: `step${index}.column`,
            type: "string",
          });
          break;
        case "Import":
          nv.push({
            display: `Import[${(step.payload as ImportDataStepPayload).file}].table`,
            path: `step${index}.table`,
            type: "string",
          });
          break;
      }
      vars.push(nv);
    });
    setStepVariables(vars);
  }, [steps]);

  // Helper to create a new action
  const createNewStep = (type: WorkflowStepType): TypedWorkflowStep => {
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
        return { type, payload: { file: "", prompt: "" } };
      case "Generate":
        return { type, payload: { count: 20, batch: 5, table: "" } };
      case "Autofill":
        return {
          type,
          payload: { count: 20, batch: 5, table: "", columns: [] },
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
    } as Workflow;
    steps.forEach((s) => {
      switch (s.type) {
        case "UserInput":
          wf.variables = (s.payload as UserInputStepPayload).variables;
          break;
        default:
          wf.steps.push(s);
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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl h-[80vh] flex flex-col">
        <DialogHeader>
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
          <DialogDescription>
            Create a workflow by adding steps and configuring their properties.
          </DialogDescription>
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
          <div className="w-3/5 border rounded-md flex flex-col">
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
                    {selectedStep.type === "UserInput" && (
                      <div className="space-y-4">
                        <div className="flex justify-between items-center">
                          <Label>Variables</Label>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={addVariable}
                            className="flex items-center gap-1"
                          >
                            <PlusCircle className="h-3.5 w-3.5" /> Add Variable
                          </Button>
                        </div>

                        {selectedStep.payload.variables.length === 0 ? (
                          <div className="text-center p-4 border rounded-md text-muted-foreground">
                            No variables defined. Click "Add Variable" to create
                            one.
                          </div>
                        ) : (
                          <div className="space-y-4">
                            {selectedStep.payload.variables.map(
                              (variable, idx) => (
                                <div
                                  key={idx}
                                  className="p-3 border rounded-md space-y-3"
                                >
                                  <div className="flex justify-between items-center">
                                    <h4 className="font-medium">
                                      Variable {idx + 1}
                                    </h4>
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => removeVariable(idx)}
                                    >
                                      <X className="h-4 w-4 text-muted-foreground" />
                                    </Button>
                                  </div>

                                  <div className="space-y-2">
                                    <Label htmlFor={`var-name-${idx}`}>
                                      Name
                                    </Label>
                                    <Input
                                      id={`var-name-${idx}`}
                                      value={variable.name}
                                      onChange={(e) =>
                                        updateVariable(idx, {
                                          ...variable,
                                          name: e.target.value,
                                        })
                                      }
                                      placeholder="e.g. tableName"
                                    />
                                  </div>

                                  <div className="space-y-2">
                                    <Label htmlFor={`var-type-${idx}`}>
                                      Type
                                    </Label>
                                    <Select
                                      value={variable.type}
                                      onValueChange={(
                                        value: WorkflowVariableType,
                                      ) =>
                                        updateVariable(idx, {
                                          ...variable,
                                          type: value,
                                        })
                                      }
                                    >
                                      <SelectTrigger id={`var-type-${idx}`}>
                                        <SelectValue placeholder="Select type" />
                                      </SelectTrigger>
                                      <SelectContent>
                                        <SelectItem value="string">
                                          String
                                        </SelectItem>
                                        <SelectItem value="integer">
                                          Integer
                                        </SelectItem>
                                        <SelectItem value="number">
                                          Number
                                        </SelectItem>
                                        <SelectItem value="file">
                                          File
                                        </SelectItem>
                                      </SelectContent>
                                    </Select>
                                  </div>

                                  <div className="space-y-2">
                                    <Label htmlFor={`var-default-${idx}`}>
                                      Default Value
                                    </Label>
                                    <MentionInput
                                      id={`var-default-${idx}`}
                                      value={variable.default_value}
                                      onChange={(v) =>
                                        updateVariable(idx, {
                                          ...variable,
                                          default_value: v.target.value,
                                        })
                                      }
                                      placeholder={
                                        variable.type === "string"
                                          ? 'e.g. "users"'
                                          : variable.type === "number"
                                            ? "e.g. 3.14"
                                            : variable.type === "integer"
                                              ? "e.g. 20"
                                              : 'e.g. {"key": "value"}'
                                      }
                                    />
                                  </div>

                                  <div className="space-y-2">
                                    <Label htmlFor={`var-options-${idx}`}>
                                      Options (optional, one per line)
                                    </Label>
                                    <Textarea
                                      id={`var-options-${idx}`}
                                      defaultValue={variable.options.join("\n")}
                                      onChange={(e) => {
                                        updateVariable(idx, {
                                          ...variable,
                                          options: e.target.value
                                            .split("\n")
                                            .filter((opt) => opt.trim() !== ""),
                                        });
                                      }}
                                      className="w-full min-h-[100px] p-2 border rounded-md"
                                      placeholder="Enter options, one per line, will select one option when workflow start instead input value manually."
                                    />
                                  </div>
                                </div>
                              ),
                            )}
                          </div>
                        )}
                      </div>
                    )}

                    {/* CreateTable properties */}
                    {selectedStep.type === "CreateTable" && (
                      <div>
                        <div className="flex flex-row mb-5 items-center">
                          <div className="w-24">
                            <Label htmlFor="onExists">On Exists</Label>
                          </div>
                          <Select
                            value={selectedStep.payload.on_exists}
                            onValueChange={(value) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  on_exists: value,
                                },
                              })
                            }
                          >
                            <SelectTrigger id="onExists">
                              <SelectValue placeholder="What to do if table exists" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="Stop">
                                Stop workflow
                              </SelectItem>
                              <SelectItem value="Recreate">
                                Recreate the table
                              </SelectItem>
                              <SelectItem value="Skip">
                                Do nothing and continue
                              </SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                        <CreateTableDialog
                          variables={stepVariables[selectedStepIndex!]}
                          isOpen={createTableDialogOpen}
                          setIsOpen={setCreateTableDialogOpen}
                          close={() => {}}
                          form={selectedStep.payload.request}
                          onSave={(v) => {
                            updateStep({
                              type: selectedStep.type,
                              payload: {
                                ...selectedStep.payload,
                                request: v,
                              },
                            });
                            setCreateTableDialogOpen(false);
                          }}
                        />
                        <div
                          className="border rounded-lg p-4 cursor-pointer hover:bg-muted/50 transition-colors"
                          onClick={() => setCreateTableDialogOpen(true)}
                        >
                          <div className="flex items-center justify-between mb-2">
                            <h3 className="font-medium">Table Configuration</h3>
                            <Pencil className="h-4 w-4 text-muted-foreground" />
                          </div>
                          <div className="space-y-2 text-sm">
                            <div className="flex items-center gap-2">
                              <span className="text-muted-foreground">
                                Name:
                              </span>
                              <span>
                                {selectedStep.payload.request.name ||
                                  "Unnamed Table"}
                              </span>
                            </div>
                            <div className="flex items-center gap-2">
                              <span className="text-muted-foreground">
                                Description:
                              </span>
                              <span>
                                {selectedStep.payload.request.description ||
                                  "No description"}
                              </span>
                            </div>
                            <div className="flex items-center gap-2">
                              <span className="text-muted-foreground">
                                Columns:
                              </span>
                              <span>
                                {selectedStep.payload.request.columns.length}{" "}
                                columns
                              </span>
                            </div>
                            <div className="flex items-center gap-2">
                              <span className="text-muted-foreground">
                                Sources:
                              </span>
                              <span>
                                {selectedStep.payload.request.sources.length}{" "}
                                sources
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>
                    )}

                    {/* DeleteTable properties */}
                    {selectedStep.type === "DeleteTable" && (
                      <div className="space-y-2">
                        <Label htmlFor="tableName">Table Name</Label>
                        <MentionInput
                          variables={stepVariables[selectedStepIndex!]}
                          id="tableName"
                          value={selectedStep.payload.table}
                          onChange={(e) =>
                            updateStep({
                              type: selectedStep.type,
                              payload: {
                                ...selectedStep.payload,
                                table: e.target.value,
                              },
                            })
                          }
                        />
                      </div>
                    )}

                    {/* CreateColumn properties */}
                    {selectedStep.type === "CreateColumn" && (
                      <div key={selectedStepIndex}>
                        <div className="space-y-2">
                          <Label htmlFor="tableName">Table Name</Label>
                          <MentionInput
                            variables={stepVariables[selectedStepIndex!]}
                            id="tableName"
                            value={selectedStep.payload.table}
                            onChange={(e) => {
                              console.log(e.target.value);
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  table: e.target.value,
                                },
                              });
                            }}
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="columnName">Column Name</Label>
                          <MentionInput
                            id="columnName"
                            variables={stepVariables[selectedStepIndex!]}
                            value={selectedStep.payload.name}
                            onChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  name: e.target.value,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="columnDescription">
                            Column Description
                          </Label>
                          <MentionInput
                            id="columnDescription"
                            variables={stepVariables[selectedStepIndex!]}
                            value={selectedStep.payload.description}
                            onChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  description: e.target.value,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="dataType">Data Type</Label>
                          <Select
                            value={selectedStep.payload.type}
                            onValueChange={(value) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  type: value,
                                },
                              })
                            }
                          >
                            <SelectTrigger id="dataType">
                              <SelectValue placeholder="Select data type" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="string">String</SelectItem>
                              <SelectItem value="integer">Integer</SelectItem>
                              <SelectItem value="number">Number</SelectItem>
                              <SelectItem value="image">Image</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>
                    )}

                    {/* DeleteColumn properties */}
                    {selectedStep.type === "DeleteColumn" && (
                      <>
                        <div className="space-y-2">
                          <Label htmlFor="tableName">Table Name</Label>
                          <MentionInput
                            id="tableName"
                            value={selectedStep.payload.table}
                            variables={stepVariables[selectedStepIndex!]}
                            onChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  table: e.target.value,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="columnName">Column Name</Label>
                          <MentionInput
                            id="columnName"
                            variables={stepVariables[selectedStepIndex!]}
                            value={selectedStep.payload.column}
                            onChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  column: e.target.value,
                                },
                              })
                            }
                          />
                        </div>
                      </>
                    )}

                    {/* Generate properties */}
                    {selectedStep.type === "Generate" && (
                      <>
                        <div className="space-y-2">
                          <Label htmlFor="tableName">Table Name</Label>
                          <MentionInput
                            id="tableName"
                            value={selectedStep.payload.table}
                            variables={stepVariables[selectedStepIndex!]}
                            onChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  table: e.target.value,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="generateCount">Count</Label>
                          <NumberInput
                            id="GenerateCount"
                            value={selectedStep.payload.count}
                            onValueChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  count: e ?? 10,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="generateBatch">Batch</Label>
                          <NumberInput
                            id="GenerateBatch"
                            value={selectedStep.payload.batch}
                            onValueChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  batch: e ?? 2,
                                },
                              })
                            }
                          />
                        </div>
                      </>
                    )}

                    {/* Autofill properties */}
                    {selectedStep.type === "Autofill" && (
                      <>
                        <div className="space-y-2">
                          <Label htmlFor="tableName">Table Name</Label>
                          <MentionInput
                            id="tableName"
                            value={selectedStep.payload.table}
                            variables={stepVariables[selectedStepIndex!]}
                            onChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  table: e.target.value,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="generateCount">Count</Label>
                          <NumberInput
                            id="GenerateCount"
                            value={selectedStep.payload.count}
                            onValueChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  count: e ?? 10,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="generateBatch">Batch</Label>
                          <NumberInput
                            id="GenerateBatch"
                            value={selectedStep.payload.count}
                            onValueChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  batch: e ?? 2,
                                },
                              })
                            }
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="generateBatch">Columns</Label>
                          <MentionInput
                            id="AutofillColumns"
                            placeholder="Columns need to be autofilled, one per line. Use @ to reference columns from previous steps. Columns will be autofilled automatically."
                            value={selectedStep.payload.columns.join("\n")}
                            variables={stepVariables[selectedStepIndex!]}
                            rows={3}
                            textarea={true}
                            onChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  columns: e.target.value
                                    .split("\n")
                                    .map((v) => v.trim()),
                                },
                              })
                            }
                          />
                        </div>
                      </>
                    )}

                    {/* ImportData properties */}
                    {selectedStep.type === "Import" && (
                      <div className="space-y-2">
                        <Label htmlFor="importFile">File Variable</Label>
                        <Select
                          value={selectedStep.payload.file}
                          onValueChange={(value) =>
                            updateStep({
                              type: selectedStep.type,
                              payload: {
                                ...selectedStep.payload,
                                file: value,
                              },
                            })
                          }
                        >
                          <SelectTrigger id="importFile">
                            <SelectValue placeholder="Select file variable" />
                          </SelectTrigger>
                          <SelectContent>
                            {stepVariables[1]
                              .filter((v) => v.type === "file")
                              .map((v, idx) => (
                                <SelectItem key={idx} value={v.path}>
                                  {v.display}
                                </SelectItem>
                              ))}
                          </SelectContent>
                        </Select>
                        <div>
                          <Label htmlFor="importPrompt">Prompt</Label>
                          <MentionInput
                            id="importPrompt"
                            className="mt-2"
                            textarea={true}
                            rows={3}
                            value={selectedStep.payload.prompt}
                            variables={stepVariables[selectedStepIndex!]}
                            onChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  prompt: e.target.value,
                                },
                              })
                            }
                          />
                        </div>
                      </div>
                    )}

                    {/* Export properties */}
                    {selectedStep.type === "ExportTable" && (
                      <div>
                        <div className="space-y-2">
                          <Label htmlFor="tableName">Table Name</Label>
                          <MentionInput
                            id="tableName"
                            value={selectedStep.payload.table}
                            variables={stepVariables[selectedStepIndex!]}
                            onChange={(e) =>
                              updateStep({
                                type: selectedStep.type,
                                payload: {
                                  ...selectedStep.payload,
                                  table: e.target.value,
                                },
                              })
                            }
                          />
                        </div>
                      </div>
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
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={saveWorkflow}>Save Workflow</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
