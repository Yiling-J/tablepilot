import { Provider, ProviderType, ProviderTypeOptions } from "@/actions";
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";
import { useEffect, useState } from "react";

interface ProviderFormDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (provider: Provider) => void;
  initialData?: Provider | null;
}

export function ProviderFormDialog({
  isOpen,
  onOpenChange,
  onSubmit,
  initialData,
}: ProviderFormDialogProps) {
  const { toast } = useToast();
  const [selectedType, setSelectedType] = useState<ProviderType | undefined>(
    initialData?.type,
  );
  // Local state for input fields if not using react-hook-form
  const [name, setName] = useState(initialData?.name || "");
  const [apiKey, setApiKey] = useState(initialData?.key || "");
  const [baseUrl, setBaseUrl] = useState(initialData?.base_url || "");

  useEffect(() => {
    if (initialData) {
      setName(initialData.name);
      setSelectedType(initialData.type);
      setApiKey(initialData.key || "");
      setBaseUrl(initialData.base_url || "");
    } else {
      setName("");
      setSelectedType("OpenAI");
      setApiKey("");
      setBaseUrl("");
    }
  }, [initialData, isOpen]);

  const handleSubmitInternal = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const data: Provider = {
      id: initialData?.id || 0,
      models: initialData?.models || [],
      name: name,
      type: selectedType || "OpenAI-Compatible",
      key: apiKey,
      base_url: selectedType === "OpenAI-Compatible" ? baseUrl : "",
      editable: initialData?.editable ?? true,
      enabled: initialData?.enabled ?? true,
    };
    onSubmit(data);
    toast({
      title: initialData ? "Provider Updated" : "Provider Created",
      description: `${data.name} has been successfully ${initialData ? "updated" : "created"}.`,
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px] bg-card text-card-foreground">
        <DialogHeader>
          <DialogTitle>
            {initialData ? "Edit Provider" : "Create New Provider"}
          </DialogTitle>
          <DialogDescription>
            {initialData
              ? "Update the details of your AI provider."
              : "Add a new AI provider to manage models."}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmitInternal} className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">
              Name
            </Label>
            <Input
              id="name"
              name="name"
              className="col-span-3 bg-input border-border"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="type" className="text-right">
              Type
            </Label>
            {initialData ? (
              <div className="col-span-3 text-sm py-2 px-3 bg-muted/50 rounded-md border border-input">
                {selectedType}
              </div>
            ) : (
              <Select
                value={selectedType}
                onValueChange={(value: ProviderType) => {
                  setSelectedType(value);
                }}
              >
                <SelectTrigger className="col-span-3 bg-input border-border">
                  <SelectValue placeholder="Select provider type" />
                </SelectTrigger>
                <SelectContent>
                  {ProviderTypeOptions.map((type: ProviderType) => (
                    <SelectItem key={type} value={type}>
                      {type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="apiKey" className="text-right">
              API Key
            </Label>
            <Input
              id="apiKey"
              name="apiKey"
              type="password"
              className="col-span-3 bg-input border-border"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
            />
          </div>
          {selectedType === "OpenAI-compatible" && (
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="baseUrl" className="text-right">
                Base URL
              </Label>
              <Input
                id="baseUrl"
                name="baseUrl"
                className="col-span-3 bg-input border-border"
                value={baseUrl}
                onChange={(e) => setBaseUrl(e.target.value)}
              />
            </div>
          )}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" variant="default">
              {initialData ? "Save Changes" : "Create Provider"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
