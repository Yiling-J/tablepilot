import { ImportDataStepPayload } from "@/actions";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { ContextVariable, MentionInput } from "@/components/ui/var-input";
import { StepContext } from "../builder";

interface ImportStepProps {
  step: ImportDataStepPayload;
  context: StepContext; // For current step's variables (prompt) and available tables
  allUserInputVariables: ContextVariable[]; // For selecting 'file' type variables from UserInput step
  onUpdateStep: (payload: ImportDataStepPayload) => void;
}

export function ImportStep({
  step,
  context,
  allUserInputVariables,
  onUpdateStep,
}: ImportStepProps) {
  const handleImportOptionChange = (value: string) => {
    const newPayload: ImportDataStepPayload = {
      ...step,
    };
    if (value === "Create new table") {
      newPayload.table = ""; // Reset existing table selection
      newPayload.truncate = false; // Not applicable for new table
      // newPayload.name = step.name; // Keep existing name or clear it: ""
    } else {
      // "Import into existing table"
      newPayload.name = ""; // Reset new table name
      // Set to first available table if one exists, otherwise empty.
      // User should ideally select one.
      newPayload.table = context.tables.length > 0 ? context.tables[0].id : "";
    }
    onUpdateStep(newPayload);
  };

  // Determine current import option based on payload state
  let currentImportOption = "Create new table"; // Default
  if (step.table && step.table !== "") {
    currentImportOption = "Import into existing table";
  } else if (step.name && step.name !== "") {
    currentImportOption = "Create new table";
  }

  return (
    <div className="space-y-4">
      {/* File Variable Select */}
      <div className="space-y-2">
        <Label htmlFor="importFile">File</Label>
        <Select
          value={step.file}
          onValueChange={(value) =>
            onUpdateStep({
              ...step,
              file: value,
            })
          }
        >
          <SelectTrigger id="importFile">
            <SelectValue placeholder="Select file variable" />
          </SelectTrigger>
          <SelectContent>
            {allUserInputVariables // Use dedicated prop for UserInput variables
              .filter((v) => v.type === "file")
              .map((v, idx) => (
                <SelectItem key={idx} value={`{{.${v.display}}}`}>
                  {v.display}
                </SelectItem>
              ))}
          </SelectContent>
        </Select>
        <p className="text-xs text-primary/80">
          You can select any file-type variable defined in the UserInput step
          here.
        </p>
      </div>

      {/* Import Option Select */}
      <div className="space-y-2">
        <Label htmlFor="importOption">Import Option</Label>
        <Select
          value={currentImportOption}
          onValueChange={handleImportOptionChange}
        >
          <SelectTrigger id="importOption">
            <SelectValue placeholder="Select import option" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="Create new table">Create new table</SelectItem>
            <SelectItem value="Import into existing table">
              Import into existing table
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Conditional Fields for "Create new table" */}
      {currentImportOption === "Create new table" && (
        <div className="space-y-2">
          <Label htmlFor="newTableName">New Table Name</Label>
          <MentionInput
            id="newTableName"
            variables={context.variables} // Current step context variables for @-mentions
            value={step.name ?? ""}
            onChange={(e) =>
              onUpdateStep({
                ...step,
                name: e.target.value,
                table: "", // Ensure table is empty for "create new"
              })
            }
            placeholder="Enter new table name or use @ for variables"
          />
        </div>
      )}

      {/* Conditional Fields for "Import into existing table" */}
      {currentImportOption === "Import into existing table" && (
        <>
          <div className="space-y-2">
            <Label htmlFor="existingTableName">Select Table</Label>
            <Select
              value={step.table ?? ""}
              onValueChange={(value) =>
                onUpdateStep({
                  ...step,
                  table: value,
                  name: "", // Ensure name is empty for "import existing"
                })
              }
            >
              <SelectTrigger id="existingTableName">
                <SelectValue placeholder="Select existing table" />
              </SelectTrigger>
              <SelectContent>
                {context.tables.map((t) => (
                  <SelectItem key={t.id} value={t.id}>
                    {t.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center space-x-2 mt-2">
            <Checkbox
              id="truncateTable"
              checked={step.truncate}
              onCheckedChange={(checked) =>
                onUpdateStep({
                  ...step,
                  truncate: !!checked,
                })
              }
            />
            <Label htmlFor="truncateTable" className="cursor-pointer">
              Truncate table before import
            </Label>
          </div>
        </>
      )}

      {/* Prompt */}
      <div className="space-y-2">
        <Label htmlFor="importPrompt">Prompt (Import Image)</Label>
        <MentionInput
          id="importPrompt"
          className="mt-2"
          textarea={true}
          rows={3}
          placeholder="Used only when importing images, as AI is required to extract data from them."
          value={step.prompt}
          variables={context.variables} // Current step context variables for @-mentions
          onChange={(e) =>
            onUpdateStep({
              ...step,
              prompt: e.target.value,
            })
          }
        />
      </div>
    </div>
  );
}
