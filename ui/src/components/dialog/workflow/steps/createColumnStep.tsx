import { CreateColumnStepPayload } from "@/actions";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { MentionInput } from "@/components/ui/var-input";
import { StepContext } from "../builder";

interface CreateColumnStepProps {
  step: CreateColumnStepPayload;
  context: StepContext;
  onUpdateStep: (payload: CreateColumnStepPayload) => void;
}

export function CreateColumnStep({
  step,
  context,
  onUpdateStep,
}: CreateColumnStepProps) {
  return (
    <div className="space-y-2">
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
            {context.tables.map((t, i) => (
              <SelectItem value={t.id} key={i}>
                {t.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="columnName">Column Name</Label>
        <MentionInput
          id="columnName"
          variables={context.variables}
          value={step.name}
          onChange={(e) =>
            onUpdateStep({
              ...step,
              name: e.target.value,
            })
          }
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="columnDescription">Column Description</Label>
        <MentionInput
          id="columnDescription"
          variables={context.variables}
          value={step.description}
          onChange={(e) =>
            onUpdateStep({
              ...step,
              description: e.target.value,
            })
          }
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="dataType">Data Type</Label>
        <Select
          value={step.type}
          onValueChange={(value) =>
            onUpdateStep({
              ...step,
              type: value,
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
  );
}
