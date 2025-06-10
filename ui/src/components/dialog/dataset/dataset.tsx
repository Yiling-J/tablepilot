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
import { Wand2, GripVertical } from "lucide-react"; // Added GripVertical for drag handle
import React, { useEffect, useRef, useState } from "react";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GenerateOptionsDialog } from "../generate-options-dialog";

interface FileItem {
  id: string;
  name: string;
  file?: File;
}

export interface CreateDatasetDialogProps {
  // Added export
  dataset?: DatasetInfo;
  isOpen: boolean;
  onClose: () => void;
  onCreate: (payload: { // Renamed 'data' to 'payload' for clarity
    name: string;
    description: string;
    type: "list" | "csv";
    options?: string[]; // For list type
    data?: string[];    // For csv type (ordered file names)
    files?: File[];     // For csv type (actual File objects for new files)
  }) => void;
  onUpdate: (
    id: string,
    payload: { // Renamed 'data' to 'payload' for clarity
      name: string;
      description: string;
      type: "list" | "csv";
      options?: string[]; // For list type
      data?: string[];    // For csv type (ordered file names)
      files?: File[];     // For csv type (actual File objects for new/changed files)
    },
  ) => void;
}

// 2. Create SortableFileItem component
interface SortableFileItemProps {
  item: FileItem;
  onRemove: (id: string) => void;
}

