'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import type { ProviderData } from '@/types';
import { useToast } from '@/hooks/use-toast';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface ImportExportDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  currentConfig: ProviderData[];
  onImport: (config: ProviderData[]) => void;
}

export function ImportExportDialog({ isOpen, onOpenChange, currentConfig, onImport }: ImportExportDialogProps) {
  const { toast } = useToast();
  const [importJson, setImportJson] = useState('');

  const handleExport = () => {
    const jsonString = JSON.stringify(currentConfig, null, 2);
    const blob = new Blob([jsonString], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'model_config.json';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast({ title: "Configuration Exported", description: "Your current configuration has been downloaded." });
  };

  const handleImport = () => {
    try {
      const parsedConfig = JSON.parse(importJson);
      // Basic validation (can be improved with Zod schema validation)
      if (Array.isArray(parsedConfig) && parsedConfig.every(p => p.id && p.name && p.type && p.models)) {
        onImport(parsedConfig as ProviderData[]);
        toast({ title: "Configuration Imported", description: "Configuration has been successfully imported." });
        onOpenChange(false);
        setImportJson('');
      } else {
        throw new Error('Invalid configuration format.');
      }
    } catch (error) {
      console.error("Import error:", error);
      toast({ variant: "destructive", title: "Import Error", description: "Failed to import configuration. Please check the JSON format." });
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg bg-card text-card-foreground">
        <DialogHeader>
          <DialogTitle>Import / Export Configuration</DialogTitle>
          <DialogDescription>
            Manage your Model configurations by importing or exporting them as JSON.
          </DialogDescription>
        </DialogHeader>
        <Tabs defaultValue="export" className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="export">Export</TabsTrigger>
            <TabsTrigger value="import">Import</TabsTrigger>
          </TabsList>
          <TabsContent value="export" className="py-4 space-y-4">
            <p className="text-sm text-muted-foreground">
              Export your current provider and model configurations as a JSON file.
              You can use this file to backup your settings or transfer them to another instance.
            </p>
            <Button onClick={handleExport} className="w-full" variant="primary">Export Configuration</Button>
          </TabsContent>
          <TabsContent value="import" className="py-4 space-y-4">
            <div>
              <Label htmlFor="importJson">Paste JSON Configuration</Label>
              <Textarea
                id="importJson"
                value={importJson}
                onChange={(e) => setImportJson(e.target.value)}
                placeholder='[{"id": "...", "name": "...", ...}]'
                className="min-h-[200px] bg-input border-border"
              />
            </div>
             <Button onClick={handleImport} className="w-full" variant="primary">Import Configuration</Button>
          </TabsContent>
        </Tabs>
        <DialogFooter className="sm:justify-start pt-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
