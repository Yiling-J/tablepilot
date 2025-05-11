import { Source, TableCreateRequest, importImage } from "@/actions";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogOverlay,
    DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { JSONArray, JSONObject } from "@/json";
import { ReloadIcon } from "@radix-ui/react-icons";
import { Plus } from "lucide-react";
import Papa from "papaparse";
import { useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ModelSelector } from "../model-selector.tsx";

interface ImportCSVDialogProps {
  isOpen: boolean;
  setIsOpen: (v: boolean) => void;
  onNext: (form: TableCreateRequest, rows: JSONObject[]) => void;
}

export function ImportCSVDialog({
  isOpen,
  setIsOpen,
  onNext,
}: ImportCSVDialogProps) {
  const [loading, setLoading] = useState<boolean>(false);
  const [file, setFile] = useState<File | null>(null);
  const [fileName, setFileName] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isImage, setIsImage] = useState<boolean>(false);
  const [prompt, setPrompt] = useState("");
  const [model, setModel] = useState("");
  const navigate = useNavigate();

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  const handleSubmitCSV = async () => {
    try {
      if (!file) {
        return;
      }
      if (isImage) {
        const reader = new FileReader();
        reader.readAsDataURL(file);
        reader.onload = async () => {
          setLoading(true);
          const id = await importImage({
            prompt: prompt,
            model: model,
            data: (reader.result as string)
              .replace("data:", "")
              .replace(/^.+,/, ""),
          });
          setLoading(false);
          navigate(`/tables/${id}`);
        };
        return;
      }
      const input = await file.text();

      const parsed = Papa.parse(input, { skipEmptyLines: true });
      const records = parsed.data;
      if (records.length === 0) {
        // show err
        return;
      }
      const columns = records[0] as string[];

      const rows = records.slice(1).map((row) => {
        const obj: JSONObject = {};
        columns.forEach((col, index) => {
          obj[col] = (row as JSONArray)[index];
        });
        return obj;
      });
      const form = {
        name: file.name.substring(0, file.name.lastIndexOf(".")) || file.name,
        description: "",
        sources: new Array<Source>(),
        columns: columns.map((c) => {
          return {
            name: c,
            description: "",
            type: "string",
            fill_mode: "ai",
            random: true,
            replacement: false,
            repeat: 1,
            linked_column: "",
            linked_context_columns: [],
          };
        }),
      };
      onNext(form, rows);
    } finally {
      setFile(null);
      setFileName("");
      setIsImage(false);
      setPrompt("");
      setModel("");
      setLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogOverlay />
      <DialogContent
        onInteractOutside={(e) => {
          e.preventDefault();
        }}
      >
        <DialogTitle className="hidden">Import</DialogTitle>
        <DialogDescription className="hidden"></DialogDescription>

        <div className="w-full max-w-4xl mx-auto p-4">
          {isImage && (
            <div className="pb-3">
              <ModelSelector
                hasImageColumn={false}
                generating={false}
                selectModel={(model) => {
                  setModel(model);
                }}
                selectImageModel={(_) => {}}
              />
            </div>
          )}
          <Card
            className="border-dashed border-2 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-900 transition-colors"
            onClick={handleClick}
          >
            <CardContent className="flex flex-col items-center justify-center py-12">
              <input
                type="file"
                data-testid="import-file-selector"
                ref={fileInputRef}
                className="hidden"
                accept=".csv, .png, .jpg, .jpeg"
                onClick={(e) => {
                  e.stopPropagation();
                }}
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) {
                    setFile(file);
                    setFileName(file.name);
                    if (
                      ["image/png", "image/jpeg", "image/jpg"].includes(
                        file.type,
                      )
                    ) {
                      setIsImage(true);
                    }
                  }
                }}
              />
              <div className="h-12 w-12 rounded-full bg-primary/10 flex items-center justify-center mb-4">
                <Plus className="h-6 w-6 text-primary" />
              </div>
              <p className="text-sm text-muted-foreground mb-1">
                {fileName ? fileName : "Click to select a CSV or image file"}
              </p>
              <p className="text-xs text-muted-foreground">
                {fileName ? "Click to change file" : ""}
              </p>
            </CardContent>
          </Card>
        </div>
        {isImage && (
          <div className="grid gap-2 px-5">
            <Label htmlFor="prompt" className="flex items-center gap-1">
              Prompt
              <span className="text-xs text-muted-foreground">(optional)</span>
            </Label>
            <Textarea
              id="prompt"
              placeholder="Provide specific instructions to guide the AI in extracting the right table from your image."
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              className="resize-none"
              rows={3}
            />
          </div>
        )}
        <div className="flex justify-end">
          <Button
            variant="outline"
            className="mr-2"
            disabled={loading}
            onClick={() => setIsOpen(false)}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmitCSV}
            disabled={!file || loading || (isImage && model === "")}
          >
            {loading ? <ReloadIcon className="animate-spin" /> : "Next"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
