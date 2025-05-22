import { Column } from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { useState } from "react";
import { AutofillInput } from "./autofill-input";

interface AutofillDialogProps {
  isOpen: boolean;
  setIsOpen: (v: boolean) => void;
  columns: Column[];
  onStart: (
    columns: string[],
    contextColumns: string[],
    prompt: string,
  ) => void;
}

export function AutofillDialog({
  isOpen,
  setIsOpen,
  columns,
  onStart,
}: AutofillDialogProps) {
  const [selectedColumns, setSelectedColumns] = useState<string[]>([]);
  const [selectedContextColumns, setSelectedContextColumns] = useState<
    string[]
  >([]);
  const [prompt, setPrompt] = useState("");

  const handleStart = () => {
    onStart(selectedColumns, selectedContextColumns, prompt);
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Autofill Configuration</DialogTitle>
        </DialogHeader>

        <AutofillInput
          allColumns={columns}
          columns={selectedColumns}
          contextColumns={selectedContextColumns}
          prompt={prompt}
          onColumnsChange={setSelectedColumns}
          onContextColumnsChange={setSelectedContextColumns}
          onPromptChange={setPrompt}
        />

        <DialogFooter className="mt-6">
          <DialogClose asChild>
            <Button variant="outline" type="button">
              Cancel
            </Button>
          </DialogClose>
          <Button type="button" onClick={handleStart}>
            Start
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
