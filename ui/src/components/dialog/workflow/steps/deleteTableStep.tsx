import { DeleteTableStepPayload } from "@/actions";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { StepContext } from "../builder";

interface DeleteTableStepProps {
  step: DeleteTableStepPayload;
  context: StepContext;
  onUpdateStep: (payload: DeleteTableStepPayload) => void;
}

export function DeleteTableStep({
  step,
  context,
  onUpdateStep,
}: DeleteTableStepProps) {
  return (
    <div className="space-y-2">
      <Label htmlFor="tableName">Table Name</Label>
      <Select
        value={step.table}
        onValueChange={(value) =>
          onUpdateStep({
            ...step,
            table: value,
          })
        }
      >
        <SelectTrigger id="tables">
          <SelectValue placeholder="Select a table" />
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
  );
}
