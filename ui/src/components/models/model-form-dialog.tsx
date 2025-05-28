import { Model, ProviderType } from "@/actions";
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
import { Switch } from "@/components/ui/switch";
import { useToast } from "@/hooks/use-toast";

interface ModelFormDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (model: Model) => void;
  providerType: ProviderType;
  providerName: string; // To link model.client field
  initialData?: Model | null;
}

export function ModelFormDialog({
  isOpen,
  onOpenChange,
  onSubmit,
  providerType,
  providerName,
  initialData,
}: ModelFormDialogProps) {
  const { toast } = useToast();

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault(); // Prevent default form submission
    const formData = new FormData(event.currentTarget);
    const data: Model = {
      model: (formData.get("model") as string) || "",
      alias: (formData.get("alias") as string) || "",
      max_tokens: parseInt((formData.get("max_tokens") as string) || "0", 10),
      rpm: parseInt((formData.get("rpm") as string) || "0", 10),
      image: formData.get("imageSupport") === "on",
    };

    onSubmit(data);
    toast({
      title: initialData ? "Model Updated" : "Model Added",
      description: `${data.model} has been successfully ${initialData ? "updated" : "added"}.`,
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px] bg-card text-card-foreground">
        <DialogHeader>
          <DialogTitle>
            {initialData ? "Edit Model" : "Add New Model"}
          </DialogTitle>
          <DialogDescription>
            {initialData
              ? `Update details for model ${initialData.model}.`
              : `Add a new model to ${providerName} (${providerType}).`}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="model" className="text-right">
              Name
            </Label>
            <div className="col-span-3">
              <Input
                id="model"
                name="model"
                className="w-full bg-input border-border"
                defaultValue={initialData?.model || ""}
              />
            </div>
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="alias" className="text-right">
              Alias
            </Label>
            <Input
              id="alias"
              name="alias"
              className="col-span-3 bg-input border-border"
              placeholder="Optional display name"
              defaultValue={initialData?.alias || ""}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="max_tokens" className="text-right">
              Max Tokens
            </Label>
            <Input
              id="max_tokens"
              name="max_tokens"
              type="number"
              className="col-span-3 bg-input border-border"
              defaultValue={initialData?.max_tokens || 6000}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="rpm" className="text-right">
              RPM
            </Label>
            <Input
              id="rpm"
              name="rpm"
              type="number"
              className="col-span-3 bg-input border-border"
              defaultValue={initialData?.rpm || 10}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="imageSupport" className="text-right">
              Image Support
            </Label>
            <div className="col-span-3 flex items-center">
              <Switch
                id="imageSupport"
                name="imageSupport"
                defaultChecked={initialData?.image || false}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" variant="default">
              {initialData ? "Save Changes" : "Add Model"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
