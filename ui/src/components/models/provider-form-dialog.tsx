'use client';

import { useState, useEffect } from 'react';
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { ProviderData, ProviderFormData, ProviderType } from '@/types';
import { ProviderTypeOptions } from '@/types';
import { useToast } from '@/hooks/use-toast';

const providerFormSchema = z.object({
  name: z.string().min(1, 'Provider name is required.'),
  type: z.enum(ProviderTypeOptions),
  apiKey: z.string().min(1, 'API Key is required.'),
  baseUrl: z.string().optional(),
}).refine(data => data.type !== 'Generic' || (data.type === 'Generic' && data.baseUrl && data.baseUrl.length > 0), {
  message: 'Base URL is required for Generic provider type.',
  path: ['baseUrl'],
});

interface ProviderFormDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (provider: ProviderData) => void;
  initialData?: ProviderData | null;
}

export function ProviderFormDialog({ isOpen, onOpenChange, onSubmit, initialData }: ProviderFormDialogProps) {
  const { toast } = useToast();
  const [selectedType, setSelectedType] = useState<ProviderType | undefined>(initialData?.type);

  const form = useForm<ProviderFormData>({
    resolver: zodResolver(providerFormSchema),
    defaultValues: initialData ? {
      name: initialData.name,
      type: initialData.type,
      apiKey: initialData.apiKey,
      baseUrl: initialData.baseUrl || '',
    } : {
      name: '',
      type: 'OpenAI',
      apiKey: '',
      baseUrl: '',
    },
  });

  useEffect(() => {
    if (initialData) {
      form.reset({
        name: initialData.name,
        type: initialData.type,
        apiKey: initialData.apiKey,
        baseUrl: initialData.baseUrl || '',
      });
      setSelectedType(initialData.type);
    } else {
      form.reset({
        name: '',
        type: 'OpenAI',
        apiKey: '',
        baseUrl: '',
      });
      setSelectedType('OpenAI');
    }
  }, [initialData, form, isOpen]);
  
  useEffect(() => {
    const subscription = form.watch((value, { name }) => {
      if (name === 'type') {
        setSelectedType(value.type);
      }
    });
    return () => subscription.unsubscribe();
  }, [form]);

  const handleSubmitInternal: SubmitHandler<ProviderFormData> = (data) => {
    const providerData: ProviderData = {
      id: initialData?.id || uuidv4(),
      ...data,
      models: initialData?.models || [],
      editable: initialData?.editable ?? true,
      enabled: initialData?.enabled ?? true, 
    };
    onSubmit(providerData);
    toast({ title: initialData ? "Provider Updated" : "Provider Created", description: `${data.name} has been successfully ${initialData ? 'updated' : 'created'}.` });
    onOpenChange(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px] bg-card text-card-foreground">
        <DialogHeader>
          <DialogTitle>{initialData ? 'Edit Provider' : 'Create New Provider'}</DialogTitle>
          <DialogDescription>
            {initialData ? 'Update the details of your AI provider.' : 'Add a new AI provider to manage models.'}
            {' '}For 'Generic' type, ensure the models are OpenAI-compatible.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(handleSubmitInternal)} className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">Name</Label>
            <Input id="name" {...form.register('name')} className="col-span-3 bg-input border-border" />
            {form.formState.errors.name && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.name.message}</p>}
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="type" className="text-right">Type</Label>
            {initialData ? (
              <div className="col-span-3 text-sm py-2 px-3 bg-muted/50 rounded-md border border-input">
                {form.getValues('type')}
              </div>
            ) : (
              <Select
                value={form.watch('type')}
                onValueChange={(value) => {
                  form.setValue('type', value as ProviderType, { shouldValidate: true });
                  setSelectedType(value as ProviderType);
                }}
              >
                <SelectTrigger className="col-span-3 bg-input border-border">
                  <SelectValue placeholder="Select provider type" />
                </SelectTrigger>
                <SelectContent>
                  {ProviderTypeOptions.map(type => (
                    <SelectItem key={type} value={type}>{type}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
            {form.formState.errors.type && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.type.message}</p>}
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="apiKey" className="text-right">API Key</Label>
            <Input id="apiKey" type="password" {...form.register('apiKey')} className="col-span-3 bg-input border-border" />
            {form.formState.errors.apiKey && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.apiKey.message}</p>}
          </div>
          {selectedType === 'Generic' && (
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="baseUrl" className="text-right">Base URL</Label>
              <Input id="baseUrl" {...form.register('baseUrl')} className="col-span-3 bg-input border-border" />
              {form.formState.errors.baseUrl && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.baseUrl.message}</p>}
            </div>
          )}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button type="submit" variant="primary">{initialData ? 'Save Changes' : 'Create Provider'}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
