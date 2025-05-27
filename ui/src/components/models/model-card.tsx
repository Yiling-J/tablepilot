
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { ModelData } from "@/types.ts";
import { Edit3, Trash2, XCircle, CheckCircle2 } from "lucide-react"; 
// import Balancer from "react-wrap-balancer"; // Removed


interface ModelCardProps {
  model: ModelData;
  onEdit: () => void;
  onDelete: () => void;
  isProviderEditable: boolean;
  isProviderEnabled: boolean;
}

export function ModelCard({ model, onEdit, onDelete, isProviderEditable, isProviderEnabled }: ModelCardProps) {
  const canInteract = isProviderEditable && isProviderEnabled;

  return (
    <Card className={`bg-card/80 backdrop-blur-sm shadow-lg hover:shadow-primary/30 transition-all duration-300 flex flex-col h-full ${!isProviderEnabled && isProviderEditable ? 'opacity-60' : ''}`}>
      <CardHeader className="pb-3">
        <div className="flex justify-between items-start">
          <div>
            <CardTitle className="text-lg font-semibold text-card-foreground">
              {model.alias || model.model}
            </CardTitle>
            {model.alias && model.alias !== model.model && (
              <CardDescription className="text-xs text-muted-foreground">{model.model}</CardDescription>
            )}
          </div>
          {model.isDefault && <Badge variant="secondary" className="text-xs shrink-0">Default</Badge>}
        </div>
      </CardHeader>
      <CardContent className="flex-grow space-y-3 text-sm">
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Max Tokens:</span>
          <span className="font-medium text-card-foreground">{(model.max_tokens ?? 0) > 0 ? (model.max_tokens ?? 0).toLocaleString() : 'N/A'}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">RPM Limit:</span>
          <span className="font-medium text-card-foreground">{(model.rpm ?? 0) > 0 ? (model.rpm ?? 0).toLocaleString() : 'N/A'}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Image Generation Support:</span>
          {model.imageSupport ? (
            <CheckCircle2 className="h-5 w-5 text-green-500" />
          ) : (
            <XCircle className="h-5 w-5 text-red-500" />
          )}
        </div>
      </CardContent>
      {/* Separator and Footer are now always rendered */}
      <>
        <Separator className="my-2 bg-border/50" />
        <CardFooter className="flex justify-end gap-2 pt-3 pb-3 px-4">
          <Button variant="ghost" size="icon" onClick={onEdit} title="Edit Model" disabled={!canInteract}>
            <Edit3 className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" onClick={onDelete} title="Delete Model" className="text-destructive hover:text-destructive hover:bg-destructive/10" disabled={!canInteract}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </CardFooter>
      </>
    </Card>
  );
}
