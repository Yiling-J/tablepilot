import {
    AiSource,
    Column,
    LinkedSource,
    ListSource,
    Source,
    TableCreateRequest,
    TableInfo,
    getTables,
} from "@/actions";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
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
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { Check, ChevronsUpDown, Edit, Plus, Trash2, X } from "lucide-react";
import { useEffect, useState } from "react";

interface SourcesFormProps {
  formData: TableCreateRequest;
  updateFormData: (data: Partial<TableCreateRequest>) => void;
}

interface SourcesFormProps {
  formData: TableCreateRequest;
  updateFormData: (data: Partial<TableCreateRequest>) => void;
}

type SourceType = "ai" | "list" | "linked";

export function SourcesForm({ formData, updateFormData }: SourcesFormProps) {
  const [sourceName, setSourceName] = useState("");
  const [sourceType, setSourceType] = useState<SourceType>("ai");
  const [prompt, setPrompt] = useState("");
  const [options, setOptions] = useState("");
  const [random, setRandom] = useState(true);
  const [replacement, setReplacement] = useState(false);

  // Add new state for linked source type
  const [tables, setTables] = useState<TableInfo[]>([]);
  const [isLoadingTables, setIsLoadingTables] = useState(false);
  const [selectedTable, setSelectedTable] = useState<string>("");
  const [tableColumns, setTableColumns] = useState<Column[]>([]);
  const [selectedColumn, setSelectedColumn] = useState<string>("");
  const [selectedContextColumns, setSelectedContextColumns] = useState<
    string[]
  >([]);
  const [openContextColumnsPopover, setOpenContextColumnsPopover] =
    useState(false);

  const [editIndex, setEditIndex] = useState<number | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  // Fetch tables when dialog opens
  useEffect(() => {
    if (isDialogOpen && sourceType === "linked" && tables.length === 0) {
      fetchTables();
    }
  }, [isDialogOpen, sourceType]);

  // Fetch tables from API
  const fetchTables = async () => {
    setIsLoadingTables(true);
    try {
      const resp = await getTables();
      setTables(resp.tables);
      setIsLoadingTables(false);
    } catch (error) {
      console.error("Error fetching tables:", error);
    } finally {
      setIsLoadingTables(false);
    }
  };

  // Update table columns when a table is selected
  useEffect(() => {
    if (selectedTable) {
      const table = tables.find((t) => t.name === selectedTable);
      if (table) {
        setTableColumns(table.columns);
        // Reset column selections when table changes
        setSelectedColumn("");
        setSelectedContextColumns([]);
      }
    }
  }, [selectedTable, tables]);

  const resetForm = () => {
    setSourceName("");
    setSourceType("ai");
    setPrompt("");
    setOptions("");
    setRandom(true);
    setReplacement(false);
    // Reset linked source fields
    setSelectedTable("");
    setSelectedColumn("");
    setSelectedContextColumns([]);
    setTableColumns([]);
    setEditIndex(null);
  };

  const handleAddSource = () => {
    let newSource: Source;

    if (sourceType === "ai") {
      newSource = {
        name: sourceName,
        type: "ai",
        prompt,
        random,
        replacement,
      };
    } else if (sourceType === "list") {
      newSource = {
        name: sourceName,
        type: "list",
        options: options.split("\n").filter(Boolean),
        random,
        replacement,
      };
    } else {
      newSource = {
        name: sourceName,
        type: "linked",
        table: selectedTable,
        column: selectedColumn,
        context_columns: selectedContextColumns,
        random,
        replacement,
      };
    }

    let updatedSources = [...formData.sources];

    if (editIndex !== null) {
      updatedSources[editIndex] = newSource;
    } else {
      updatedSources = [...updatedSources, newSource];
    }

    updateFormData({ sources: updatedSources });
    resetForm();
    setIsDialogOpen(false);
  };

  const handleEditSource = (index: number) => {
    const source = formData.sources[index];
    setSourceName(source.name);
    setSourceType(source.type as SourceType);

    if (source.type === "ai") {
      const aiSource = source as AiSource;
      setPrompt(aiSource.prompt);
      setRandom(aiSource.random);
      setReplacement(aiSource.replacement);
    } else if (source.type === "list") {
      const listSource = source as ListSource;
      setOptions(listSource.options.join("\n"));
      setRandom(listSource.random);
      setReplacement(listSource.replacement);
    } else if (source.type === "linked") {
      const linkedSource = source as LinkedSource;
      setSelectedTable(linkedSource.table);
      setSelectedColumn(linkedSource.column);
      setSelectedContextColumns(linkedSource.context_columns || []);

      // Fetch tables if not already loaded
      if (tables.length === 0) {
        fetchTables();
      } else {
        // Set table columns based on the selected table
        const table = tables.find((t) => t.name === linkedSource.table);
        if (table) {
          setTableColumns(table.columns);
        }
      }
    }

    setEditIndex(index);
    setIsDialogOpen(true);
  };

  const handleDeleteSource = (index: number) => {
    const updatedSources = formData.sources.filter((_, i) => i !== index);
    updateFormData({ sources: updatedSources });
  };

  const isSourceValid = () => {
    if (!sourceName) return false;

    if (sourceType === "ai" && !prompt) return false;
    if (sourceType === "list" && !options) return false;
    if (sourceType === "linked" && (!selectedTable || !selectedColumn))
      return false;

    return true;
  };

  const renderSourceDetails = (source: Source) => {
    if (source.type === "ai") {
      const aiSource = source as AiSource;
      return (
        <>
          <div className="flex gap-2 mb-1">
            <span className="font-medium">Prompt:</span>
            <span>{aiSource.prompt}</span>
          </div>
        </>
      );
    } else if (source.type === "list") {
      const listSource = source as ListSource;
      return (
        <>
          <div className="flex gap-2 mb-1">
            <span className="font-medium">Options:</span>
            <span>{listSource.options.join(", ")}</span>
          </div>
        </>
      );
    } else if (source.type === "linked") {
      const linkedSource = source as LinkedSource;
      return (
        <>
          <div className="flex gap-2 mb-1">
            <span className="font-medium">Table:</span>
            <span>{linkedSource.table}</span>
          </div>
          <div className="flex gap-2 mb-1">
            <span className="font-medium">Column:</span>
            <span>{linkedSource.column}</span>
          </div>
          {linkedSource.context_columns &&
            linkedSource.context_columns.length > 0 && (
              <div className="flex gap-2 mb-1">
                <span className="font-medium">Context Columns:</span>
                <span>{linkedSource.context_columns.join(", ")}</span>
              </div>
            )}
        </>
      );
    }

    return null;
  };

  return (
    <div className="space-y-4 py-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-medium">Data Sources</h3>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button
              onClick={() => {
                resetForm();
                setIsDialogOpen(true);
              }}
            >
              <Plus className="mr-2 h-4 w-4" /> Add Source
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>
                {editIndex !== null ? "Edit Source" : "Add New Source"}
              </DialogTitle>
              <DialogDescription>
                Define a data source for your table columns
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="sourceName">Source Name</Label>
                <Input
                  id="sourceName"
                  placeholder="e.g., cuisines, meals, customer"
                  value={sourceName}
                  onChange={(e) => setSourceName(e.target.value)}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="sourceType">Source Type</Label>
                <Select
                  value={sourceType}
                  onValueChange={(value: SourceType) => setSourceType(value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="ai">AI Generated</SelectItem>
                    <SelectItem value="list">List of Options</SelectItem>
                    <SelectItem value="linked">Linked Table</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {sourceType === "ai" && (
                <div className="grid gap-2">
                  <Label htmlFor="prompt">AI Prompt</Label>
                  <Textarea
                    id="prompt"
                    placeholder="e.g., Generate 20 recipe cuisines."
                    value={prompt}
                    onChange={(e) => setPrompt(e.target.value)}
                    rows={3}
                  />
                </div>
              )}

              {sourceType === "list" && (
                <div className="grid gap-2">
                  <Label htmlFor="options">Options (one per line)</Label>
                  <Textarea
                    id="options"
                    placeholder="e.g., Dinner&#10;Breakfast&#10;Lunch"
                    value={options}
                    onChange={(e) => setOptions(e.target.value)}
                    rows={5}
                  />
                </div>
              )}

              {sourceType === "linked" && (
                <>
                  <div className="grid gap-2">
                    <Label htmlFor="table">Select Table</Label>
                    <Select
                      value={selectedTable}
                      onValueChange={setSelectedTable}
                      disabled={isLoadingTables || tables.length === 0}
                    >
                      <SelectTrigger>
                        <SelectValue
                          placeholder={
                            isLoadingTables
                              ? "Loading tables..."
                              : "Select a table"
                          }
                        />
                      </SelectTrigger>
                      <SelectContent>
                        {tables.map((table) => (
                          <SelectItem key={table.name} value={table.name}>
                            {table.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {selectedTable && (
                    <>
                      <div className="grid gap-2">
                        <Label htmlFor="column">Select Column</Label>
                        <Select
                          value={selectedColumn}
                          onValueChange={setSelectedColumn}
                          disabled={tableColumns.length === 0}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Select a column" />
                          </SelectTrigger>
                          <SelectContent>
                            {tableColumns.map((column) => (
                              <SelectItem key={column.id} value={column.name}>
                                {column.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="grid gap-2">
                        <Label>Context Columns</Label>
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
                                      <button
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
                                            selectedContextColumns.filter(
                                              (c) => c !== column,
                                            ),
                                          );
                                        }}
                                      >
                                        <X className="h-3 w-3" />
                                      </button>
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
                                  {tableColumns.map((column) => (
                                    <CommandItem
                                      key={column.id}
                                      value={column.name}
                                      onSelect={() => {
                                        setSelectedContextColumns((prev) =>
                                          prev.includes(column.name)
                                            ? prev.filter(
                                                (c) => c !== column.name,
                                              )
                                            : [...prev, column.name],
                                        );
                                      }}
                                    >
                                      <div
                                        className={cn(
                                          "mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary",
                                          selectedContextColumns.includes(
                                            column.name,
                                          )
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
                        <p className="text-xs text-muted-foreground">
                          Select columns to provide context when using this
                          source
                        </p>
                      </div>
                    </>
                  )}
                </>
              )}

              <>
                <div className="flex items-center space-x-2">
                  <Switch
                    id="random"
                    checked={random}
                    onCheckedChange={setRandom}
                  />
                  <Label htmlFor="random">Random Selection</Label>
                </div>

                <div className="flex items-center space-x-2">
                  <Switch
                    id="replacement"
                    checked={replacement}
                    onCheckedChange={setReplacement}
                  />
                  <Label htmlFor="replacement">
                    Selection with Replacement
                  </Label>
                </div>
              </>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleAddSource} disabled={!isSourceValid()}>
                {editIndex !== null ? "Update" : "Add"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {formData.sources.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">
          No sources added yet. Click the "Add Source" button to create one.
        </div>
      ) : (
        <div className="grid gap-4 mt-4">
          {formData.sources.map((source, index) => (
            <Card key={index}>
              <CardHeader className="py-4 px-6">
                <div className="flex justify-between items-center">
                  <CardTitle className="text-base font-medium">
                    {source.name}
                  </CardTitle>
                  <div className="flex gap-2">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleEditSource(index)}
                    >
                      <Edit className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDeleteSource(index)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="py-2 px-6">
                <div className="text-sm">
                  <div className="flex gap-2 mb-1">
                    <span className="font-medium">Type:</span>
                    <span>
                      {source.type === "ai"
                        ? "AI Generated"
                        : source.type === "list"
                          ? "List of Options"
                          : "Linked Table"}
                    </span>
                  </div>

                  {renderSourceDetails(source)}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
