'use client';

import { useState, useEffect } from 'react';
import { useForm, type SubmitHandler } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
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
import type { ProviderType, ModelData, AISuggestionInput, AISuggestionOutput } from '@/types';
import { suggestModelConfiguration } from '@/ai/flows/suggest-model-configuration';
import { useToast } from '@/hooks/use-toast';
import { Loader2 } from 'lucide-react';

const optimizeSchema = z.object({
  usagePatterns: z.string().min(10, 'Please describe usage patterns in at least 10 characters.'),
  currentMaxTokens: z.coerce.number().optional(),
  currentRpm: z.coerce.number().optional(),
});

type OptimizeFormData = z.infer<typeof optimizeSchema>;

interface OptimizeConfigDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onApplyOptimization: (optimizedValues: { max_tokens: number; rpm: number }) => void;
  providerType: ProviderType;
  model: ModelData | null;
}

export function OptimizeConfigDialog({
  isOpen,
  onOpenChange,
  onApplyOptimization,
  providerType,
  model,
}: OptimizeConfigDialogProps) {
  const { toast } = useToast();
  const [isLoading, setIsLoading] = useState(false);
  const [suggestion, setSuggestion] = useState<AISuggestionOutput | null>(null);

  const form = useForm<OptimizeFormData>({
    resolver: zodResolver(optimizeSchema),
    defaultValues: {
      usagePatterns: '',
      currentMaxTokens: model?.max_tokens || 0,
      currentRpm: model?.rpm || 0,
    },
  });

  useEffect(() => {
    if (model && isOpen) {
      form.reset({
        usagePatterns: '',
        currentMaxTokens: model.max_tokens,
        currentRpm: model.rpm,
      });
      setSuggestion(null); // Reset suggestion when dialog opens or model changes
    }
  }, [model, isOpen, form]);

  const handleSubmit: SubmitHandler<OptimizeFormData> = async (data) => {
    if (!model) return;
    setIsLoading(true);
    setSuggestion(null);

    const input: AISuggestionInput = {
      providerType: providerType,
      usagePatterns: data.usagePatterns,
      currentMaxTokens: data.currentMaxTokens,
      currentRpm: data.currentRpm,
    };

    try {
      const result = await suggestModelConfiguration(input);
      setSuggestion(result);
      toast({ title: "Suggestion Ready", description: "AI has provided an optimization suggestion." });
    } catch (error) {
      console.error("Error fetching AI suggestion:", error);
      toast({ variant: "destructive", title: "Error", description: "Failed to get AI suggestion." });
    } finally {
      setIsLoading(false);
    }
  };

  const handleApply = () => {
    if (suggestion) {
      onApplyOptimization({
        max_tokens: suggestion.suggestedMaxTokens,
        rpm: suggestion.suggestedRpm,
      });
      toast({ title: "Optimization Applied", description: "Model configuration has been updated." });
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => { if (!isLoading) onOpenChange(open); }}>
      <DialogContent className="sm:max-w-md bg-card text-card-foreground">
        <DialogHeader>
          <DialogTitle>Optimize Model Configuration (AI)</DialogTitle>
          <DialogDescription>
            Get AI-powered suggestions for {model?.model} ({providerType}) based on usage patterns.
          </DialogDescription>
        </DialogHeader>
        {!suggestion ? (
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <div>
              <Label htmlFor="usagePatterns">Usage Patterns</Label>
              <Textarea
                id="usagePatterns"
                {...form.register('usagePatterns')}
                placeholder="e.g., High volume summarization tasks, low latency chat bot interactions..."
                className="bg-input border-border min-h-[100px]"
              />
              {form.formState.errors.usagePatterns && <p className="text-sm text-destructive">{form.formState.errors.usagePatterns.message}</p>}
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="currentMaxTokens">Current Max Tokens</Label>
                <Input id="currentMaxTokens" type="number" {...form.register('currentMaxTokens')} className="bg-input border-border" />
              </div>
              <div>
                <Label htmlFor="currentRpm">Current RPM</Label>
                <Input id="currentRpm" type="number" {...form.register('currentRpm')} className="bg-input border-border" />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading}>Cancel</Button>
              <Button type="submit" variant="primary" disabled={isLoading}>
                {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Get Suggestion
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <div className="space-y-4">
            <h3 className="font-semibold text-lg text-primary">AI Suggestion:</h3>
            <p><strong>Max Tokens:</strong> <span className="text-primary">{suggestion.suggestedMaxTokens}</span></p>
            <p><strong>RPM:</strong> <span className="text-primary">{suggestion.suggestedRpm}</span></p>
            <p><strong>Reasoning:</strong></p>
            <p className="text-sm text-muted-foreground p-3 bg-input rounded-md">{suggestion.reasoning}</p>
            <DialogFooter>
              <Button variant="outline" onClick={() => setSuggestion(null)} disabled={isLoading}>Refine Input</Button>
              <Button variant="primary" onClick={handleApply} disabled={isLoading}>Apply Suggestion</Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
