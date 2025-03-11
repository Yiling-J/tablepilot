import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogOverlay,
    DialogTitle,
} from "@/components/ui/dialog";

interface CellTextDialogProps {
  isOpen: boolean;
  setIsOpen: (v: boolean) => void;
  text: unknown;
}

export function CellTextDialog({
  text,
  isOpen,
  setIsOpen,
}: CellTextDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogOverlay />
      <DialogContent className="scrollbar-thin whitespace-pre-wrap overflow-auto max-h-[65vh] max-w-[50vw]">
        <DialogTitle></DialogTitle>
        <DialogDescription></DialogDescription>
        <div className="">{String(text)}</div>
      </DialogContent>
    </Dialog>
  );
}
