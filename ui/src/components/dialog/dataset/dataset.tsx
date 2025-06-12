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
import { imageUrl } from "@/urls";
import {
    closestCenter,
    DndContext,
    DragEndEvent,
    KeyboardSensor,
    PointerSensor,
    useSensor,
    useSensors,
} from "@dnd-kit/core";
import {
    arrayMove,
    SortableContext,
    sortableKeyboardCoordinates,
    useSortable,
    verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, Wand2 } from "lucide-react";
import React, { useEffect, useRef, useState } from "react";
import { GenerateOptionsDialog } from "../generate-options-dialog";

interface FileItem {
  id: string;
  name: string;
  file?: File;
  thumbnail?: string;
}

export interface CreateDatasetDialogProps {
  dataset?: DatasetInfo;
  isOpen: boolean;
  onClose: () => void;
  onCreate: (payload: {
    name: string;
    description: string;
    type: "list" | "csv" | "image";
    data?: string[];
    files?: File[];
  }) => void;
  onUpdate: (
    id: string,
    payload: {
      name: string;
      description: string;
      type: "list" | "csv" | "image";
      data?: string[];
      files?: File[];
    },
  ) => void;
}

interface SortableFileItemProps {
  item: FileItem;
  datasetType: DatasetType;
  onRemove: (id: string) => void;
}

function SortableFileItem({
  item,
  datasetType,
  onRemove,
}: SortableFileItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: item.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.8 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      className={`flex justify-between items-center text-sm pl-1 pr-1 py-1 bg-muted/50 rounded mb-1 ${isDragging ? "shadow-lg border border-primary" : ""}`}
    >
      <div className="flex items-center flex-grow truncate">
        <button
          {...listeners}
          className="p-1 cursor-grab mr-2 hover:bg-slate-200 dark:hover:bg-slate-700 rounded"
          aria-label="Drag to reorder"
        >
          <GripVertical className="h-4 w-4 text-gray-500 dark:text-gray-400" />
        </button>
        {item.thumbnail && datasetType === "image" && (
          <img
            src={item.thumbnail}
            alt={`Image thumbnail for ${item.name}`}
            className="h-8 w-8 object-cover mr-2 rounded-sm"
          />
        )}
        <span className="truncate max-w-[calc(100%-4rem)]" title={item.name}>
          {item.name}
          {item.file
            ? `(${(item.file.size / 1024).toFixed(2)} KB)`
            : "(existing)"}
        </span>
      </div>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => onRemove(item.id)}
        aria-label={`Remove ${item.name}`}
        className="p-1 h-auto"
      >
        &times;
      </Button>
    </div>
  );
}

