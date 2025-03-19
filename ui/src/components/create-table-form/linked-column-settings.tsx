import { Column } from "@/actions";
import { Badge } from "@/components/ui/badge";
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { Check, ChevronsUpDown, X } from "lucide-react";
import { useState } from "react";

export function LinkedColumnSettings({
  linkedTableColumns,
  selectedColumn,
  setSelectedColumn,
  selectedContextColumns,
  setSelectedContextColumns,
}: {
  linkedTableColumns: Column[];
  selectedColumn: string;
  setSelectedColumn: React.Dispatch<React.SetStateAction<string>>;
  selectedContextColumns: string[];
  setSelectedContextColumns: React.Dispatch<React.SetStateAction<string[]>>;
}) {
  const [openContextColumnsPopover, setOpenContextColumnsPopover] =
    useState(false);

  return (
    <>
      <div className="grid gap-2 pt-2">
        <Label htmlFor="column">Linked Column</Label>
        <Select
          value={selectedColumn}
          onValueChange={setSelectedColumn}
          disabled={linkedTableColumns.length === 0}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select a column" />
          </SelectTrigger>
          <SelectContent>
            {linkedTableColumns.map((column) => (
              <SelectItem key={column.id} value={column.name}>
                {column.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-2 pt-2">
        <Label>Linked Context Columns</Label>
        <Popover
          open={openContextColumnsPopover}
          onOpenChange={setOpenContextColumnsPopover}
        >
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              role="combobox"
              aria-expanded={openContextColumnsPopover}
              className="justify-between h-auto min-h-10"
            >
              {selectedContextColumns.length > 0 ? (
                <div className="flex flex-wrap gap-1">
                  {selectedContextColumns.map((column) => (
                    <Badge
                      key={column}
                      variant="secondary"
                      className="mr-1 mb-1"
                    >
                      {column}
                      <Button
                        asChild
                        variant="ghost"
                        size="icon"
                        className="ml-1 ring-offset-background rounded-full outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                        onKeyDown={(e) => {
                          if (e.key === "Enter") {
                            setSelectedContextColumns(
                              selectedContextColumns.filter(
                                (c) => c !== column,
                              ),
                            );
                          }
                        }}
                        onMouseDown={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          setSelectedContextColumns(
                            selectedContextColumns.filter((c) => c !== column),
                          );
                        }}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </Badge>
                  ))}
                </div>
              ) : (
                "Select context columns"
              )}
              <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="p-0 w-[300px]">
            <Command>
              <CommandInput placeholder="Search columns..." />
              <CommandList>
                <CommandEmpty>No columns found.</CommandEmpty>
                <CommandGroup>
                  {linkedTableColumns.map((column) => (
                    <CommandItem
                      key={column.id}
                      value={column.name}
                      onSelect={() => {
                        setSelectedContextColumns((prev) =>
                          prev.includes(column.name)
                            ? prev.filter((c) => c !== column.name)
                            : [...prev, column.name],
                        );
                      }}
                    >
                      <div
                        className={cn(
                          "mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary",
                          selectedContextColumns.includes(column.name)
                            ? "bg-primary text-primary-foreground"
                            : "opacity-50 [&_svg]:invisible",
                        )}
                      >
                        <Check className="h-4 w-4" />
                      </div>
                      {column.name}
                    </CommandItem>
                  ))}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
      </div>
    </>
  );
}
