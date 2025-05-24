import { Column } from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
} from "@/components/ui/command";
import { Label } from "@/components/ui/label";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { Check, ChevronsUpDown } from "lucide-react";
import { useState } from "react";
import { ContextVariable, MentionInput } from "../ui/var-input";

interface AutofillInputProps {
  allColumns: Column[];
  columns: string[];
  contextColumns: string[];
  prompt: string;
  onColumnsChange: (columns: string[]) => void;
  onContextColumnsChange: (columns: string[]) => void;
  onPromptChange: (prompt: string) => void;
  variables?: ContextVariable[];
}

export function AutofillInput({
  allColumns,
  columns,
  contextColumns,
  prompt,
  onColumnsChange,
  onContextColumnsChange,
  onPromptChange,
  variables,
}: AutofillInputProps) {
  const [columnsOpen, setColumnsOpen] = useState(false);
  const [contextColumnsOpen, setContextColumnsOpen] = useState(false);

  const toggleColumn = (column: string) => {
    onColumnsChange(
      columns.includes(column)
        ? columns.filter((c) => c !== column)
        : [...columns, column],
    );
  };

  const toggleContextColumn = (column: string) => {
    onContextColumnsChange(
      contextColumns.includes(column)
        ? contextColumns.filter((c) => c !== column)
        : [...contextColumns, column],
    );
  };

  const addAllColumns = () => {
    onColumnsChange(allColumns.map((e) => e.id));
  };

  const addAllContextColumns = () => {
    onContextColumnsChange(allColumns.map((e) => e.id));
  };

  return (
    <div className="grid gap-6 py-4">
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label htmlFor="columns">Columns</Label>
          <Button
            variant="ghost"
            size="sm"
            onClick={addAllColumns}
            className="h-8 text-xs"
            type="button"
          >
            Add All
          </Button>
        </div>
        <Popover open={columnsOpen} onOpenChange={setColumnsOpen} modal={true}>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              role="combobox"
              aria-expanded={columnsOpen}
              className="w-full justify-between"
            >
              {columns.length > 0
                ? `${columns.length} selected`
                : "Select columns..."}
              <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-full p-0">
            <Command>
              <CommandInput placeholder="Search columns..." />
              <CommandList>
                <CommandEmpty>No column found.</CommandEmpty>
                <CommandGroup>
                  {allColumns.map((column) => (
                    <CommandItem
                      key={column.id}
                      value={column.id}
                      onSelect={() => toggleColumn(column.id)}
                    >
                      <Check
                        className={cn(
                          "mr-2 h-4 w-4",
                          columns.includes(column.id)
                            ? "opacity-100"
                            : "opacity-0",
                        )}
                      />
                      {column.name}
                    </CommandItem>
                  ))}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
        {columns.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {columns.map((column) => (
              <div
                key={column}
                className="flex items-center rounded-md bg-muted px-2 py-1 text-sm"
              >
                {allColumns.find((c) => c.id === column)?.name}
                <button
                  className="ml-1 rounded-full text-muted-foreground hover:text-foreground"
                  onClick={() => toggleColumn(column)}
                >
                  ×
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label htmlFor="contextColumns">Context Columns</Label>
          <Button
            variant="ghost"
            size="sm"
            onClick={addAllContextColumns}
            className="h-8 text-xs"
            type="button"
          >
            Add All
          </Button>
        </div>
        <Popover
          open={contextColumnsOpen}
          onOpenChange={setContextColumnsOpen}
          modal={true}
        >
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              role="combobox"
              aria-expanded={contextColumnsOpen}
              className="w-full justify-between"
            >
              {contextColumns.length > 0
                ? `${contextColumns.length} selected`
                : "Select context columns..."}
              <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-full p-0">
            <Command>
              <CommandInput placeholder="Search context columns..." />
              <CommandList>
                <CommandEmpty>No column found.</CommandEmpty>
                <CommandGroup>
                  {allColumns.map((column) => (
                    <CommandItem
                      key={column.id}
                      value={column.id}
                      onSelect={() => toggleContextColumn(column.id)}
                    >
                      <Check
                        className={cn(
                          "mr-2 h-4 w-4",
                          contextColumns.includes(column.id)
                            ? "opacity-100"
                            : "opacity-0",
                        )}
                      />
                      {column.name}
                    </CommandItem>
                  ))}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
        {contextColumns.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {contextColumns.map((column) => (
              <div
                key={column}
                className="flex items-center rounded-md bg-muted px-2 py-1 text-sm"
              >
                {allColumns.find((c) => c.id === column)?.name}
                <button
                  className="ml-1 rounded-full text-muted-foreground hover:text-foreground"
                  onClick={() => toggleContextColumn(column)}
                >
                  ×
                </button>
              </div>
            ))}
          </div>
        )}
        <div className="space-y-2">
          <Label htmlFor="autofillPrompt">Prompt</Label>
          <MentionInput
            id="autofillPrompt"
            value={prompt}
            onChange={(e) => {
              onPromptChange(e.target.value);
            }}
            textarea={true}
            rows={3}
            placeholder="Optional prompt for autofill. This is especially helpful when your columns and context columns are the same, since Tablepilot can only infer your intent from the column descriptions, this prompt serves as an additional guide."
            variables={variables}
          />
        </div>
      </div>
    </div>
  );
}
