import {
  ExportStepPayload, // Assuming this type exists, if not, it might be ExportTableStepPayload
} from "@/actions";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { StepContext } from "../builder";

interface ExportTableStepProps {
  step: ExportStepPayload; // Or ExportTableStepPayload
  context: StepContext;
  onUpdateStep: (payload: ExportStepPayload) => void; // Or ExportTableStepPayload
}

export function ExportTableStep({
  step,
  context,
  onUpdateStep,
}: ExportTableStepProps) {
  return (
    <div>
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
            <SelectValue placeholder="Select a table to export" />
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
    </div>
  );
}
