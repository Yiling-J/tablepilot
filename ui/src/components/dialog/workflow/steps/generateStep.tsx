import { GenerateStepPayload } from "@/actions";
import { Label } from "@/components/ui/label";
import { NumberInput } from "@/components/ui/number-input";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { StepContext } from "../builder";

interface GenerateStepProps {
  step: GenerateStepPayload;
  context: StepContext;
  onUpdateStep: (payload: GenerateStepPayload) => void;
}

export function GenerateStep({
  step,
  context,
  onUpdateStep,
}: GenerateStepProps) {
  return (
    <>
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
      <div className="space-y-2">
        <Label htmlFor="GenerateCount">Count</Label>
        <NumberInput
          id="GenerateCount"
          value={step.count}
          onValueChange={(value) => {
            onUpdateStep({
              ...step,
              count: value ?? 0,
            });
          }}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="GenerateBatch">Batch</Label>
        <NumberInput
          id="GenerateBatch"
          value={step.batch}
          onValueChange={(value) =>
            onUpdateStep({
              ...step,
              batch: value ?? 0,
            })
          }
        />
      </div>
    </>
  );
}
