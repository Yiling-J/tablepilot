
'use client';

import { useEffect, useState } from 'react';
import { useForm, type SubmitHandler } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { v4 as uuidv4 } from 'uuid';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { ModelData, ModelFormData, ProviderType } from '@/types';
import { predefinedModels } from '@/lib/mock-data';
import { useToast } from '@/hooks/use-toast';

const modelFormSchema = z.object({
  model: z.string().min(1, 'Model name is required.'),
  alias: z.string().optional(),
  max_tokens: z.coerce.number().int().min(0, 'Max tokens must be a non-negative integer.'),
  rpm: z.coerce.number().int().min(0, 'RPM must be a non-negative integer.'),
  isDefault: z.boolean().optional(),
  imageSupport: z.boolean(),
});

interface ModelFormDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (model: ModelData) => void;
  providerType: ProviderType;
  providerName: string; // To link model.client field
  initialData?: ModelData | null;
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
  const isKnownProviderWithPredefinedModels = providerType !== 'Generic' && predefinedModels[providerType]?.length > 0;
  const [useCustomModelName, setUseCustomModelName] = useState(false);

  const form = useForm<ModelFormData>({
    resolver: zodResolver(modelFormSchema),
    // Default values are set in the useEffect hook based on initialData and useCustomModelName
  });

  useEffect(() => {
    if (!isOpen) return; // Only run when dialog is open

    let defaultModelValue = '';
    let initialCustomNameState = false;

    if (initialData) {
      const isPredefined = isKnownProviderWithPredefinedModels && predefinedModels[providerType].includes(initialData.model);
      initialCustomNameState = !isPredefined;
      defaultModelValue = initialData.model;
      form.reset({
        ...initialData,
        alias: initialData.alias || '',
      });
    } else {
      // For new models, default to selecting from list if available
      initialCustomNameState = false; 
      if (isKnownProviderWithPredefinedModels && predefinedModels[providerType]?.length > 0) {
        defaultModelValue = predefinedModels[providerType][0];
      }
      form.reset({
        model: defaultModelValue,
        alias: '',
        max_tokens: 6000, // Updated default
        rpm: 10,         // Updated default
        isDefault: false,
        imageSupport: false,
      });
    }
    
    setUseCustomModelName(initialCustomNameState);
    // Ensure model field is correctly set after state update for custom name
    form.setValue('model', defaultModelValue, { shouldValidate: initialData ? true : false });


  }, [initialData, isOpen, providerType, isKnownProviderWithPredefinedModels, form]);


  const handleToggleCustomModelName = (checked: boolean) => {
    setUseCustomModelName(checked);
    if (!checked && isKnownProviderWithPredefinedModels && predefinedModels[providerType]?.length > 0) {
      // Switching from manual to select: set to first predefined model
      form.setValue('model', predefinedModels[providerType][0], { shouldValidate: true });
    }
    // If switching to manual, current value (either selected or previously manual) remains
  };

  const handleSubmit: SubmitHandler<ModelFormData> = (data) => {
    const modelData: ModelData = {
      id: initialData?.id || uuidv4(),
      client: providerName,
      ...data,
      alias: data.alias || data.model, // Default alias to model name if empty
    };
    onSubmit(modelData);
    toast({ title: initialData ? "Model Updated" : "Model Added", description: `${data.model} has been successfully ${initialData ? 'updated' : 'added'}.` });
    onOpenChange(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px] bg-card text-card-foreground">
        <DialogHeader>
          <DialogTitle>{initialData ? 'Edit Model' : 'Add New Model'}</DialogTitle>
          <DialogDescription>
            {initialData ? `Update details for model ${initialData.model}.` : `Add a new model to ${providerName} (${providerType}).`}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(handleSubmit)} className="grid gap-4 py-4">
          
          {isKnownProviderWithPredefinedModels && (
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="useCustomModelNameSwitch" className="text-right col-span-3">
                Enter model name manually
              </Label>
              <div className="col-span-1 flex items-center justify-start">
                <Switch
                  id="useCustomModelNameSwitch"
                  checked={useCustomModelName}
                  onCheckedChange={handleToggleCustomModelName}
                />
              </div>
            </div>
          )}

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="model" className="text-right">Model Name</Label>
            <div className="col-span-3">
              {(isKnownProviderWithPredefinedModels && !useCustomModelName) ? (
                <Select
                  value={form.watch('model')}
                  onValueChange={(value) => form.setValue('model', value, { shouldValidate: true })}
                >
                  <SelectTrigger className="w-full bg-input border-border">
                    <SelectValue placeholder="Select a model" />
                  </SelectTrigger>
                  <SelectContent>
                    {predefinedModels[providerType].map(modelName => (
                      <SelectItem key={modelName} value={modelName}>{modelName}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : (
                <Input id="model" {...form.register('model')} className="w-full bg-input border-border" />
              )}
            </div>
            {form.formState.errors.model && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.model.message}</p>}
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="alias" className="text-right">Alias</Label>
            <Input id="alias" {...form.register('alias')} className="col-span-3 bg-input border-border" placeholder="Optional display name" />
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="max_tokens" className="text-right">Max Tokens</Label>
            <Input id="max_tokens" type="number" {...form.register('max_tokens')} className="col-span-3 bg-input border-border" />
            {form.formState.errors.max_tokens && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.max_tokens.message}</p>}
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="rpm" className="text-right">RPM</Label>
            <Input id="rpm" type="number" {...form.register('rpm')} className="col-span-3 bg-input border-border" />
            {form.formState.errors.rpm && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.rpm.message}</p>}
          </div>
          
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="imageSupport" className="text-right">Image Support</Label>
            <div className="col-span-3 flex items-center">
              <Switch
                id="imageSupport"
                checked={form.watch('imageSupport')}
                onCheckedChange={(checked) => form.setValue('imageSupport', checked)}
              />
            </div>
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="isDefault" className="text-right">Set as Default</Label>
             <div className="col-span-3 flex items-center">
              <Switch
                id="isDefault"
                checked={form.watch('isDefault')}
                onCheckedChange={(checked) => form.setValue('isDefault', checked)}
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button type="submit" variant="primary">{initialData ? 'Save Changes' : 'Add Model'}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
