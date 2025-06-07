import {
    Column,
    DatasetInfo,
    SourceType,
    // LinkedSource, // Removed as unused
    TableCreateRequest,
    TableInfo,
    getDatasets,
    getTables,
} from "@/actions";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import { Edit, Plus, Trash2, Wand2 } from "lucide-react";
import { useEffect, useState } from "react";
import { GenerateOptionsDialog } from "../dialog/generate-options-dialog";
import { ContextVariable, MentionInput } from "../ui/var-input";
import { LinkedColumnSettings } from "./linked-column-settings";

interface ColumnsFormProps {
  formData: TableCreateRequest;
  updateFormData: (data: Partial<TableCreateRequest>) => void;
  disabled: boolean;
  variables?: ContextVariable[];
  tables?: TableInfo[];
  privateDatasets?: DatasetInfo[];
}

export function ColumnsForm({
  formData,
  updateFormData,
  disabled,
  variables,
  tables: tablesProp,
  privateDatasets,
}: ColumnsFormProps) {
  const [columnName, setColumnName] = useState("");
  const [columnDescription, setColumnDescription] = useState("");
  const [columnType, setColumnType] = useState("string");
  const [fillMode, setFillMode] = useState("ai");
  const [contextLength, setContextLength] = useState<number | undefined>(
    undefined,
  );
  const [sourceID, setSourceID] = useState<string | undefined>(undefined);
  const [sourceType, setSourceType] = useState<SourceType | undefined>(
    undefined,
  );
  const [editIndex, setEditIndex] = useState<number | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [random, setRandom] = useState(true);
  const [replacement, setReplacement] = useState(false);
  const [repeat, setRepeat] = useState(1);
  const [selectedColumn, setSelectedColumn] = useState<string>("");
  const [tables, setTables] = useState<TableInfo[]>([]);
  const [datasets, setDatasets] = useState<DatasetInfo[]>([]);
  const [linkedTableColumns, setLinkedTableColumns] = useState<Column[]>([]);
  const [selectedContextColumns, setSelectedContextColumns] = useState<
    string[]
  >([]);
  const [listOptions, setListOptions] = useState("");
  const [isGenerateOptionsDialogOpen, setIsGenerateOptionsDialogOpen] =
    useState(false);

  const resetForm = () => {
    setColumnName("");
    setColumnDescription("");
    setColumnType("string");
    setFillMode("ai");
    setContextLength(undefined);
    setSourceID(undefined);
    setSourceType(undefined);
    setEditIndex(null);
    setRandom(true);
    setReplacement(false);
    setRepeat(1);
    setSelectedColumn("");
    setSelectedContextColumns([]);
    setListOptions("");
  };

  useEffect(() => {
    if (tables.length === 0) {
      fetchTables();
    }
    fetchDatasets();
  }, []);

  // Fetch tables from API
  const fetchTables = async () => {
    try {
      if (tablesProp) {
        setTables(tablesProp);
      } else {
        const resp = await getTables();
        setTables(resp.tables ?? []);
      }
    } catch (error) {
      console.error("Error fetching tables:", error);
    }
  };

  const fetchDatasets = async () => {
    try {
      const ds = await getDatasets();
      setDatasets([...(privateDatasets ?? []), ...ds.datasets]);
    } catch (error) {
      console.error("Error fetching datasets:", error);
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
      context_length: contextLength,
      source_id: sourceID,
      source_type: sourceType,
      options: listOptions.split("\n").filter((i) => i !== ""),
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
    setSourceID(column.source_id);
    setEditIndex(index);
    setRandom(column.random);
    setReplacement(column.replacement);
    setRepeat(column.repeat);
    setSelectedColumn(column.linked_column);
    setSelectedContextColumns(column.linked_context_columns || []);
    setIsDialogOpen(true);
    setListOptions((column.options ?? []).join("\n"));
  };

  const handleSelectSource = (source: string | undefined) => {
    setSourceID(source);
    setSelectedColumn("");
    setSelectedContextColumns([]);

    switch (sourceType) {
      case "table":
        const table = tables.find((t) => t.id === source);
        if (table) {
          setLinkedTableColumns(table.columns);
        }
        break;
      case "dataset":
        const ds = datasets.find((t) => t.id === source);
        if (ds && ds.type === "csv") {
          setLinkedTableColumns(
            ds.columns.map((c) => {
              return {
                id: c,
                name: c,
                description: "",
                type: "string",
                fill_mode: "ai",
              };
            }),
          );
        }
        break;
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
                <MentionInput
                  id="columnName"
                  placeholder="e.g., Name, Ingredients"
                  value={columnName}
                  variables={variables}
                  onChange={(e) => setColumnName(e.target.value)}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="columnDescription">Description</Label>
                <MentionInput
                  id="columnDescription"
                  placeholder="e.g., recipe name, list of ingredients"
                  value={columnDescription}
                  variables={variables}
                  textarea={true}
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
                    <SelectItem value="image">Image</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="fillMode">Fill Mode</Label>
                <Select
                  value={`${fillMode}-${sourceType}`}
                  onValueChange={(v) => {
                    switch (v) {
                      case "ai-undefined":
                        setFillMode("ai");
                        setSourceType(undefined);
                        break;
                      case "pick-table":
                        setFillMode("pick");
                        setSourceType("table");
                        break;
                      case "pick-dataset":
                        setFillMode("pick");
                        setSourceType("dataset");
                        break;
                      case "pick-options":
                        setFillMode("pick");
                        setSourceType("options");
                        break;
                    }
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select fill mode" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="ai-undefined">AI Generated</SelectItem>
                    <SelectItem value="pick-table">
                      Select from Table
                    </SelectItem>
                    <SelectItem value="pick-dataset">
                      Select from Dataset
                    </SelectItem>
                    <SelectItem value="pick-options">
                      Select from Options
                    </SelectItem>
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

              {fillMode === "pick" && sourceType === "table" && (
                <div className="grid gap-2">
                  <Label htmlFor="source">Table</Label>
                  <Select
                    value={sourceID}
                    onValueChange={handleSelectSource}
                    disabled={tables.length === 0}
                  >
                    <SelectTrigger>
                      <SelectValue
                        placeholder={
                          tables.length === 0
                            ? "No tables available"
                            : "Select a table"
                        }
                      />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {tables.map((tb, index) => (
                          <SelectItem key={`shared-${index}`} value={tb.id}>
                            {tb.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>

                  <LinkedColumnSettings
                    linkedTableColumns={linkedTableColumns}
                    selectedColumn={selectedColumn}
                    setSelectedColumn={setSelectedColumn}
                    selectedContextColumns={selectedContextColumns}
                    setSelectedContextColumns={setSelectedContextColumns}
                  />

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
                </div>
              )}

              {fillMode === "pick" && sourceType === "dataset" && (
                <div className="grid gap-2">
                  <Label htmlFor="source">Dataset</Label>
                  <Select
                    value={sourceID}
                    onValueChange={handleSelectSource}
                    disabled={datasets.length === 0}
                  >
                    <SelectTrigger>
                      <SelectValue
                        placeholder={
                          datasets.length === 0
                            ? "No datasets available"
                            : "Select a dataset"
                        }
                      />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {datasets.map((ds, index) => (
                          <SelectItem key={`shared-${index}`} value={ds.id}>
                            {ds.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <LinkedColumnSettings
                    linkedTableColumns={linkedTableColumns}
                    selectedColumn={selectedColumn}
                    setSelectedColumn={setSelectedColumn}
                    selectedContextColumns={selectedContextColumns}
                    setSelectedContextColumns={setSelectedContextColumns}
                  />
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
                </div>
              )}

              {fillMode === "pick" && sourceType === "options" && (
                <div>
                  <GenerateOptionsDialog
                    isOpen={isGenerateOptionsDialogOpen}
                    onClose={() => setIsGenerateOptionsDialogOpen(false)}
                    datasetName={columnName}
                    currentOptions={listOptions.split("\n")}
                    datasetDescription={columnDescription}
                    onGenerationComplete={(generatedOptions: string[]) => {
                      if (generatedOptions.length > 0) {
                        setListOptions((prevOptions) => {
                          const currentOptions = prevOptions.trim();
                          const newOptionsString = generatedOptions.join("\n");
                          if (currentOptions === "") {
                            return newOptionsString;
                          }
                          return currentOptions + "\n" + newOptionsString;
                        });
                      }
                      setIsGenerateOptionsDialogOpen(false);
                    }}
                  />
                  <Label htmlFor="list-options" className="text-right pt-2">
                    Options
                  </Label>
                  <div className="col-span-3 relative">
                    <MentionInput
                      variables={variables}
                      id="list-options"
                      value={listOptions}
                      onChange={(e) => {
                        setListOptions(e.target.value);
                      }}
                      textarea={true}
                      rows={4}
                      placeholder=""
                      className="pr-12 hide-scrollbar mt-2"
                    />
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            aria-label="wand-button"
                            variant="ghost"
                            size="icon"
                            onClick={() => setIsGenerateOptionsDialogOpen(true)}
                            className="absolute bottom-7 right-3 p-1.5 h-auto w-auto rounded-md bg-gradient-to-r from-orange-400 to-orange-600 text-white hover:opacity-80 hover:scale-105 transform transition-all"
                          >
                            <Wand2 className="size-5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Generate options with AI</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                    <p className="text-xs text-muted-foreground mt-1">
                      Each line will be treated as a separate option.
                    </p>
                  </div>
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
                </div>
              )}
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleAddColumn}
                disabled={fillMode === "pick" && !sourceID}
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
                  <div className="flex gap-2" data-testid="column-ops">
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
                    {column.source_id && (
                      <span>Source: {column.source_id}</span>
                    )}
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
