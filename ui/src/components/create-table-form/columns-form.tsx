"use client";

import { TableCreateRequest } from "@/actions";
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
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { Edit, Plus, Trash2 } from "lucide-react";
import { useState } from "react";

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
    setIsDialogOpen(true);
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
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>
                {editIndex !== null ? "Edit Column" : "Add New Column"}
              </DialogTitle>
              <DialogDescription>
                Define a column for your table
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
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
                    <SelectItem value="user">User Input</SelectItem>
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
                    onValueChange={setSource}
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
              <CardContent className="py-2 px-6">
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
