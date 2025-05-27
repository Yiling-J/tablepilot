'use client';

import { useState, useEffect } from 'react'; // Removed: useForm, SubmitHandler, zodResolver, z
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
import { Textarea } from '@/components/ui/textarea';
import type { ProviderType, ModelData } from '@/types.ts'; // Removed AISuggestionInput, AISuggestionOutput
// import { suggestModelConfiguration } from '@/ai/flows/suggest-model-configuration'; // Removed
import { useToast } from '@/hooks/use-toast';
import { Loader2 } from 'lucide-react';

// const optimizeSchema = z.object({ // Schema removed
//   usagePatterns: z.string().min(10, 'Please describe usage patterns in at least 10 characters.'),
//   currentMaxTokens: z.coerce.number().optional(),
//   currentRpm: z.coerce.number().optional(),
// });

// type OptimizeFormData = z.infer<typeof optimizeSchema>; // Type removed

interface OptimizeConfigDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  // onApplyOptimization: (optimizedValues: { max_tokens: number; rpm: number }) => void; // Removed prop
  providerType: ProviderType;
  model: ModelData | null;
}

export function OptimizeConfigDialog({
  isOpen,
  onOpenChange,
  // onApplyOptimization, // Removed prop
  providerType,
  model,
}: OptimizeConfigDialogProps) {
  const { toast } = useToast();
  const [isLoading, setIsLoading] = useState(false);
  // const [suggestion, setSuggestion] = useState<AISuggestionOutput | null>(null); // Removed state

  // TODO: Restore form handling if react-hook-form and zod are added to package.json
  // const form = useForm<OptimizeFormData>({
  //   resolver: zodResolver(optimizeSchema),
  //   defaultValues: {
  //     usagePatterns: '',
  //     currentMaxTokens: model?.max_tokens || 0,
  //     currentRpm: model?.rpm || 0,
  //   },
  // });

  useEffect(() => {
    if (model && isOpen) {
      // form.reset({ // Form logic removed
      //   usagePatterns: '',
      //   currentMaxTokens: model.max_tokens,
      //   currentRpm: model.rpm,
      // });
      // setSuggestion(null); // State removed
    }
  }, [model, isOpen]); // Removed form from dependencies

  // TODO: Restore form handling if react-hook-form and zod are added to package.json
  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault(); // Prevent default form submission
    if (!model) return;
    setIsLoading(true);
    // setSuggestion(null); // State removed

    // const input: AISuggestionInput = { // Unused variable 'input'
    //   providerType: providerType,
    //   usagePatterns: data.usagePatterns, // data is not defined here anymore
    //   currentMaxTokens: data.currentMaxTokens, // data is not defined here anymore
    //   currentRpm: data.currentRpm, // data is not defined here anymore
    // };
    // if (!model) return; // Duplicate check removed
    // setIsLoading(true); // Duplicate call removed
    // setSuggestion(null); // Duplicate call removed

    // const input: AISuggestionInput = { // Unused variable 'input'
    //   providerType: providerType,
    //   usagePatterns: data.usagePatterns,
    //   currentMaxTokens: data.currentMaxTokens,
    //   currentRpm: data.currentRpm,
    // };

    // try {
    //   // const result = await suggestModelConfiguration(input); // Removed
    //   // setSuggestion(result);
    //   // toast({ title: "Suggestion Ready", description: "AI has provided an optimization suggestion." });
    //   // console.log("AI Suggestion feature called with input:", input); 
    // } catch (error) {
    //   console.error("Error fetching AI suggestion:", error);
    //   toast({ variant: "destructive", title: "Error", description: "Failed to get AI suggestion." });
    // } finally {
    //   setIsLoading(false);
    // }
    setIsLoading(false); // Ensure loading state is reset
    toast({ title: "AI Suggestion (Disabled)", description: "This feature is currently disabled as the AI service is not available." });
  };

  // const handleApply = () => { // Unused function 'handleApply'
  //   // This function will effectively not be callable if 'suggestion' state is always null
  //   if (suggestion) {
  //     onApplyOptimization({
  //       max_tokens: suggestion.suggestedMaxTokens,
  //       rpm: suggestion.suggestedRpm,
  //     });
  //     toast({ title: "Optimization Applied", description: "Model configuration has been updated." });
  //     onOpenChange(false);
  //   }
  // };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => { if (!isLoading) onOpenChange(open); }}>
      <DialogContent className="sm:max-w-md bg-card text-card-foreground">
        <DialogHeader>
          <DialogTitle>Optimize Model Configuration (AI)</DialogTitle>
          <DialogDescription>
            Describe usage patterns for {model?.model} ({providerType}). 
            Note: AI-powered suggestion feature is currently disabled.
            {/* TODO: Restore form handling if react-hook-form and zod are added to package.json - Form validation messages will be missing */}
          </DialogDescription>
        </DialogHeader>
        {/* Since 'suggestion' state is removed, this always shows the form. */}
        <form onSubmit={handleSubmit} className="space-y-4"> {/* Changed to placeholder handleSubmit */}
          <div>
            <Label htmlFor="usagePatterns">Usage Patterns</Label>
              <Textarea
                id="usagePatterns"
                name="usagePatterns" // Added name for FormData
                placeholder="e.g., High volume summarization tasks, low latency chat bot interactions..."
                className="bg-input border-border min-h-[100px]"
                defaultValue={''} // Changed from model?.usagePatterns
              />
              {/* {form.formState.errors.usagePatterns && <p className="text-sm text-destructive">{form.formState.errors.usagePatterns.message}</p>} */} {/* Form error display removed */}
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="currentMaxTokens">Current Max Tokens</Label>
                <Input id="currentMaxTokens" name="currentMaxTokens" type="number" className="bg-input border-border" defaultValue={model?.max_tokens || 0} />
              </div>
              <div>
                <Label htmlFor="currentRpm">Current RPM</Label>
                <Input id="currentRpm" name="currentRpm" type="number" className="bg-input border-border" defaultValue={model?.rpm || 0} />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading}>Cancel</Button>
              <Button type="submit" variant="default" disabled={isLoading}>
                {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Get Suggestion (Disabled)
              </Button>
            </DialogFooter>
          </form>
        {/* ) : ( // This part is effectively removed as 'suggestion' will always be null
          <div className="space-y-4">
            <h3 className="font-semibold text-lg text-primary">AI Suggestion:</h3>
            <p><strong>Max Tokens:</strong> <span className="text-primary">{suggestion.suggestedMaxTokens}</span></p>
            <p><strong>RPM:</strong> <span className="text-primary">{suggestion.suggestedRpm}</span></p>
            <p><strong>Reasoning:</strong></p>
            <p className="text-sm text-muted-foreground p-3 bg-input rounded-md">{suggestion.reasoning}</p>
            <DialogFooter>
              <Button variant="outline" onClick={() => setSuggestion(null)} disabled={isLoading}>Refine Input</Button>
              <Button variant="default" onClick={handleApply} disabled={isLoading}>Apply Suggestion</Button>
            </DialogFooter>
          </div> */}
        </DialogContent>
    </Dialog>
  );
}
