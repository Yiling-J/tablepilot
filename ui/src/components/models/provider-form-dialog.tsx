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
import type { ProviderData, ProviderFormData, ProviderType } from "@/types.ts";
import { ProviderTypeOptions } from "@/types.ts";
import { useEffect, useState } from "react";
import { v4 as uuidv4 } from "uuid";

interface ProviderFormDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (provider: ProviderData) => void;
  initialData?: ProviderData | null;
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
  const [apiKey, setApiKey] = useState(initialData?.apiKey || "");
  const [baseUrl, setBaseUrl] = useState(initialData?.baseUrl || "");

  useEffect(() => {
    if (initialData) {
      setName(initialData.name);
      setSelectedType(initialData.type);
      setApiKey(initialData.apiKey || "");
      setBaseUrl(initialData.baseUrl || "");
    } else {
      setName("");
      setSelectedType("OpenAI");
      setApiKey("");
      setBaseUrl("");
    }
  }, [initialData, isOpen]); // Removed form from dependencies

  const handleSubmitInternal = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault(); // Prevent default form submission
    const data: ProviderFormData = {
      name: name,
      type: selectedType || "Generic", // Ensure selectedType has a default
      apiKey: apiKey,
      baseUrl: selectedType === "Generic" ? baseUrl : undefined,
    };
    const providerData: ProviderData = {
      id: initialData?.id || uuidv4(),
      ...data,
      models: initialData?.models || [],
      editable: initialData?.editable ?? true,
      enabled: initialData?.enabled ?? true,
    };
    onSubmit(providerData);
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
              : "Add a new AI provider to manage models."}{" "}
            For 'Generic' type, ensure the models are OpenAI-compatible.
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
          {selectedType === "Generic" && (
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
