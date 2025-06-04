import { useState, useEffect } from "react";
import toast from "react-hot-toast";

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
import { ModelSelector } from "@/components/model-selector";
import { generateOptions } from "../../actions";

export interface GenerateOptionsDialogProps { // Exported interface
  isOpen: boolean;
  onClose: () => void;
  onGenerationComplete: (generatedOptions: string[]) => void; // Modified prop
  datasetName?: string;
  datasetDescription?: string;
}

export function GenerateOptionsDialog({
  isOpen,
  onClose,
  onGenerationComplete, // Renamed prop
  datasetName,
  datasetDescription,
}: GenerateOptionsDialogProps) {
  const [selectedModel, setSelectedModel] = useState<string>("");
  const [prompt, setPrompt] = useState<string>("");
  const [isLoading, setIsLoading] = useState<boolean>(false);

  useEffect(() => {
    if (isOpen) {
      // Reset state when dialog opens, except for prompt which gets prefilled
      setSelectedModel(""); // Reset selected model
      // setIsLoading(false); // isLoading should reset if dialog was closed while loading

      let initialPrompt = "Based on a dataset";
      if (datasetName) {
        initialPrompt += ` named '${datasetName}'`;
      }
      if (datasetDescription) {
        initialPrompt += ` (Description: '${datasetDescription}')`;
      }
      initialPrompt += ", generate a list of relevant options. The options should be distinct, actionable, and suitable for a list-based dataset. Provide each option on a new line.";
      setPrompt(initialPrompt);
    }
  }, [isOpen, datasetName, datasetDescription]);

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
      const options = await generateOptions({ model: selectedModel, prompt });
      onGenerationComplete(options); // Pass the generated options back
      toast.success("Options generated successfully!");
      onClose(); // Close the dialog
    } catch (error) {
      console.error("Failed to generate options:", error);
      toast.error(error instanceof Error ? error.message : "An unknown error occurred during generation.");
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) {
    return null;
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => {
      if (!open) {
        onClose();
      }
    }}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Generate AI Options</DialogTitle>
          <DialogDescription>
            Select a model and refine the prompt to generate options based on your dataset.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="model-selector" className="text-right">
              Model
            </Label>
            <ModelSelector // className="col-span-3" removed
              selectModel={setSelectedModel}
              hasImageColumn={false}
              generating={isLoading} // Added generating prop
              selectImageModel={() => {}} // Added selectImageModel prop as no-op
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
          <Button onClick={handleGenerateClick} disabled={isLoading || !selectedModel || !prompt.trim()}>
            {isLoading ? "Generating..." : "Generate"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
