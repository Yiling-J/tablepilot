import { TableCreateRequest } from "@/actions";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { Copy } from "lucide-react";

interface JsonPreviewProps {
  formData: TableCreateRequest;
}

export function JsonPreview({ formData }: JsonPreviewProps) {
  const { toast } = useToast();

  const handleCopyJson = () => {
    navigator.clipboard.writeText(JSON.stringify(formData, null, 2));
    toast({
      title: "Copied to clipboard",
      description: "The JSON has been copied to your clipboard",
    });
  };

  return (
    <div className="relative">
      <Button
        variant="outline"
        size="sm"
        className="absolute right-3 top-3"
        onClick={handleCopyJson}
      >
        <Copy className="h-4 w-4 mr-2" />
        Copy
      </Button>
      <pre
        className="bg-muted p-2 rounded-md text-sm"
        data-testid="json-preview"
      >
        {JSON.stringify(formData, null, 2)}
      </pre>
    </div>
  );
}
