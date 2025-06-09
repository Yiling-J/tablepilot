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
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>What is a Dataset?</DialogTitle>
          <DialogDescription className="space-y-2 text-sm">
            <p>
              A dataset is a structured collection of data—such as a CSV file of
              customers or a list of recipe cuisines.
            </p>

            <p>
              The primary purpose of a dataset is to provide context when
              generating rows in a table. For example, when generating recipes,
              you can use a dataset to populate the <strong>cuisine</strong>{" "}
              column with real values instead of letting AI generate them
              freely. This ensures each recipe has a cuisine from the list
              (randomly or sequentially selected), which then guides the AI to
              generate other columns accordingly. This results in more diverse
              and controlled outputs.
            </p>

            <p>
              Currently, two types of datasets are supported:{" "}
              <strong>CSV</strong> and <strong>List</strong>.
            </p>

            <ul className="list-disc pl-5">
              <li>
                <strong>CSV Dataset:</strong> Upload CSV files with a consistent
                header/schema. You can select one column as the fill value and
                others as context data.
              </li>
              <li>
                <strong>List Dataset:</strong> Manually input values or ask AI
                to help generate them.
              </li>
            </ul>

            <p>
              When defining table columns, set the fill mode to{" "}
              <strong>"Select from dataset"</strong> and choose your dataset.
              For example, if you're generating a sales plan using a customer
              dataset, you might use the <strong>Name</strong> column as the
              fill value, and <strong>Age</strong>, <strong>Job</strong>, and{" "}
              <strong>Salary</strong> as context fields to help the AI generate
              more relevant data.
            </p>
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button onClick={onClose}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
