import type React from "react";

import { Model, Provider } from "@/actions";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Switch } from "@/components/ui/switch";
import { PlusCircle, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";

interface ProviderDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  provider: Provider | null;
  onSave: (provider: Provider) => void;
}

export function ProviderDialog({
  open,
  onOpenChange,
  provider,
  onSave,
}: ProviderDialogProps) {
  const [formData, setFormData] = useState<Provider>({
    models: [] as Model[],
  } as Provider);

  useEffect(() => {
    if (provider) {
      setFormData(provider);
    } else {
      setFormData({
        id: 0,
        name: "",
        type: "openai",
        key: "",
        base_url: "https://api.openai.com/v1",
        models: [],
        editable: true,
      });
    }
  }, [provider, open]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleAddModel = () => {
    const newModel: Model = {
      model: "",
      alias: "",
      max_tokens: 6000,
      rpm: 15,
      image: false,
    };
    setFormData((prev) => ({
      ...prev,
      models: [...prev.models, newModel],
    }));
  };

  const handleModelChange = (index: number, field: keyof Model, value: any) => {
    setFormData((prev) => {
      const updatedModels = [...prev.models];
      updatedModels[index] = {
        ...updatedModels[index],
        [field]:
          field === "max_tokens" || field === "rpm" ? Number(value) : value,
      };

      return {
        ...prev,
        models: updatedModels,
      };
    });
  };

  const handleModelSwitchChange = (index: number, checked: boolean) => {
    setFormData((prev) => {
      const updatedModels = [...prev.models];
      updatedModels[index] = {
        ...updatedModels[index],
        image: checked,
      };
      return { ...prev, models: updatedModels };
    });
  };

  const handleRemoveModel = (index: number) => {
    setFormData((prev) => {
      const updatedModels = [...prev.models];
      updatedModels.splice(index, 1);
      return {
        ...prev,
        models: updatedModels,
      };
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl p-0">
        <form onSubmit={handleSubmit} className="flex flex-col max-h-[90vh]">
          <div className="p-6 border-b">
            <DialogHeader>
              <DialogTitle>
                {provider ? "Edit Provider" : "Add Provider"}
              </DialogTitle>
              <DialogDescription>
                {provider
                  ? "Update the provider details below."
                  : "Fill in the details to add a new provider."}
              </DialogDescription>
            </DialogHeader>

            <div className="grid gap-6 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Provider Name</Label>
                  <Input
                    id="name"
                    name="name"
                    value={formData.name}
                    onChange={handleChange}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="type">Provider Type</Label>
                  <Select
                    value={formData.type}
                    onValueChange={(value) =>
                      handleChange({
                        target: { name: "type", value },
                      } as React.ChangeEvent<HTMLInputElement>)
                    }
                  >
                    <SelectTrigger id="type" name="type">
                      <SelectValue placeholder="Select a provider" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="openai">
                        OpenAI Compatible API
                      </SelectItem>
                      <SelectItem value="gemini">
                        Gemini (Image Generation Only)
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="key">API Key</Label>
                  <Input
                    id="key"
                    name="key"
                    value={formData.key}
                    onChange={handleChange}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="base_url">Base URL</Label>
                  <Input
                    id="base_url"
                    name="base_url"
                    value={formData.base_url}
                    onChange={handleChange}
                    required
                  />
                </div>
              </div>
            </div>
          </div>

          <div className="overflow-y-auto p-6 scrollbar-thumb-rounded-full scrollbar-track-rounded-full scrollbar scrollbar-thumb-stone-500 scrollbar-track-background">
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <h3 className="text-lg font-medium">Models</h3>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleAddModel}
                >
                  <PlusCircle className="mr-2 h-4 w-4" />
                  Add Model
                </Button>
              </div>

              {formData.models.map((model, index) => (
                <Card key={index}>
                  <CardHeader className="pb-2">
                    <div className="flex justify-between items-center">
                      <CardTitle className="text-md">
                        Model {index + 1}
                      </CardTitle>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRemoveModel(index)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor={`model-${index}`}>Model Name</Label>
                        <Input
                          id={`model-${index}`}
                          value={model.model}
                          onChange={(e) =>
                            handleModelChange(index, "model", e.target.value)
                          }
                          required
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor={`alias-${index}`}>Alias</Label>
                        <Input
                          id={`alias-${index}`}
                          value={model.alias}
                          onChange={(e) =>
                            handleModelChange(index, "alias", e.target.value)
                          }
                          required
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor={`max-tokens-${index}`}>
                          Max Tokens
                        </Label>
                        <Input
                          id={`max-tokens-${index}`}
                          type="number"
                          value={model.max_tokens}
                          onChange={(e) =>
                            handleModelChange(
                              index,
                              "max_tokens",
                              e.target.value,
                            )
                          }
                          required
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor={`rpm-${index}`}>RPM</Label>
                        <Input
                          id={`rpm-${index}`}
                          type="number"
                          value={model.rpm}
                          onChange={(e) =>
                            handleModelChange(index, "rpm", e.target.value)
                          }
                          required
                        />
                      </div>
                      <div className="flex items-center space-x-2 pt-6">
                        <Switch
                          id={`image-${index}`}
                          checked={model.image}
                          onCheckedChange={(checked) =>
                            handleModelSwitchChange(index, checked)
                          }
                        />
                        <Label htmlFor={`image-${index}`}>Image Model</Label>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}

              {formData.models.length === 0 && (
                <div className="text-center py-4 text-muted-foreground">
                  No models added. Click "Add Model" to add one.
                </div>
              )}
            </div>
          </div>

          <div className="p-6 border-t mt-auto">
            <DialogFooter className="mt-0">
              <Button type="submit">
                {provider ? "Update Provider" : "Add Provider"}
              </Button>
            </DialogFooter>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
