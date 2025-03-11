import { TableCreateRequest } from "@/actions";
import { Button } from "@/components/ui/button";
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
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { Upload } from "lucide-react";
import { useEffect, useState } from "react";

interface NameDescriptionFormProps {
  formData: TableCreateRequest;
  updateFormData: (data: Partial<TableCreateRequest>) => void;
}

export function NameDescriptionForm({
  formData,
  updateFormData,
}: NameDescriptionFormProps) {
  const [jsonInput, setJsonInput] = useState("");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [parseError, setParseError] = useState<string | null>(null);
  const { toast } = useToast();
  const [error, setError] = useState("");
  const [name, setName] = useState("");

  const validateName = (name: string) => {
    if (!name) {
      setError("Table name cannot be empty.");
      return false;
    }
    const isValid = /^[a-zA-Z][a-zA-Z0-9_]*$/.test(name);
    setError(
      isValid
        ? ""
        : "Table name must start with a letter and contain only letters, numbers, or underscores.",
    );
    return isValid;
  };

  useEffect(() => {
    setName(formData.name);
    validateName(formData.name);
  }, []);

  const handleNameChange = (data: Partial<TableCreateRequest>) => {
    setName(data.name ?? "");
    const newName = data.name ?? "";
    if (validateName(newName)) {
      updateFormData({ name: newName });
    }
  };

  const handleParseJson = () => {
    try {
      setParseError(null);
      const parsedData = JSON.parse(jsonInput);

      // Validate the structure
      if (!parsedData.name || typeof parsedData.name !== "string") {
        throw new Error("JSON must contain a 'name' property of type string");
      }

      if (
        !parsedData.description ||
        typeof parsedData.description !== "string"
      ) {
        throw new Error(
          "JSON must contain a 'description' property of type string",
        );
      }

      if (!Array.isArray(parsedData.sources)) {
        throw new Error("JSON must contain a 'sources' array");
      }

      if (!Array.isArray(parsedData.columns)) {
        throw new Error("JSON must contain a 'columns' array");
      }

      // Update the form data with the parsed JSON
      updateFormData(parsedData);

      // Close the dialog
      setIsDialogOpen(false);

      // Show success toast
      toast({
        title: "JSON Parsed Successfully",
        description: "Your configuration has been loaded",
      });
    } catch (error) {
      setParseError(
        error instanceof Error ? error.message : "Invalid JSON format",
      );
    }
  };

  return (
    <div className="space-y-4 py-4">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-medium">Basic Information</h3>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button variant="outline">
              <Upload className="mr-2 h-4 w-4" /> Import JSON
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[600px]">
            <DialogHeader>
              <DialogTitle>Import Configuration</DialogTitle>
              <DialogDescription>
                Paste your JSON configuration to quickly fill the form
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <Textarea
                placeholder="Paste your JSON here..."
                value={jsonInput}
                onChange={(e) => setJsonInput(e.target.value)}
                rows={10}
                className="font-mono text-sm"
              />
              {parseError && (
                <div className="text-sm text-destructive">{parseError}</div>
              )}
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleParseJson}>Parse JSON</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <div className="space-y-2">
        <Label htmlFor="name">Table Name</Label>
        <Input
          id="name"
          placeholder="Only letters, numbers, and underscores, and start with a letter"
          value={name}
          onChange={(e) => handleNameChange({ name: e.target.value })}
        />
        {error && <p className="text-red-500 text-sm">{error}</p>}
      </div>
      <div className="space-y-2">
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          placeholder="Enter table description"
          value={formData.description}
          onChange={(e) => updateFormData({ description: e.target.value })}
          rows={3}
        />
      </div>
    </div>
  );
}
