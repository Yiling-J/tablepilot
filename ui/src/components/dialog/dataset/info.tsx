import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import React from "react";

interface DatasetInfoDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

export const DatasetInfoDialog: React.FC<DatasetInfoDialogProps> = ({
  isOpen,
  onClose,
}) => {
  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>What is a Dataset?</DialogTitle>
          <DialogDescription>
            A dataset is a structured collection of data. It can be thought of
            as a table, where each row represents an item and each column
            represents a property of that item.
          </DialogDescription>
        </DialogHeader>
        <div className="py-4">
          <p>
            Datasets are used to train machine learning models, generate insights,
            or simply store and organize information. In Tablepilot, datasets
            can be created from CSV files or by defining a list of options.
          </p>
        </div>
        <DialogFooter>
          <Button onClick={onClose}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
