import { Pencil } from "lucide-react";
import { useState } from "react";

import { CreateTableStepPayload } from "@/actions";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { CreateTableDialog } from "../../create-table";
import { StepContext } from "../builder";

interface CreateTableStepProps {
  step: CreateTableStepPayload;
  context: StepContext;
  onUpdateStep: (payload: CreateTableStepPayload) => void;
}

export function CreateTableStep({
  step,
  context,
  onUpdateStep,
}: CreateTableStepProps) {
  const [createTableDialogOpen, setCreateTableDialogOpen] = useState(false);

  return (
    <div>
      <div className="flex flex-row mb-5 items-center">
        <div className="w-24">
          <Label htmlFor="onExists">On Exists</Label>
        </div>
        <Select
          value={step.on_exists}
          onValueChange={(value) =>
            onUpdateStep({
              ...step,
              on_exists: value,
            })
          }
        >
          <SelectTrigger id="onExists">
            <SelectValue placeholder="What to do if table exists" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="Stop">Stop workflow</SelectItem>
            <SelectItem value="Recreate">Recreate the table</SelectItem>
            <SelectItem value="Skip">Do nothing and continue</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <CreateTableDialog
        variables={context.variables}
        isOpen={createTableDialogOpen}
        setIsOpen={setCreateTableDialogOpen}
        close={() => {}} // Assuming close can be a no-op if handled by setIsOpen
        form={step.request}
        onSave={(v) => {
          onUpdateStep({
            ...step,
            request: v,
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
            <span className="text-muted-foreground">Name:</span>
            <span>{step.request.name || "Unnamed Table"}</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-muted-foreground">Description:</span>
            <span>{step.request.description || "No description"}</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-muted-foreground">Columns:</span>
            <span>{step.request.columns.length} columns</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-muted-foreground">Sources:</span>
            <span>{step.request.sources.length} sources</span>
          </div>
        </div>
      </div>
    </div>
  );
}