type DatasetType = "list" | "csv" | "image";

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

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (over && active.id !== over.id) {
      setFileItems((items) => {
        const oldIndex = items.findIndex((item) => item.id === active.id);
        const newIndex = items.findIndex((item) => item.id === over.id);
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
      switch (dataset.type) {
        case "csv":
        case "image":
          const initialFileItems = dataset.data.map((fileName, index) => ({
            id: `${dataset.id}-${fileName}-${index}`,
            name: fileName as string,
            file: undefined,
            thumbnail: imageUrl(`datasets/shared/${dataset.id}/${fileName}`),
          }));
          setFileItems(initialFileItems);
          break;
        case "list":
          setListOptions(dataset.data.join("\n"));
      }
    }
  }, [isOpen]);

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
    } else if (
      type === "image" &&
      dataset === undefined &&
      fileItems.length === 0
    ) {
      setFilesError("Please select at least one image file");
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
          data: listOptions
            .split("\n")
            .map((opt) => opt.trim())
            .map((opt) => opt.trim())
            .filter((opt) => opt),
        });
      } else if (type === "csv") {
        const orderedFileNames = fileItems.map((item) => item.name);
        const newFiles = fileItems
          .filter((item) => item.file)
          .map((item) => item.file as File);

        onUpdate(dataset.id, {
          name,
          description,
          type,
          data: orderedFileNames,
          files: newFiles.length > 0 ? newFiles : undefined,
        });
      } else if (type === "image") {
        const orderedFileNames = fileItems.map((item) => item.name);
        const newFiles = fileItems
          .filter((item) => item.file)
          .map((item) => item.file as File);
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
          data: listOptions
            .split("\n")
            .map((opt) => opt.trim())
            .filter((opt) => opt),
        });
      } else if (type === "csv") {
        const orderedFileNames = fileItems.map((item) => item.name);
        const newFiles = fileItems
          .filter((item) => item.file)
          .map((item) => item.file as File);

        onCreate({
          name,
          description,
          type,
          data: orderedFileNames,
          files: newFiles,
        });
      } else if (type === "image") {
        const orderedFileNames = fileItems.map((item) => item.name);
        const newFiles = fileItems
          .filter((item) => item.file)
          .map((item) => item.file as File);
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
      let processedFiles: File[] = [];
      let errorMessage = "";

      if (type === "csv") {
        processedFiles = filesArray.filter(
          (file) => file.type === "text/csv" || file.name.endsWith(".csv"),
        );
        if (processedFiles.length !== filesArray.length) {
          errorMessage = "Only CSV files are allowed.";
        }
      } else if (type === "image") {
        const imageMimeTypes = ["image/png", "image/jpeg", "image/gif"];
        const imageExtensions = [".png", ".jpg", ".jpeg", ".gif"];
        processedFiles = filesArray.filter(
          (file) =>
            imageMimeTypes.includes(file.type) ||
            imageExtensions.some((ext) =>
              file.name.toLowerCase().endsWith(ext),
            ),
        );
        if (processedFiles.length !== filesArray.length) {
          errorMessage = "Only image files (PNG, JPG, GIF) are allowed.";
        }
      } else {
        processedFiles = filesArray;
      }

      if (errorMessage) {
        setFilesError(errorMessage);
      } else {
        setFilesError("");
        const preliminaryFileItems: FileItem[] = [];
        const replaced = new Map();

        // Process existing files: replace if new one with same name is uploaded
        fileItems.forEach((f) => {
          const newFile = processedFiles.find((pf) => pf.name === f.name);
          if (newFile) {
            replaced.set(newFile.name, true);
            preliminaryFileItems.push({
              id: `${newFile.name}-${newFile.size}`,
              name: newFile.name,
              file: newFile,
              thumbnail: type === "image" ? undefined : f.thumbnail,
            });
          } else {
            preliminaryFileItems.push(f);
          }
        });

        // Add new files that weren't replacements
        processedFiles.forEach((pf) => {
          if (!replaced.get(pf.name)) {
            preliminaryFileItems.push({
              id: `${pf.name}-${pf.size}`,
              name: pf.name,
              file: pf,
              thumbnail: type === "image" ? undefined : undefined,
            });
          }
        });

        if (type === "image") {
          const thumbnailPromises = preliminaryFileItems.map((item) => {
            if (item.file && item.file.type.startsWith("image/")) {
              return new Promise<FileItem>((resolve) => {
                const reader = new FileReader();
                reader.onload = (e) => {
                  const img = new Image();
                  img.onload = () => {
                    const MAX_WIDTH = 50;
                    const MAX_HEIGHT = 50;
                    let width = img.width;
                    let height = img.height;

                    if (width > height) {
                      if (width > MAX_WIDTH) {
                        height *= MAX_WIDTH / width;
                        width = MAX_WIDTH;
                      }
                    } else {
                      if (height > MAX_HEIGHT) {
                        width *= MAX_HEIGHT / height;
                        height = MAX_HEIGHT;
                      }
                    }

                    const canvas = document.createElement("canvas");
                    canvas.width = width;
                    canvas.height = height;
                    const ctx = canvas.getContext("2d");
                    if (ctx) {
                      ctx.drawImage(img, 0, 0, width, height);
                      item.thumbnail = canvas.toDataURL(
                        item.file?.type || "image/png",
                      );
                    } else {
                      item.thumbnail = undefined;
                    }
                    resolve(item);
                  };
                  img.onerror = () => {
                    item.thumbnail = undefined;
                    resolve(item);
                  };
                  img.src = e.target?.result as string;
                };
                reader.onerror = () => {
                  item.thumbnail = undefined;
                  resolve(item);
                };
                if (item.file) {
                  reader.readAsDataURL(item.file);
                }
              });
            }
            return Promise.resolve(item);
          });

          Promise.all(thumbnailPromises).then((updatedFileItems) => {
            setFileItems(updatedFileItems);
            if (filesError === "Please select at least one image file") {
              setFilesError("");
            }
          });
        } else {
          // For CSV or other types, set items directly
          setFileItems(preliminaryFileItems);
          if (
            type === "csv" &&
            filesError === "Please select at least one CSV file"
          ) {
            setFilesError("");
          }
        }
      }
    }
  };

  const removeFile = (idToRemove: string) => {
    // Changed parameter to id
    setFileItems((prevItems) =>
      prevItems.filter((item) => item.id !== idToRemove),
    );
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
      <DialogContent className="sm:max-w-[550px]">
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
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="image" id="type-image" />
                <Label htmlFor="type-image">Image</Label>
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
                      items={fileItems.map((item) => item.id)}
                      strategy={verticalListSortingStrategy}
                    >
                      <ScrollArea className="max-h-[280px] w-full rounded-md border p-2 overflow-y-auto">
                        <div className="space-y-1">
                          {fileItems.map((item) => (
                            <SortableFileItem
                              datasetType={type}
                              key={item.id}
                              item={item}
                              onRemove={removeFile}
                            />
                          ))}
                        </div>
                      </ScrollArea>
                    </SortableContext>
                  </DndContext>
                )}
                <p className="text-xs text-muted-foreground mt-1">
                  {dataset && type === "csv"
                    ? "Add new CSV files to replace existing ones. Leave empty to keep current files. You can reorder files by dragging."
                    : "Select one or more CSV files."}
                </p>
              </div>
            </div>
          )}

          {type === "image" && (
            <div className="grid grid-cols-4 items-start gap-4">
              <Label htmlFor="image-files" className="text-right pt-2">
                Image Files
              </Label>
              <div className="col-span-3">
                <Input
                  id="image-files"
                  type="file"
                  multiple
                  accept="image/png, image/jpeg, image/gif, .png, .jpg, .jpeg, .gif"
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
                      items={fileItems.map((item) => item.id)}
                      strategy={verticalListSortingStrategy}
                    >
                      <ScrollArea className="max-h-[280px] w-full rounded-md border p-2 overflow-y-auto">
                        <div className="space-y-1">
                          {fileItems.map((item) => (
                            <SortableFileItem
                              datasetType={type}
                              key={item.id}
                              item={item}
                              onRemove={removeFile}
                            />
                          ))}
                        </div>
                      </ScrollArea>
                    </SortableContext>
                  </DndContext>
                )}
                <p className="text-xs text-muted-foreground mt-1">
                  {dataset && type === "image"
                    ? "Add new image files to replace existing ones. Leave empty to keep current files. You can reorder files by dragging."
                    : "Select one or more image files (PNG, JPG, GIF)."}
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
