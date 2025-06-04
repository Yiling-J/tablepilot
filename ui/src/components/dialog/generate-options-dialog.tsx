import { useEffect, useState } from "react";
import toast from "react-hot-toast";

import { ModelSelector } from "@/components/model-selector";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { generateOptions } from "../../actions";

export interface GenerateOptionsDialogProps {
  isOpen: boolean;
  currentOptions: string[];
  onClose: () => void;
  onGenerationComplete: (generatedOptions: string[]) => void;
  datasetName?: string;
  datasetDescription?: string;
}

export function GenerateOptionsDialog({
  isOpen,
  currentOptions,
  onClose,
  onGenerationComplete,
  datasetName,
  datasetDescription,
}: GenerateOptionsDialogProps) {
  const [selectedModel, setSelectedModel] = useState<string>("");
  const [prompt, setPrompt] = useState<string>("");
  const [isLoading, setIsLoading] = useState<boolean>(false);

  useEffect(() => {
    if (isOpen) {
      setSelectedModel("");

      let initialPrompt = "Based on a dataset";
      if (datasetName) {
        initialPrompt += ` named '${datasetName}'`;
      }
      if (datasetDescription) {
        initialPrompt += ` (Description: '${datasetDescription}')`;
      }
      initialPrompt +=
        ", generate a list of options for this dataset. The options should be distinct.";
      setPrompt(initialPrompt);
    }
  }, [isOpen]);

  const handleGenerateClick = async () => {
    if (!selectedModel) {
      toast.error("Please select a model.");
      return;
    }
    if (!prompt.trim()) {
      toast.error("Prompt cannot be empty.");
      return;
    }

    setIsLoading(true);
    try {
      const options = await generateOptions({
        model: selectedModel,
        prompt,
        options: currentOptions,
      });
      onGenerationComplete(options);
      toast.success("Options generated successfully!");
      onClose(); // Close the dialog
    } catch (error) {
      console.error("Failed to generate options:", error);
      toast.error(
        error instanceof Error
          ? error.message
          : "An unknown error occurred during generation.",
      );
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) {
    return null;
  }

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) {
          onClose();
        }
      }}
    >
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Generate AI Options</DialogTitle>
          <DialogDescription>
            Select a model and refine the prompt to generate options based on
            your dataset.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="flex items-center">
            <Label htmlFor="model-selector" className="mr-2">
              Model
            </Label>
            <ModelSelector
              selectModel={setSelectedModel}
              hasImageColumn={false}
              generating={isLoading}
              selectImageModel={() => {}}
            />
          </div>
          <div className="grid w-full gap-1.5">
            <Label htmlFor="prompt">Prompt</Label>
            <Textarea
              id="prompt"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="Enter your prompt here..."
              className="min-h-[150px]"
              disabled={isLoading}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={isLoading}>
            Cancel
          </Button>
          <Button
            onClick={handleGenerateClick}
            disabled={isLoading || !selectedModel || !prompt.trim()}
          >
            {isLoading ? "Generating..." : "Generate"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
