import {
  DeleteColumnStepPayload,
  TableInfo, // Assuming TableInfo is needed for context.tables
} from "@/actions";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { StepContext } from "../builder"; // Import StepContext from builder

interface DeleteColumnStepProps {
  step: DeleteColumnStepPayload;
  context: StepContext;
  onUpdateStep: (payload: DeleteColumnStepPayload) => void;
}

export function DeleteColumnStep({
  step,
  context,
  onUpdateStep,
}: DeleteColumnStepProps) {
  const selectedTableInfo = context.tables.find((t) => t.id === step.table);

  const handleTableChange = (tableId: string) => {
    const newTableInfo = context.tables.find((t) => t.id === tableId);
    let column = step.column;
    // Reset column if it's not in the new table
    if (newTableInfo && !newTableInfo.columns.find((c) => c.id === column)) {
      column = ""; // Or set to a default/placeholder if applicable
    }
    onUpdateStep({
      table: tableId,
      column: column,
    });
  };

  const handleColumnChange = (columnId: string) => {
    onUpdateStep({
      ...step,
      column: columnId,
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
      <div className="space-y-2">
        <Label htmlFor="columnName">Column</Label>
        <Select
          value={step.column}
          onValueChange={handleColumnChange}
          disabled={!step.table} // Disable if no table is selected
        >
          <SelectTrigger id="column">
            <SelectValue placeholder="Select a column" />
          </SelectTrigger>
          <SelectContent>
            {(selectedTableInfo?.columns ?? []).map((c) => (
              <SelectItem key={c.id} value={c.id}>
                {c.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </>
  );
}
