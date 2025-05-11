import { ModelList, getModels } from "@/actions";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import { useEffect, useState } from "react";
import { ProvidersListDialog } from "./dialog/providers.tsx";
import { Button } from "./ui/button";

import {
    Select,
    SelectContent,
    SelectGroup,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { GearIcon } from "@radix-ui/react-icons";
import * as SelectPrimitive from "@radix-ui/react-select";
import { BookTypeIcon, Check, ImageIcon } from "lucide-react";

interface ModelSelectorProps {
  hasImageColumn: boolean;
  generating: boolean;
  selectModel: (model: string) => void;
  selectImageModel: (model: string) => void;
}

export function ModelSelector({
  hasImageColumn,
  generating,
  selectModel,
  selectImageModel,
}: ModelSelectorProps) {
  const [models, setModels] = useState<ModelList | undefined>(undefined);
  const [model, setModel] = useState("");
  const [modelSelectOpen, setModelSelectOpen] = useState(false);
  const [providerListOpen, setProviderListOpen] = useState(false);
  const [imageModel, setImageModel] = useState("");
  const [imageModelSelectOpen, setImageModelSelectOpen] = useState(false);

  const fetchData = async () => {
    const models = await getModels();
    setModels(models);
    setModel(models.default_model);
    setImageModel(models.default_image_model);
    selectModel(models.default_model);
    selectImageModel(models.default_image_model);
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <div className="flex">
      <ProvidersListDialog
        isOpen={providerListOpen}
        setIsOpen={async (v) => {
          const models = await getModels();
          setModels(models);
          setProviderListOpen(v);
        }}
      />
      {(models === undefined ||
        models?.models === null ||
        models.models.length === 0) && (
        <div className="flex border rounded-sm">
          <Button
            variant="outline"
            onClick={() => {
              setProviderListOpen(true);
            }}
          >
            Add Models
          </Button>
        </div>
      )}
      {models && models.models && models.models.length > 0 && (
        <div
          className={cn(
            "flex rounded-sm items-center",
            hasImageColumn ? "" : "border",
          )}
        >
          {hasImageColumn && <BookTypeIcon className="mr-2" />}
          <Select
            value={model}
            disabled={generating}
            onValueChange={async (v) => {
              selectModel(v);
              setModel(v);
            }}
            open={modelSelectOpen}
            onOpenChange={setModelSelectOpen}
          >
            <SelectTrigger className="w-[180px] ring-0 border-0 focus:ring-offset-0 focus:ring-0 focus:border-0">
              <SelectValue placeholder="Select a model" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                {models?.models.map((model) => (
                  <SelectPrimitive.Item
                    value={model.name}
                    key={model.name}
                    className={cn(
                      "relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
                      "",
                    )}
                  >
                    <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
                      <SelectPrimitive.ItemIndicator>
                        <Check className="h-4 w-4" />
                      </SelectPrimitive.ItemIndicator>
                    </span>

                    <div>
                      <SelectPrimitive.ItemText>
                        <p>{model.name}</p>
                      </SelectPrimitive.ItemText>
                    </div>
                  </SelectPrimitive.Item>
                ))}
              </SelectGroup>
              <Separator className="my-1" />

              <div>
                <div
                  className="flex pt-2 pl-2 pb-2 text-sm items-center cursor-pointer"
                  onClick={() => {
                    setModelSelectOpen(false);
                    setProviderListOpen(true);
                  }}
                >
                  <GearIcon className="mr-1" />
                  <p>Model Settings</p>
                </div>
              </div>
            </SelectContent>
          </Select>
        </div>
      )}

      {hasImageColumn &&
        models &&
        models.models &&
        models.models.length > 0 && (
          <div className="flex ml-4 rounded-sm items-center">
            <ImageIcon className="mr-2" />
            <Select
              value={imageModel}
              disabled={generating}
              onValueChange={async (v) => {
                selectImageModel(v);
                setImageModel(v);
              }}
              open={imageModelSelectOpen}
              onOpenChange={setImageModelSelectOpen}
            >
              <SelectTrigger className="w-[200px] ring-0 border-0 focus:ring-offset-0 focus:ring-0 focus:border-0">
                <SelectValue placeholder="Select image gen model" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {models?.models
                    .filter((m) => m.image)
                    .map((model) => (
                      <SelectPrimitive.Item
                        value={model.name}
                        key={model.name}
                        className={cn(
                          "relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
                          "",
                        )}
                      >
                        <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
                          <SelectPrimitive.ItemIndicator>
                            <Check className="h-4 w-4" />
                          </SelectPrimitive.ItemIndicator>
                        </span>

                        <div>
                          <SelectPrimitive.ItemText>
                            <p>{model.name}</p>
                          </SelectPrimitive.ItemText>
                        </div>
                      </SelectPrimitive.Item>
                    ))}
                </SelectGroup>
                <Separator className="my-1" />

                <div>
                  <div
                    className="flex pt-2 pl-2 pb-2 text-sm items-center cursor-pointer"
                    onClick={() => {
                      setModelSelectOpen(false);
                      setProviderListOpen(true);
                    }}
                  >
                    <GearIcon className="mr-1" />
                    <p>Model Settings</p>
                  </div>
                </div>
              </SelectContent>
            </Select>
          </div>
        )}
    </div>
  );
}
