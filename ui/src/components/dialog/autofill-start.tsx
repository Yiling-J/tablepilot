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
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { Check, ChevronsUpDown } from "lucide-react";
import { useState } from "react";

interface AutofillDialogProps {
  isOpen: boolean;
  setIsOpen: (v: boolean) => void;
  columns: Column[];
  onStart: (columns: Column[], contextColumns: Column[]) => void;
}

export function AutofillDialog({
  isOpen,
  setIsOpen,
  columns,
  onStart,
}: AutofillDialogProps) {
  const [selectedColumns, setSelectedColumns] = useState<Column[]>([]);
  const [selectedContextColumns, setSelectedContextColumns] = useState<
    Column[]
  >([]);

  const [columnsOpen, setColumnsOpen] = useState(false);
  const [contextColumnsOpen, setContextColumnsOpen] = useState(false);

  // Available options for each select (excluding what's selected in the other)
  const availableColumns = columns.filter(
    (col) => !selectedContextColumns.includes(col),
  );
  const availableContextColumns = columns.filter(
    (col) => !selectedColumns.includes(col),
  );

  const toggleColumn = (column: Column) => {
    setSelectedColumns((current) =>
      current.includes(column)
        ? current.filter((c) => c !== column)
        : [...current, column],
    );
  };

  const toggleContextColumn = (column: Column) => {
    setSelectedContextColumns((current) =>
      current.includes(column)
        ? current.filter((c) => c !== column)
        : [...current, column],
    );
  };

  const addAllColumns = () => {
    setSelectedColumns(availableColumns);
  };

  const addAllContextColumns = () => {
    setSelectedContextColumns(availableContextColumns);
  };

  const handleStart = () => {
    console.log("Starting with selections:", {
      columns: selectedColumns,
      contextColumns: selectedContextColumns,
    });
    // Close the dialog after starting
    setColumnsOpen(false);
    setContextColumnsOpen(false);
    onStart(selectedColumns, selectedContextColumns);
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Autofill Configuration</DialogTitle>
        </DialogHeader>

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
            <Popover
              open={columnsOpen}
              onOpenChange={setColumnsOpen}
              modal={true}
            >
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={columnsOpen}
                  className="w-full justify-between"
                >
                  {selectedColumns.length > 0
                    ? `${selectedColumns.length} selected`
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
                      {availableColumns.map((column) => (
                        <CommandItem
                          key={column.id}
                          value={column.id}
                          onSelect={() => toggleColumn(column)}
                        >
                          <Check
                            className={cn(
                              "mr-2 h-4 w-4",
                              selectedColumns.includes(column)
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
            {selectedColumns.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1">
                {selectedColumns.map((column) => (
                  <div
                    key={column.id}
                    className="flex items-center rounded-md bg-muted px-2 py-1 text-sm"
                  >
                    {column.name}
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
                  {selectedContextColumns.length > 0
                    ? `${selectedContextColumns.length} selected`
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
                      {availableContextColumns.map((column) => (
                        <CommandItem
                          key={column.id}
                          value={column.id}
                          onSelect={() => toggleContextColumn(column)}
                        >
                          <Check
                            className={cn(
                              "mr-2 h-4 w-4",
                              selectedContextColumns.includes(column)
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
            {selectedContextColumns.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1">
                {selectedContextColumns.map((column) => (
                  <div
                    key={column.id}
                    className="flex items-center rounded-md bg-muted px-2 py-1 text-sm"
                  >
                    {column.name}
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
          </div>
        </div>
        <DialogFooter className="mt-6">
          <DialogClose asChild>
            <Button variant="outline" type="button">
              Cancel
            </Button>
          </DialogClose>
          <Button type="button" onClick={handleStart}>
            Start
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
