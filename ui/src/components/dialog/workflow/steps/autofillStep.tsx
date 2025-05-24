import {
  AutofillStepPayload,
  // ColumnInfo, // For AutofillInput allColumns - Removed as unused
} from "@/actions";
import { Label } from "@/components/ui/label";
import { NumberInput } from "@/components/ui/number-input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AutofillInput } from "../../autofill-input"; // Corrected path
import { StepContext } from "../builder";

interface AutofillStepProps {
  step: AutofillStepPayload;
  context: StepContext;
  onUpdateStep: (payload: AutofillStepPayload) => void;
}

export function AutofillStep({
  step,
  context,
  onUpdateStep,
}: AutofillStepProps) {
  const selectedTableInfo = context.tables.find((t) => t.id === step.table);

  const handleTableChange = (tableId: string) => {
    onUpdateStep({
      ...step,
      table: tableId,
      columns: [], // Reset columns
      context_columns: [], // Reset context_columns
    });
  };

  return (
    <>
      <div className="space-y-2">
        <Label htmlFor="tableName">Table Name</Label>
        <Select value={step.table} onValueChange={handleTableChange}>
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
      <AutofillInput
        allColumns={selectedTableInfo?.columns ?? []}
        variables={context.variables}
        columns={step.columns}
        contextColumns={step.context_columns}
        prompt={step.prompt}
        onColumnsChange={(v) => {
          onUpdateStep({
            ...step,
            columns: [...v],
          });
        }}
        onContextColumnsChange={(v) => {
          onUpdateStep({
            ...step,
            context_columns: [...v],
          });
        }}
        onPromptChange={(v) => {
          onUpdateStep({
            ...step,
            prompt: v,
          });
        }}
      />
      <div className="flex flex-row items-center space-x-4">
        <div className="flex flex-row items-center space-x-2">
          <Label htmlFor="AutofillCount">Count</Label>
          <NumberInput
            id="AutofillCount"
            value={step.count}
            onValueChange={(value) =>
              onUpdateStep({
                ...step,
                count: value ?? 0,
              })
            }
          />
        </div>
        <div className="flex flex-row items-center space-x-2">
          <Label htmlFor="AutofillBatch">Batch</Label>
          <NumberInput
            id="AutofillBatch"
            value={step.batch}
            onValueChange={(value) =>
              onUpdateStep({
                ...step,
                batch: value ?? 0,
              })
            }
          />
        </div>
      </div>
    </>
  );
}
