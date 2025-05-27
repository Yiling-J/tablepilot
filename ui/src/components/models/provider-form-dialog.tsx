'use client';

import { useState, useEffect } from 'react'; // Removed: useForm, SubmitHandler, zodResolver, z
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
import type { ProviderData, ProviderFormData, ProviderType } from '@/types.ts';
import { ProviderTypeOptions } from '@/types.ts';
import { useToast } from '@/hooks/use-toast';

// const providerFormSchema = z.object({ // Schema removed
//   name: z.string().min(1, 'Provider name is required.'),
//   type: z.enum(ProviderTypeOptions),
//   apiKey: z.string().min(1, 'API Key is required.'),
//   baseUrl: z.string().optional(),
// }).refine(data => data.type !== 'Generic' || (data.type === 'Generic' && data.baseUrl && data.baseUrl.length > 0), {
//   message: 'Base URL is required for Generic provider type.',
//   path: ['baseUrl'],
// });

interface ProviderFormDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (provider: ProviderData) => void;
  initialData?: ProviderData | null;
}

export function ProviderFormDialog({ isOpen, onOpenChange, onSubmit, initialData }: ProviderFormDialogProps) {
  const { toast } = useToast();
  const [selectedType, setSelectedType] = useState<ProviderType | undefined>(initialData?.type);
  // Local state for input fields if not using react-hook-form
  const [name, setName] = useState(initialData?.name || '');
  const [apiKey, setApiKey] = useState(initialData?.apiKey || '');
  const [baseUrl, setBaseUrl] = useState(initialData?.baseUrl || '');


  // TODO: Restore form handling if react-hook-form and zod are added to package.json
  // const form = useForm<ProviderFormData>({ 
  //   resolver: zodResolver(providerFormSchema), 
  //   defaultValues: ... 
  // });

  useEffect(() => {
    if (initialData) {
      setName(initialData.name);
      setSelectedType(initialData.type);
      setApiKey(initialData.apiKey || '');
      setBaseUrl(initialData.baseUrl || '');
      // form.reset({ ... }); // Form logic removed
    } else {
      setName('');
      setSelectedType('OpenAI');
      setApiKey('');
      setBaseUrl('');
      // form.reset({ ... }); // Form logic removed
    }
  }, [initialData, isOpen]); // Removed form from dependencies
  
  // useEffect(() => { // Form logic removed
  //   const subscription = form.watch((value: ProviderFormData, { name }: { name: keyof ProviderFormData }) => {
  //     if (name === 'type' && value.type) {
  //       setSelectedType(value.type);
  //     }
  //   });
  //   return () => subscription.unsubscribe();
  // }, [form]);

  // TODO: Restore form handling if react-hook-form and zod are added to package.json
  const handleSubmitInternal = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault(); // Prevent default form submission
    // This is a placeholder. Actual data submission logic needs to be restored.
    const data: ProviderFormData = {
      name: name,
      type: selectedType || 'Generic', // Ensure selectedType has a default
      apiKey: apiKey,
      baseUrl: selectedType === 'Generic' ? baseUrl : undefined,
    };
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
            {/* TODO: Restore form handling if react-hook-form and zod are added to package.json - Form validation messages will be missing */}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmitInternal} className="grid gap-4 py-4"> {/* Changed to placeholder handleSubmitInternal */}
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">Name</Label>
            <Input id="name" name="name" className="col-span-3 bg-input border-border" value={name} onChange={(e) => setName(e.target.value)} />
            {/* {form.formState.errors.name && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.name.message}</p>} */} {/* Form error display removed */}
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="type" className="text-right">Type</Label>
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
                    <SelectItem key={type} value={type}>{type}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
            {/* {form.formState.errors.type && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.type.message}</p>} */} {/* Form error display removed */}
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="apiKey" className="text-right">API Key</Label>
            <Input id="apiKey" name="apiKey" type="password" className="col-span-3 bg-input border-border" value={apiKey} onChange={(e) => setApiKey(e.target.value)} />
            {/* {form.formState.errors.apiKey && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.apiKey.message}</p>} */} {/* Form error display removed */}
          </div>
          {selectedType === 'Generic' && (
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="baseUrl" className="text-right">Base URL</Label>
              <Input id="baseUrl" name="baseUrl" className="col-span-3 bg-input border-border" value={baseUrl} onChange={(e) => setBaseUrl(e.target.value)} />
              {/* {form.formState.errors.baseUrl && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.baseUrl.message}</p>} */} {/* Form error display removed */}
            </div>
          )}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            {/* TODO: Restore form handling if react-hook-form and zod are added to package.json - Button type might change */}
            <Button type="submit" variant="default">{initialData ? 'Save Changes' : 'Create Provider'}</Button> 
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
