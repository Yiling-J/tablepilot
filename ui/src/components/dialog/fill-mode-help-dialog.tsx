import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";

interface FillModeHelpDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

export function FillModeHelpDialog({ isOpen, onClose }: FillModeHelpDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Fill Mode Definitions</DialogTitle>
          <DialogDescription>
            Understand the different methods for populating column data.
          </DialogDescription>
        </DialogHeader>
        <div className="py-4 space-y-3 text-sm max-h-[60vh] overflow-y-auto scrollbar-thin">
          <div>
            <h4 className="font-semibold mb-1">AI Generated</h4>
            <p className="text-muted-foreground">
              The column name and description will be used to guide the AI in generating data for this column.
              The 'Context Length' parameter can be used to specify how many previous rows (if any) the AI should consider when generating content for the current row, allowing for contextual generation.
            </p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Select from Table</h4>
            <p className="text-muted-foreground">
              Populate this column by selecting values from an existing column in another table.
              The 'Linked Column' you choose from the source table will provide the values for this column.
              'Linked Context Columns' are additional columns from the source table that can be used as contextual information by AI when generating other columns in the current row. For example, if you're generating a recipe (AI column), and you link a 'Cuisine Type' column from another table, that cuisine type can guide the AI's recipe generation.
              Options for random selection, selection with replacement, and repeating selections are available.
            </p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Select from Dataset</h4>
            <p className="text-muted-foreground">
              Choose values from an existing dataset. This can be a CSV dataset or a List dataset.
              For CSV datasets: The 'Linked Column' you select from the dataset will provide the values for this column. 'Linked Context Columns' from the dataset can be used as context for AI generating other parts of the row (similar to 'Select from Table').
              For List datasets: This mode works like 'Select from Options', where values are picked from a predefined list. Using a List dataset allows you to share and reuse the same set of options across multiple tables.
              Both types offer options for random selection, selection with replacement, and repetition.
            </p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Select from Options</h4>
            <p className="text-muted-foreground">
              Define a specific list of possible values for this column. The system will then pick from these options to populate the cells.
              You can enter each option on a new line.
              This mode also provides settings for random selection, selection with replacement, and how many times each selection can be repeated.
              You can also use AI to help generate these options based on the column name and description.
            </p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
