import {
    Column,
    LinkedSource,
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
import { NumberInput } from "@/components/ui/number-input";
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

interface ColumnsFormProps {
  formData: TableCreateRequest;
  updateFormData: (data: Partial<TableCreateRequest>) => void;
  disabled: boolean;
}

export function ColumnsForm({
  formData,
  updateFormData,
  disabled,
}: ColumnsFormProps) {
  const [columnName, setColumnName] = useState("");
  const [columnDescription, setColumnDescription] = useState("");
  const [columnType, setColumnType] = useState("string");
  const [fillMode, setFillMode] = useState("ai");
  const [contextLength, setContextLength] = useState<number | undefined>(
    undefined,
  );
  const [source, setSource] = useState<string | undefined>(undefined);
  const [editIndex, setEditIndex] = useState<number | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [random, setRandom] = useState(true);
  const [replacement, setReplacement] = useState(false);
  const [repeat, setRepeat] = useState(1);
  const [selectedColumn, setSelectedColumn] = useState<string>("");
  const [tables, setTables] = useState<TableInfo[]>([]);
  const [linkedTableColumns, setLinkedTableColumns] = useState<Column[]>([]);
  const [selectedContextColumns, setSelectedContextColumns] = useState<
    string[]
  >([]);
  const [openContextColumnsPopover, setOpenContextColumnsPopover] =
    useState(false);

  const resetForm = () => {
    setColumnName("");
    setColumnDescription("");
    setColumnType("string");
    setFillMode("ai");
    setContextLength(undefined);
    setSource(undefined);
    setEditIndex(null);
    setRandom(true);
    setReplacement(false);
    setRepeat(1);
    setSelectedColumn("");
    setSelectedContextColumns([]);
  };

  useEffect(() => {
    if (tables.length === 0) {
      fetchTables();
    }
  }, []);

  // Fetch tables from API
  const fetchTables = async () => {
    try {
      const resp = await getTables();
      setTables(resp.tables);
    } catch (error) {
      console.error("Error fetching tables:", error);
    }
  };

  const handleAddColumn = () => {
    const newColumn = {
      name: columnName,
      description: columnDescription,
      type: columnType,
      fill_mode: fillMode,
      random: random,
      replacement: replacement,
      repeat: repeat,
      linked_column: selectedColumn,
      linked_context_columns: selectedContextColumns,
      ...(contextLength ? { context_length: contextLength } : {}),
      ...(source ? { source } : {}),
    };

    let updatedColumns = [...formData.columns];

    if (editIndex !== null) {
      updatedColumns[editIndex] = newColumn;
    } else {
      updatedColumns = [...updatedColumns, newColumn];
    }

    updateFormData({ columns: updatedColumns });
    resetForm();
    setIsDialogOpen(false);
  };

  const handleEditColumn = (index: number) => {
    const column = formData.columns[index];
    setColumnName(column.name);
    setColumnDescription(column.description);
    setColumnType(column.type);
    setFillMode(column.fill_mode);
    setContextLength(column.context_length);
    setSource(column.source);
    setEditIndex(index);
    setRandom(column.random);
    setReplacement(column.replacement);
    setRepeat(column.repeat);
    setSelectedColumn(column.linked_column);
    setSelectedContextColumns(column.linked_context_columns || []);
    if (column.source) {
      const source = formData.sources.find((s) => s.name == column.source);
      if (source && source.type === "linked") {
        const linkedSource = source as LinkedSource;
        const table = tables.find((t) => t.name === linkedSource.table);
        if (table) {
          setLinkedTableColumns(table.columns);
        }
      }
    }
    setIsDialogOpen(true);
  };

  const handleSelectSource = (source: string | undefined) => {
    setSource(source);
    setSelectedColumn("");
    setSelectedContextColumns([]);
    if (source) {
      const ls = formData.sources.find((s) => s.name === source);
      if (ls && ls.type === "linked") {
        const linkedSource = ls as LinkedSource;
        const table = tables.find((t) => t.name === linkedSource.table);
        if (table) {
          setLinkedTableColumns(table.columns);
        }
      }
    }
  };

  const handleDeleteColumn = (index: number) => {
    const updatedColumns = formData.columns.filter((_, i) => i !== index);
    updateFormData({ columns: updatedColumns });
  };

  return (
    <div className="space-y-4 py-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-medium">Table Columns</h3>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button
              disabled={disabled}
              onClick={() => {
                resetForm();
                setIsDialogOpen(true);
              }}
            >
              <Plus className="mr-2 h-4 w-4" /> Add Column
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[530px] scrollbar-thumb-rounded-full scrollbar-track-rounded-full scrollbar scrollbar-thumb-stone-500 scrollbar-track-background">
            <DialogHeader>
              <DialogTitle>
                {editIndex !== null ? "Edit Column" : "Add New Column"}
              </DialogTitle>
              <DialogDescription>
                Define a column for your table
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4 px-2 max-h-[65vh] overflow-auto scrollbar-thin">
              <div className="grid gap-2">
                <Label htmlFor="columnName">Column Name</Label>
                <Input
                  id="columnName"
                  placeholder="e.g., Name, Ingredients"
                  value={columnName}
                  onChange={(e) => setColumnName(e.target.value)}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="columnDescription">Description</Label>
                <Textarea
                  id="columnDescription"
                  placeholder="e.g., recipe name, list of ingredients"
                  value={columnDescription}
                  onChange={(e) => setColumnDescription(e.target.value)}
                  rows={2}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="columnType">Data Type</Label>
                <Select value={columnType} onValueChange={setColumnType}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="string">String</SelectItem>
                    <SelectItem value="integer">Integer</SelectItem>
                    <SelectItem value="number">Float or Integer</SelectItem>
                    <SelectItem value="boolean">Boolean</SelectItem>
                    <SelectItem value="array">Array</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="fillMode">Fill Mode</Label>
                <Select value={fillMode} onValueChange={setFillMode}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select fill mode" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="ai">AI Generated</SelectItem>
                    <SelectItem value="pick">Pick from Source</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {fillMode === "ai" && (
                <div className="grid gap-2">
                  <Label htmlFor="contextLength">
                    Context Length (optional)
                  </Label>
                  <Input
                    id="contextLength"
                    type="number"
                    placeholder="e.g., 5"
                    value={contextLength || ""}
                    onChange={(e) =>
                      setContextLength(
                        e.target.value
                          ? Number.parseInt(e.target.value)
                          : undefined,
                      )
                    }
                  />
                </div>
              )}

              {fillMode === "pick" && (
                <div className="grid gap-2">
                  <Label htmlFor="source">Source</Label>
                  <Select
                    value={source}
                    onValueChange={handleSelectSource}
                    disabled={formData.sources.length === 0}
                  >
                    <SelectTrigger>
                      <SelectValue
                        placeholder={
                          formData.sources.length === 0
                            ? "No sources available"
                            : "Select source"
                        }
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {formData.sources.map((src, index) => (
                        <SelectItem key={index} value={src.name}>
                          {src.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>

                  {formData.sources.find((s) => s.name === source)?.type ===
                    "linked" && (
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
                                  {linkedTableColumns.map((column) => (
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
                      </div>
                    </>
                  )}

                  <div className="flex items-center space-x-2 pt-2">
                    <Switch
                      id="random"
                      checked={random}
                      onCheckedChange={setRandom}
                    />
                    <Label htmlFor="random">Random Selection</Label>
                  </div>

                  <div className="flex items-center space-x-2 pt-2">
                    <Switch
                      id="replacement"
                      checked={replacement}
                      onCheckedChange={setReplacement}
                    />
                    <Label htmlFor="replacement">
                      Selection with Replacement
                    </Label>
                  </div>

                  <div className="flex items-center space-x-2 pt-2">
                    <NumberInput
                      id="repeat"
                      value={repeat > 0 ? repeat : 1}
                      onValueChange={(v) => setRepeat(v ?? 1)}
                    />
                    <Label htmlFor="reppeat">Repeat selection</Label>
                  </div>
                  {formData.sources.length === 0 && (
                    <p className="text-xs text-muted-foreground mt-1">
                      Add sources in the previous step to use this fill mode
                    </p>
                  )}
                </div>
              )}
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleAddColumn}
                disabled={fillMode === "pick" && !source}
              >
                {editIndex !== null ? "Update" : "Add"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {formData.columns.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">
          No columns added yet. Click the "Add Column" button to create one.
        </div>
      ) : (
        <div className="grid gap-4 mt-4">
          {formData.columns.map((column, index) => (
            <Card key={index}>
              <CardHeader className="py-1 px-6">
                <div className="flex justify-between items-center">
                  <CardTitle className="text-base font-medium">
                    {column.name}
                  </CardTitle>
                  <div className="flex gap-2">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleEditColumn(index)}
                      disabled={disabled}
                    >
                      <Edit className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDeleteColumn(index)}
                      disabled={disabled}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="pt-0 pb-2 px-6">
                <div className="text-sm">
                  <div className="mb-1">{column.description}</div>
                  <div className="flex gap-4 mt-2 text-xs text-muted-foreground">
                    <span>Type: {column.type}</span>
                    <span>Fill: {column.fill_mode}</span>
                    {column.context_length && (
                      <span>Context: {column.context_length}</span>
                    )}
                    {column.source && <span>Source: {column.source}</span>}
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
