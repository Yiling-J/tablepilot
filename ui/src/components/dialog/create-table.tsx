import CreateTableForm from "@/components/create-table-form/create-table-form";
import {
    Dialog,
    DialogContent,
    DialogOverlay,
    DialogTitle,
} from "@/components/ui/dialog";

interface CreateTableDialogProps {
  isOpen: boolean;
  setIsOpen: (v: boolean) => void;
  close: () => void;
}

export function CreateTableDialog({
  isOpen,
  setIsOpen,
  close,
}: CreateTableDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogOverlay />
      <DialogContent
        className="min-w-[650px]"
        onInteractOutside={(e) => {
          e.preventDefault();
        }}
      >
        <DialogTitle>Create New Table</DialogTitle>
        <div className="mx-2 mt-2">
          <CreateTableForm close={close} />
        </div>
      </DialogContent>
    </Dialog>
  );
}
