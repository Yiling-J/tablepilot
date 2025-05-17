import { WorkflowVariable } from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
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
import { JSONObject, JSONValue } from "@/json";
import { useEffect, useState } from "react";

interface VariablesDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  variables: WorkflowVariable[];
  onSave: (values: JSONObject) => void;
}

export function VariablesDialog({
  open,
  onOpenChange,
  variables,
  onSave,
}: VariablesDialogProps) {
  const [values, setValues] = useState<Record<string, JSONValue>>({});
  const [files, setFiles] = useState<Record<string, File | null>>({});

  if (!variables || variables.length === 0) {
    return <div></div>;
  }
  // Initialize values with default values
  useEffect(() => {
    const initialValues: Record<string, JSONValue> = {};
    const initialFiles: Record<string, File | null> = {};

    variables.forEach((variable) => {
      if (variable.type === "file") {
        initialFiles[variable.name] = null;
      } else {
        initialValues[variable.name] = variable.default_value;
      }
    });

    setValues(initialValues);
    setFiles(initialFiles);
  }, [variables]);

  const handleInputChange = (name: string, value: JSONValue) => {
    setValues((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleFileChange = (name: string, file: File | null) => {
    setFiles((prev) => ({
      ...prev,
      [name]: file,
    }));
  };

  const handleSubmit = () => {
    // Combine regular values and file information
    const combinedValues = { ...values };

    // Add file names for file type variables
    Object.entries(files).forEach(([name, file]) => {
      if (!file) {
        return;
      }
      const reader = new FileReader();
      reader.readAsDataURL(file);
      combinedValues[name] = reader.result as string;
    });
    onSave(combinedValues);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Input Variables</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          {variables.map((variable) => (
            <div key={variable.name} className="grid gap-2">
              <Label htmlFor={variable.name} className="text-left">
                {variable.name}
              </Label>
              <div>
                {renderInputForVariable(
                  variable,
                  values[variable.name],
                  files[variable.name],
                  (value) => handleInputChange(variable.name, value),
                  (file) => handleFileChange(variable.name, file),
                )}
              </div>
            </div>
          ))}
        </div>
        <DialogFooter>
          <Button type="button" onClick={handleSubmit}>
            Start
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function renderInputForVariable(
  variable: WorkflowVariable,
  value: JSONValue,
  file: File | null,
  onChange: (value: JSONValue) => void,
  onFileChange: (file: File | null) => void,
) {
  // Handle file type
  if (variable.type === "file") {
    return (
      <div className="flex flex-col gap-2">
        <Input
          id={variable.name}
          type="file"
          onChange={(e) => {
            const selectedFile = e.target.files?.[0] || null;
            onFileChange(selectedFile);
          }}
        />
        {file && <p className="text-sm text-gray-500">Selected: {file.name}</p>}
      </div>
    );
  }

  // Handle options for select dropdown
  if (variable.options && variable.options.length > 0) {
    return (
      <Select
        value={value?.toString()}
        onValueChange={(newValue) => {
          // Convert to appropriate type
          let typedValue: string | number = newValue;
          if (variable.type === "integer") {
            typedValue = Number.parseInt(newValue, 10);
          } else if (variable.type === "number") {
            typedValue = Number.parseFloat(newValue);
          }
          onChange(typedValue);
        }}
      >
        <SelectTrigger id={variable.name}>
          <SelectValue placeholder={`Select ${variable.name}`} />
        </SelectTrigger>
        <SelectContent>
          {variable.options.map((option) => (
            <SelectItem key={option.toString()} value={option.toString()}>
              {option.toString()}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    );
  }

  // Handle text/number input
  if (variable.type === "integer") {
    return (
      <Input
        id={variable.name}
        type="number"
        step="1"
        value={value || ""}
        onChange={(e) => onChange(Number.parseInt(e.target.value, 10) || 0)}
      />
    );
  }

  if (variable.type === "number") {
    return (
      <Input
        id={variable.name}
        type="number"
        step="any"
        value={value || ""}
        onChange={(e) => onChange(Number.parseFloat(e.target.value) || 0)}
      />
    );
  }

  // Default to string input
  return (
    <Input
      id={variable.name}
      type="text"
      value={value || ""}
      onChange={(e) => onChange(e.target.value)}
    />
  );
}
