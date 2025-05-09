import { TableInfo } from "@/actions";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { Check, Sparkles } from "lucide-react";
import { useState } from "react";

interface RegenerateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRegenerate: (config: RegenerateConfig) => void;
  tableInfo: TableInfo;
}

interface RegenerateConfig {
  columnsToRegenerate: string[];
  prompt: string;
}

export function RegenerateDialog({
  open,
  onOpenChange,
  onRegenerate,
  tableInfo,
}: RegenerateDialogProps) {
  const [columnsToRegenerate, setColumnsToRegenerate] = useState<string[]>([]);
  const [prompt, setPrompt] = useState("");

  const handleRegenerateClick = () => {
    const config: RegenerateConfig = {
      columnsToRegenerate,
      prompt,
    };
    onRegenerate(config);
    onOpenChange(false);
  };

  const toggleRegenerateColumn = (columnId: string) => {
    setColumnsToRegenerate((prev) =>
      prev.includes(columnId)
        ? prev.filter((id) => id !== columnId)
        : [...prev, columnId],
    );
  };

  const toggleAllColumns = (checked: boolean) => {
    if (checked) {
      setColumnsToRegenerate(tableInfo.columns.map((col) => col.id));
    } else {
      setColumnsToRegenerate([]);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[550px] p-0 overflow-hidden">
        <DialogHeader className="px-6 pt-6 pb-2">
          <div className="flex items-center gap-2 mb-1">
            <Sparkles className="h-5 w-5" />
            <DialogTitle className="text-xl font-semibold">
              Regenerate
            </DialogTitle>
          </div>
          <DialogDescription className="text-sm">
            Regenerate selected rows
          </DialogDescription>
        </DialogHeader>

        <div className="px-6 py-4 space-y-6">
          <div>
            <div className="flex items-center justify-between mb-3">
              <Label className="text-sm font-medium">
                Columns to Regenerate
              </Label>
              <div className="flex items-center gap-2">
                <Label
                  htmlFor="select-all-columns"
                  className="text-sm font-medium cursor-pointer"
                >
                  {columnsToRegenerate.length === tableInfo.columns.length
                    ? "Deselect all"
                    : "Select all"}
                </Label>
                <Switch
                  id="select-all-columns"
                  checked={
                    columnsToRegenerate.length === tableInfo.columns.length
                  }
                  onCheckedChange={toggleAllColumns}
                />
              </div>
            </div>

            <Card>
              <CardContent className="p-2">
                <ScrollArea className="h-[180px] pr-4">
                  <div className="space-y-1 py-1">
                    {tableInfo.columns.map((column) => (
                      <div
                        key={column.id}
                        className={cn(
                          "flex items-center justify-between px-3 py-2 rounded-md cursor-pointer transition-all duration-200",
                          "hover:bg-muted/50",
                          columnsToRegenerate.includes(column.id) &&
                            "bg-accent",
                        )}
                        onClick={() => toggleRegenerateColumn(column.id)}
                      >
                        <div className="flex items-center gap-3">
                          <div
                            className={cn(
                              "flex items-center justify-center w-5 h-5 rounded-md border",
                              columnsToRegenerate.includes(column.id) &&
                                "bg-primary border-primary",
                            )}
                          >
                            {columnsToRegenerate.includes(column.id) && (
                              <Check className="h-3.5 w-3.5 text-primary-foreground" />
                            )}
                          </div>
                          <span
                            data-testid="regen-span"
                            className={cn(
                              "font-medium",
                              columnsToRegenerate.includes(column.id) &&
                                "text-accent-foreground",
                            )}
                          >
                            {column.name}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>

            {columnsToRegenerate.length > 0 && (
              <div className="mt-2 text-sm text-muted-foreground">
                {columnsToRegenerate.length} column
                {columnsToRegenerate.length !== 1 && "s"} selected
              </div>
            )}
          </div>

          {/* Instructions */}
          <div>
            <Label className="text-sm font-medium mb-2 block">Prompt</Label>
            <Card>
              <CardContent className="p-3">
                <Textarea
                  placeholder="E.g., Make the product descriptions more engaging and highlight key features."
                  className="min-h-[120px] resize-none"
                  value={prompt}
                  onChange={(e) => setPrompt(e.target.value)}
                />
              </CardContent>
            </Card>
          </div>
        </div>

        <DialogFooter className="px-6 py-4 border-t">
          {" "}
          <div className="flex items-center justify-end w-full gap-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleRegenerateClick}
              disabled={columnsToRegenerate.length === 0}
            >
              <Sparkles className="mr-2 h-4 w-4" />
              Regenerate
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
