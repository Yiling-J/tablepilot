import { DatasetInfo } from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Textarea } from "@/components/ui/textarea";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import { Wand2 } from "lucide-react";
import React, { useEffect, useRef, useState } from "react";
import { GenerateOptionsDialog } from "../generate-options-dialog";

export interface CreateDatasetDialogProps {
  // Added export
  dataset?: DatasetInfo;
  isOpen: boolean;
  onClose: () => void;
  onCreate: (data: {
    name: string;
    description: string;
    type: "list" | "csv";
    options?: string[];
    files?: File[];
  }) => void;
  onUpdate: (
    id: string,
    data: {
      name: string;
      description: string;
      type: "list" | "csv";
      options?: string[];
      files?: File[];
    },
  ) => void;
}

type DatasetType = "list" | "csv";

export function CreateDatasetDialog({
  dataset,
  isOpen,
  onClose,
  onCreate,
  onUpdate,
}: CreateDatasetDialogProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [type, setType] = useState<DatasetType>("list");
  const [listOptions, setListOptions] = useState("");
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isGenerateOptionsDialogOpen, setIsGenerateOptionsDialogOpen] =
    useState(false);

  const [nameError, setNameError] = useState("");
  const [listOptionsError, setListOptionsError] = useState("");
  const [filesError, setFilesError] = useState("");

  const internalCloseInitiatedRef = useRef(false);

  useEffect(() => {
    resetForm();
    if (dataset) {
      setName(dataset.name);
      setDescription(dataset.description);
      setType(dataset.type);
      if (dataset.data) {
        setListOptions(dataset.data.join("\n"));
      }
    }
  }, [isOpen]);

  const resetForm = () => {
    setName("");
    setDescription("");
    setType("list");
    setListOptions("");
    setSelectedFiles([]);
    setNameError("");
    setListOptionsError("");
    setFilesError("");
  };

  const handleDialogShouldClose = () => {
    internalCloseInitiatedRef.current = true;
    resetForm();
    onClose();
  };

  const validate = (): boolean => {
    let isValid = true;
    if (!name.trim()) {
      setNameError("Name cannot be empty");
      isValid = false;
    } else {
      setNameError("");
    }

    if (type === "list" && !listOptions.trim()) {
      setListOptionsError("List options cannot be empty");
      isValid = false;
    } else {
      setListOptionsError("");
    }

    if (type === "csv" && dataset === undefined && selectedFiles.length === 0) {
      setFilesError("Please select at least one CSV file");
      isValid = false;
    } else {
      setFilesError("");
    }
    return isValid;
  };

  const handleSubmit = () => {
    if (!validate()) {
      return;
    }

    if (dataset) {
      if (type === "list") {
        onUpdate(dataset.id, {
          name,
          description,
          type,
          options: listOptions
            .split("\n")
            .map((opt) => opt.trim())
            .filter((opt) => opt),
        });
      } else if (type === "csv") {
        onUpdate(dataset.id, {
          name,
          description,
          type,
          files: selectedFiles,
        });
      }
    } else {
      if (type === "list") {
        onCreate({
          name,
          description,
          type,
          options: listOptions
            .split("\n")
            .map((opt) => opt.trim())
            .filter((opt) => opt),
        });
      } else if (type === "csv") {
        onCreate({
          name,
          description,
          type,
          files: selectedFiles,
        });
      }
    }
    handleDialogShouldClose();
  };

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files) {
      const filesArray = Array.from(event.target.files);
      const csvFiles = filesArray.filter(
        (file) => file.type === "text/csv" || file.name.endsWith(".csv"),
      );
      if (csvFiles.length !== filesArray.length) {
        setFilesError("Only CSV files are allowed.");
      } else {
        setFilesError("");
      }
      setSelectedFiles((prev) =>
        [...prev, ...csvFiles].filter(
          (f, i, self) =>
            self.findIndex((t) => t.name === f.name && t.size === f.size) === i,
        ),
      );
    }
  };

  const removeFile = (fileName: string) => {
    setSelectedFiles((prev) => prev.filter((file) => file.name !== fileName));
  };

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(newOpenState) => {
        if (!newOpenState) {
          if (internalCloseInitiatedRef.current) {
            internalCloseInitiatedRef.current = false;
          } else {
            resetForm();
            onClose();
          }
        }
      }}
    >
      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>
            {dataset ? "Update Dataset" : "Create New Dataset"}
          </DialogTitle>
          <DialogDescription>
            Enter the details for your dataset.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">
              Name
            </Label>
            <div className="col-span-3">
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className={nameError ? "border-red-500" : ""}
              />
              {nameError && (
                <p className="text-xs text-red-500 mt-1">{nameError}</p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="description" className="text-right">
              Description
            </Label>
            <Input
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="col-span-3"
            />
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="type-list" className="text-right">
              Type
            </Label>
            <RadioGroup
              id="type-radio-group"
              value={type}
              onValueChange={(value: string) => setType(value as DatasetType)}
              className="col-span-3 flex gap-4"
              disabled={dataset !== undefined}
            >
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="list" id="type-list" />
                <Label htmlFor="type-list">List</Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="csv" id="type-csv" />
                <Label htmlFor="type-csv">CSV</Label>
              </div>
            </RadioGroup>
          </div>

          {type === "list" && (
            <div className="grid grid-cols-4 items-start gap-4">
              <Label htmlFor="list-options" className="text-right pt-2">
                Options
              </Label>
              <div className="col-span-3 relative">
                <Textarea
                  id="list-options"
                  value={listOptions}
                  onChange={(e) => {
                    setListOptions(e.target.value);
                  }}
                  placeholder="Enter each option on a new line"
                  className={`min-h-[150px] pr-12 ${listOptionsError ? "border-red-500" : ""} hide-scrollbar`}
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
                {listOptionsError && (
                  <p className="text-xs text-red-500 mt-1">
                    {listOptionsError}
                  </p>
                )}
                <p className="text-xs text-muted-foreground mt-1">
                  Each line will be treated as a separate option.
                </p>
              </div>
            </div>
          )}

          {type === "csv" && (
            <div className="grid grid-cols-4 items-start gap-4">
              <Label htmlFor="csv-files" className="text-right pt-2">
                CSV Files
              </Label>
              <div className="col-span-3">
                <Input
                  id="csv-files"
                  type="file"
                  multiple
                  accept=".csv"
                  onChange={handleFileChange}
                  className="mb-2"
                />
                {filesError && (
                  <p className="text-xs text-red-500 mt-1 mb-2">{filesError}</p>
                )}
                {selectedFiles.length > 0 && (
                  <ScrollArea className="h-32 w-full rounded-md border p-2">
                    <div className="space-y-1">
                      {selectedFiles.map((file) => (
                        <div
                          key={file.name + file.size}
                          className="flex justify-between items-center text-sm pl-2 p-1 bg-muted/50 rounded"
                        >
                          <span className="truncate max-w-[80%]">
                            {file.name} ({(file.size / 1024).toFixed(2)} KB)
                          </span>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => removeFile(file.name)}
                            aria-label={`Remove ${file.name}`}
                          >
                            &times;
                          </Button>
                        </div>
                      ))}
                    </div>
                  </ScrollArea>
                )}
                <p className="text-xs text-muted-foreground mt-1">
                  {dataset
                    ? "Select one or more CSV files to REPLACE original data or leave it empty if you don't wan t to change."
                    : "Select one or more CSV files."}
                </p>
              </div>
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleDialogShouldClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={name == ""}>
            {dataset ? "Update" : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
      <GenerateOptionsDialog
        isOpen={isGenerateOptionsDialogOpen}
        onClose={() => setIsGenerateOptionsDialogOpen(false)}
        datasetName={name}
        currentOptions={listOptions.split("\n")}
        datasetDescription={description}
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
    </Dialog>
  );
}
