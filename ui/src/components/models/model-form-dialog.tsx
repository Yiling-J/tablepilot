
'use client';

import { useEffect } from 'react'; // Removed: useState, useForm, SubmitHandler, zodResolver, z
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
// Unused Select imports removed
import type { ModelData, ModelFormData, ProviderType } from '@/types.ts';
// import { predefinedModels } from '@/lib/mock-data'; // Removed
import { useToast } from '@/hooks/use-toast';

// const modelFormSchema = z.object({ // Schema removed
//   model: z.string().min(1, 'Model name is required.'),
//   alias: z.string().optional(),
//   max_tokens: z.coerce.number().int().min(0, 'Max tokens must be a non-negative integer.'),
//   rpm: z.coerce.number().int().min(0, 'RPM must be a non-negative integer.'),
//   isDefault: z.boolean().optional(),
//   imageSupport: z.boolean(),
// });

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
  // Logic related to predefinedModels has been simplified or removed
  // const isKnownProviderWithPredefinedModels = false; // Simplified: always false
  // const [useCustomModelName, setUseCustomModelName] = useState(true); // Unused state variable removed

  // TODO: Restore form handling if react-hook-form and zod are added to package.json
  // const form = useForm<ModelFormData>({
  //   resolver: zodResolver(modelFormSchema),
  //   // Default values are set in the useEffect hook based on initialData and useCustomModelName
  // });

  useEffect(() => {
    if (!isOpen) return; 

    // let defaultModelValue = ''; // Unused variable removed
    // let initialCustomNameState = true; 

    if (initialData) {
      // initialCustomNameState = true; 
      // defaultModelValue = initialData.model; // Unused assignment
      // form.reset({ // Form logic removed
      //   ...initialData,
      //   alias: initialData.alias || '',
      // });
    } else {
      // initialCustomNameState = true; 
      // defaultModelValue = '';  // Unused assignment
      // form.reset({ // Form logic removed
      //   model: defaultModelValue,
      //   alias: '',
      //   max_tokens: 6000, 
      //   rpm: 10,         
      //   isDefault: false,
      //   imageSupport: false,
      // });
    }
    
    // setUseCustomModelName(initialCustomNameState); // Unused state setter removed
    // form.setValue('model', defaultModelValue, { shouldValidate: initialData ? true : false }); // Form logic removed

  }, [initialData, isOpen, providerType]); // Removed form and isKnownProviderWithPredefinedModels from dependencies


  // const handleToggleCustomModelName = (checked: boolean) => { // Unused function removed
  //   setUseCustomModelName(checked);
  // };

  // TODO: Restore form handling if react-hook-form and zod are added to package.json
  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault(); // Prevent default form submission
    // This is a placeholder. Actual data submission logic needs to be restored.
    // Example: Manually construct data from form inputs if not using react-hook-form
    const formData = new FormData(event.currentTarget);
    const data: ModelFormData = {
      model: formData.get('model') as string || '',
      alias: formData.get('alias') as string || undefined,
      max_tokens: parseInt(formData.get('max_tokens') as string || '0', 10),
      rpm: parseInt(formData.get('rpm') as string || '0', 10),
      isDefault: (formData.get('isDefault') === 'on'),
      imageSupport: (formData.get('imageSupport') === 'on'),
    };

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
            {/* TODO: Restore form handling if react-hook-form and zod are added to package.json - Form validation messages will be missing */}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="grid gap-4 py-4"> {/* Changed to placeholder handleSubmit */}
          
          {/* {isKnownProviderWithPredefinedModels && ( // This section is removed as predefinedModels logic is gone
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
          )} */}

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="model" className="text-right">Model Name</Label>
            <div className="col-span-3">
              <Input id="model" name="model" className="w-full bg-input border-border" defaultValue={initialData?.model || ''} />
            </div>
            {/* {form.formState.errors.model && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.model.message}</p>} */} {/* Form error display removed */}
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="alias" className="text-right">Alias</Label>
            <Input id="alias" name="alias" className="col-span-3 bg-input border-border" placeholder="Optional display name" defaultValue={initialData?.alias || ''} />
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="max_tokens" className="text-right">Max Tokens</Label>
            <Input id="max_tokens" name="max_tokens" type="number" className="col-span-3 bg-input border-border" defaultValue={initialData?.max_tokens || 6000} />
            {/* {form.formState.errors.max_tokens && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.max_tokens.message}</p>} */} {/* Form error display removed */}
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="rpm" className="text-right">RPM</Label>
            <Input id="rpm" name="rpm" type="number" className="col-span-3 bg-input border-border" defaultValue={initialData?.rpm || 10} />
            {/* {form.formState.errors.rpm && <p className="col-span-4 text-sm text-destructive text-right">{form.formState.errors.rpm.message}</p>} */} {/* Form error display removed */}
          </div>
          
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="imageSupport" className="text-right">Image Support</Label>
            <div className="col-span-3 flex items-center">
              <Switch
                id="imageSupport"
                name="imageSupport"
                defaultChecked={initialData?.imageSupport || false}
                // checked={form.watch('imageSupport')} // Form logic removed
                // onCheckedChange={(checked) => form.setValue('imageSupport', checked)} // Form logic removed
              />
            </div>
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="isDefault" className="text-right">Set as Default</Label>
             <div className="col-span-3 flex items-center">
              <Switch
                id="isDefault"
                name="isDefault"
                defaultChecked={initialData?.isDefault || false}
                // checked={form.watch('isDefault')} // Form logic removed
                // onCheckedChange={(checked) => form.setValue('isDefault', checked)} // Form logic removed
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            {/* TODO: Restore form handling if react-hook-form and zod are added to package.json - Button type might change */}
            <Button type="submit" variant="default">{initialData ? 'Save Changes' : 'Add Model'}</Button> 
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
