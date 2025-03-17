import { TableCreateRequest } from "@/actions";
import CreateTableForm from "@/components/create-table-form/create-table-form";
import {
    Dialog,
    DialogContent,
    DialogOverlay,
    DialogTitle,
} from "@/components/ui/dialog";
import { JSONObject } from "@/json";

interface CreateTableDialogProps {
  isOpen: boolean;
  setIsOpen: (v: boolean) => void;
  close: () => void;
  form?: TableCreateRequest;
  rows?: JSONObject[];
}

export function CreateTableDialog({
  isOpen,
  setIsOpen,
  close,
  form,
  rows,
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
        <div className="mx-2 mt-2 scrollbar-thumb-rounded-full scrollbar-track-rounded-full scrollbar scrollbar-thumb-stone-500 scrollbar-track-background">
          <CreateTableForm close={close} form={form} rows={rows} />
        </div>
      </DialogContent>
    </Dialog>
  );
}
