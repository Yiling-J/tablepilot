import {
    AiSource,
    LinkedSource,
    ListSource,
    Source,
    TableCreateRequest,
    TableInfo,
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

import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Edit, Plus, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { ContextVariable, MentionInput } from "../ui/var-input";

interface SourcesFormProps {
  formData: TableCreateRequest;
  updateFormData: (data: Partial<TableCreateRequest>) => void;
  variables?: ContextVariable[];
  tables?: TableInfo[];
}

type SourceType = "ai" | "list" | "linked";

export function SourcesForm({
  formData,
  updateFormData,
  variables,
  tables: tablesProp,
}: SourcesFormProps) {
  const [sourceName, setSourceName] = useState("");
  const [sourceType, setSourceType] = useState<SourceType>("ai");
  const [prompt, setPrompt] = useState("");
  const [options, setOptions] = useState("");

  // Add new state for linked source type
  const [tables, setTables] = useState<TableInfo[]>([]);
  const [isLoadingTables, setIsLoadingTables] = useState(false);
  const [selectedTable, setSelectedTable] = useState<string>("");

  const [editIndex, setEditIndex] = useState<number | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  useEffect(() => {
    if (tables.length === 0) {
      fetchTables();
    }
  }, []);

  // Fetch tables from API
  const fetchTables = async () => {
    setIsLoadingTables(true);
    try {
      if (tablesProp) {
        setTables(tablesProp);
      } else {
        const resp = await getTables();
        setTables(resp.tables);
      }
      setIsLoadingTables(false);
    } catch (error) {
      console.error("Error fetching tables:", error);
    } finally {
      setIsLoadingTables(false);
    }
  };

  const changeSelectedTable = (selected: string) => {
    const table = tables.find((t) => t.name === selected);
    if (table) {
      setSelectedTable(selected);
    }
  };

  const resetForm = () => {
    setSourceName("");
    setSourceType("ai");
    setPrompt("");
    setOptions("");
    // Reset linked source fields
    setSelectedTable("");
    setEditIndex(null);
  };

  const handleAddSource = () => {
    let newSource: Source;

    if (sourceType === "ai") {
      newSource = {
        name: sourceName,
        type: "ai",
        prompt,
      };
    } else if (sourceType === "list") {
      newSource = {
        name: sourceName,
        type: "list",
        options: options.split("\n").filter(Boolean),
      };
    } else {
      newSource = {
        name: sourceName,
        type: "linked",
        table: selectedTable,
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
    } else if (source.type === "list") {
      const listSource = source as ListSource;
      setOptions(listSource.options.join("\n"));
    } else if (source.type === "linked") {
      const linkedSource = source as LinkedSource;
      setSelectedTable(linkedSource.table);
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
    if (sourceType === "linked" && !selectedTable) return false;

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
                  <MentionInput
                    id="prompt"
                    variables={variables}
                    placeholder="e.g., Generate 20 recipe cuisines."
                    value={prompt}
                    onChange={(e) => setPrompt(e.target.value)}
                    rows={3}
                    textarea={true}
                  />
                </div>
              )}

              {sourceType === "list" && (
                <div className="grid gap-2">
                  <Label htmlFor="options">Options (one per line)</Label>
                  <MentionInput
                    id="options"
                    variables={variables}
                    placeholder="e.g., Dinner&#10;Breakfast&#10;Lunch"
                    value={options}
                    onChange={(e) => setOptions(e.target.value)}
                    rows={5}
                    textarea={true}
                  />
                </div>
              )}

              {sourceType === "linked" && (
                <>
                  <div className="grid gap-2">
                    <Label htmlFor="table">Select Table</Label>
                    <Select
                      value={selectedTable}
                      onValueChange={changeSelectedTable}
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
                </>
              )}
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
              <CardHeader className="py-1 px-6">
                <div className="flex justify-between items-center">
                  <CardTitle className="text-base font-medium">
                    {source.name}
                  </CardTitle>
                  <div className="flex gap-2" data-testid="source-ops">
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
              <CardContent className="pt-0 pb-2 px-6">
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