function SortableFileItem({ item, onRemove }: SortableFileItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging, // Can be used for styling if needed
  } = useSortable({ id: item.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.8 : 1, // Example: reduce opacity when dragging
    // zIndex: isDragging ? 100 : 'auto', // Example: ensure dragging item is on top
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      className={`flex justify-between items-center text-sm pl-1 pr-1 py-1 bg-muted/50 rounded mb-1 ${isDragging ? 'shadow-lg border border-primary' : ''}`}
    >
      <div className="flex items-center flex-grow truncate">
        <button {...listeners} className="p-1 cursor-grab mr-2 hover:bg-slate-200 dark:hover:bg-slate-700 rounded" aria-label="Drag to reorder">
          <GripVertical className="h-4 w-4 text-gray-500 dark:text-gray-400" />
        </button>
        <span className="truncate max-w-[calc(100%-4rem)]" title={item.name}> {/* Adjust max-width if needed */}
          {item.name}{" "}
          {item.file ? `(${(item.file.size / 1024).toFixed(2)} KB)` : "(existing)"}
        </span>
      </div>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => onRemove(item.id)}
        aria-label={`Remove ${item.name}`}
        className="p-1 h-auto" // Adjusted padding and height for a smaller button
      >
        &times;
      </Button>
    </div>
  );
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
  const [fileItems, setFileItems] = useState<FileItem[]>([]);
  const [isGenerateOptionsDialogOpen, setIsGenerateOptionsDialogOpen] =
    useState(false);

  const [nameError, setNameError] = useState("");
  const [listOptionsError, setListOptionsError] = useState("");
  const [filesError, setFilesError] = useState("");

  const internalCloseInitiatedRef = useRef(false);

  // DnD Sensors
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  // DnD Drag End Handler
  function handleDragEnd(event: DragEndEvent) {
    const {active, over} = event;
    if (over && active.id !== over.id) {
      setFileItems((items) => {
        const oldIndex = items.findIndex(item => item.id === active.id);
        const newIndex = items.findIndex(item => item.id === over.id);
        return arrayMove(items, oldIndex, newIndex);
      });
    }
  }

  useEffect(() => {
    resetForm();
    if (dataset) {
      setName(dataset.name);
      setDescription(dataset.description);
      setType(dataset.type);
      if (dataset.type === "csv" && dataset.data && Array.isArray(dataset.data)) {
        const initialFileItems = dataset.data.map((fileName, index) => ({
          id: `${dataset.id}-${fileName}-${index}`,
          name: fileName as string,
          file: undefined,
        }));
        setFileItems(initialFileItems);
      } else if (dataset.type === "list" && dataset.data && Array.isArray(dataset.data)) {
        setListOptions(dataset.data.join("\n"));
      } else {
        // Clear if not matching conditions (e.g. new dataset, or existing non-csv/list with no specific data handling here)
        setFileItems([]);
        if (dataset.type === "list") { // only clear listOptions if it's a list type
          setListOptions("");
        }
      }
    } else {
      // This 'else' corresponds to '!dataset', meaning we are creating a new entity.
      // resetForm() is already called at the beginning of useEffect,
      // so fileItems and listOptions will be in their initial empty state.
    }
  }, [isOpen, dataset]); // Ensure dataset is in the dependency array

  const resetForm = () => {
    setName("");
    setDescription("");
    setType("list");
    setListOptions("");
    setFileItems([]);
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

    if (type === "csv" && dataset === undefined && fileItems.length === 0) {
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
            .map((opt) => opt.trim())
            .filter((opt) => opt),
        });
      } else if (type === "csv") {
        const orderedFileNames = fileItems.map(item => item.name);
        const newFiles = fileItems.filter(item => item.file).map(item => item.file as File);

        onUpdate(dataset.id, {
          name,
          description,
          type,
          data: orderedFileNames,
          files: newFiles.length > 0 ? newFiles : undefined,
        });
      }
    } else {
      // Creating a new dataset
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
        const orderedFileNames = fileItems.map(item => item.name);
        // For creation, all files in fileItems should ideally be new files.
        // If a fileItem lacks a .file, it implies an issue upstream or an edge case not handled,
        // as new items should always have a .file.
        const newFiles = fileItems.filter(item => item.file).map(item => item.file as File);

        // It's crucial that for creation, if fileItems has entries, newFiles should also have entries.
        // The existing validation `if (type === "csv" && dataset === undefined && fileItems.length === 0)`
        // should prevent submission if no files are selected at all.
        onCreate({
          name,
          description,
          type,
          data: orderedFileNames,
          files: newFiles,
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
        // Do not proceed to add non-CSV files
      } else {
        setFilesError(""); // Clear error if all files are CSVs or no files selected initially

        const newFileItems: FileItem[] = csvFiles
          .map((file) => ({
            id: `${file.name}-${file.size}-${Date.now()}`, // Unique ID
            name: file.name,
            file: file,
          }))
          .filter( // Prevent adding duplicates based on name and file size
            (newItem) =>
              !fileItems.some(
                (existingItem) =>
                  existingItem.name === newItem.name &&
                  existingItem.file?.size === newItem.file?.size
              )
          );

        if (newFileItems.length > 0) {
          setFileItems((prevItems) => [...prevItems, ...newFileItems]);
          // If the specific error "Please select at least one CSV file" was shown, clear it.
          if (filesError === "Please select at least one CSV file") {
            setFilesError("");
          }
        }
      }
    }
  };

  const removeFile = (idToRemove: string) => { // Changed parameter to id
    setFileItems((prevItems) => prevItems.filter((item) => item.id !== idToRemove));
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
                {fileItems.length > 0 && (
                  <DndContext
                    sensors={sensors}
                    collisionDetection={closestCenter}
                    onDragEnd={handleDragEnd}
                  >
                    <SortableContext
                      items={fileItems.map(item => item.id)}
                      strategy={verticalListSortingStrategy}
                    >
                      <ScrollArea className="h-32 w-full rounded-md border p-2">
                        <div className="space-y-1"> {/* This div helps with spacing between SortableFileItem */}
                          {fileItems.map((item) => (
                            <SortableFileItem key={item.id} item={item} onRemove={removeFile} />
                          ))}
                        </div>
                      </ScrollArea>
                    </SortableContext>
                  </DndContext>
                )}
                <p className="text-xs text-muted-foreground mt-1">
                  {dataset && type === 'csv'
                    ? "Add new CSV files to replace existing ones. Leave empty to keep current files. You can reorder files by dragging."
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
