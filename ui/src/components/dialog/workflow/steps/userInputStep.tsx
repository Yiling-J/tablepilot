import { PlusCircle, X } from "lucide-react";

import {
    UserInputStepPayload,
    WorkflowVariable,
    WorkflowVariableType,
} from "@/actions";
import { StepContext } from "../builder";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { MentionInput } from "@/components/ui/var-input";

interface UserInputStepProps {
  step: UserInputStepPayload;
  context: StepContext;
  onAddVariable: () => void;
  onUpdateVariable: (
    variableIndex: number,
    updatedVariable: WorkflowVariable,
  ) => void;
  onRemoveVariable: (variableIndex: number) => void;
}

export function UserInputStep({
  step,
  context,
  onAddVariable,
  onUpdateVariable,
  onRemoveVariable,
}: UserInputStepProps) {
  // Helper to update a specific variable
  const handleUpdateVariable = (
    idx: number,
    updatedVariable: WorkflowVariable,
  ) => {
    onUpdateVariable(idx, updatedVariable);
  };

  return (
    <div className="space-y-4">
      <p className="text-xs">
        This step lets you add variables used in the workflow. When creating
        other steps, you can type '@' to reference any variables defined here.
        When the workflow starts, you’ll be prompted to input or select values
        for all variables first.
      </p>
      <div className="flex justify-between items-center">
        <Label>Variables</Label>
        <Button
          size="sm"
          variant="outline"
          onClick={onAddVariable}
          className="flex items-center gap-1"
        >
          <PlusCircle className="h-3.5 w-3.5" /> Add Variable
        </Button>
      </div>

      {step.variables.length === 0 ? (
        <div className="text-center p-4 border rounded-md text-muted-foreground">
          No variables defined. Click "Add Variable" to create one.
        </div>
      ) : (
        <div className="space-y-4">
          {step.variables.map((variable, idx) => (
            <div key={idx} className="p-3 border rounded-md space-y-3">
              <div className="flex justify-between items-center">
                <h4 className="font-medium">Variable {idx + 1}</h4>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => onRemoveVariable(idx)}
                >
                  <X className="h-4 w-4 text-muted-foreground" />
                </Button>
              </div>

              <div className="space-y-2">
                <Label htmlFor={`var-name-${idx}`}>Name</Label>
                <Input
                  id={`var-name-${idx}`}
                  value={variable.name}
                  onChange={(e) => {
                    const newName = e.target.value;
                    const isValid = /^[a-zA-Z0-9_]*$/.test(newName);
                    if (isValid) {
                      handleUpdateVariable(idx, {
                        ...variable,
                        name: newName,
                      });
                    }
                    // If not isValid and not empty, the name is not updated.
                  }}
                  placeholder="e.g. tableName"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor={`var-type-${idx}`}>Type</Label>
                <Select
                  value={variable.type}
                  onValueChange={(value: WorkflowVariableType) =>
                    handleUpdateVariable(idx, {
                      ...variable,
                      type: value,
                    })
                  }
                >
                  <SelectTrigger id={`var-type-${idx}`}>
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="string">String</SelectItem>
                    <SelectItem value="integer">Integer</SelectItem>
                    <SelectItem value="number">Number</SelectItem>
                    <SelectItem value="file">File</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {variable.type !== "file" && (
                <div className="space-y-2">
                  <Label htmlFor={`var-default-${idx}`}>Default Value</Label>
                  <MentionInput
                    id={`var-default-${idx}`}
                    variables={context.variables}
                    value={variable.default_value}
                    onChange={(v) =>
                      handleUpdateVariable(idx, {
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
              )}

              {variable.type !== "file" && (
                <div className="space-y-2">
                  <Label htmlFor={`var-options-${idx}`}>
                    Options (optional, one per line)
                  </Label>
                  <Textarea
                    id={`var-options-${idx}`}
                    defaultValue={variable.options.join("\n")}
                    onChange={(e) => {
                      handleUpdateVariable(idx, {
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
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
