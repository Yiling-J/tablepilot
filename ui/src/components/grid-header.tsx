import { GenerateRequest } from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { NumberInput } from "@/components/ui/number-input";
import { Separator } from "@/components/ui/separator";
import { Slider } from "@/components/ui/slider";
import { Cross1Icon, ReloadIcon } from "@radix-ui/react-icons";
import { MutableRefObject, useEffect, useState } from "react";

export function GridHeader({
  genRequestRef,
  clearData,
}: {
  clearData: () => Promise<void>;
  genRequestRef: MutableRefObject<GenerateRequest>;
}) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [sliderValue, setSliderValue] = useState(0.6);
  const [batch, setBatch] = useState(10);
  const [count, setCount] = useState(50);

  useEffect(() => {
    setBatch(genRequestRef.current.batch);
    setCount(genRequestRef.current.count);
    setSliderValue(genRequestRef.current.temperature);
  });

  const handleClearData = async () => {
    setIsLoading(true);
    await clearData();
    setDialogOpen(false);
    setIsLoading(false);
  };

  return (
    <div>
      <header className="flex h-12 items-center gap-4 border-b bg-transparent px-4 md:px-6 pb-2">
        <div className="flex h-5 items-center space-x-4 text-sm">
          <div className="select-none">Count</div>
          <NumberInput
            value={count}
            defaultValue={count}
            onValueChange={(v) => {
              genRequestRef.current.count = v ?? 50;
              setCount(genRequestRef.current.count);
            }}
          />
          <Separator orientation="vertical" />
          <div className="select-none">Batch</div>
          <NumberInput
            value={batch}
            defaultValue={batch}
            onValueChange={(v) => {
              genRequestRef.current.batch = v ?? 10;
              setBatch(genRequestRef.current.batch);
            }}
          />
          <Separator orientation="vertical" />

          <div className="flex ml-4">Temperature</div>
          <Slider
            min={0}
            value={[genRequestRef.current.temperature]}
            max={2}
            step={0.01}
            className="w-36 ml-4"
            onValueChange={(v: Array<number>) => {
              setSliderValue(Number(v[0]));
              genRequestRef.current.temperature = v[0];
            }}
          />
          <div className="w-[25px]">
            {sliderValue.toLocaleString("en-us", { minimumFractionDigits: 2 })}
          </div>

          <Separator orientation="vertical" />

          <div
            className={`flex items-center gap-2 cursor-pointer ${isLoading ? "opacity-50 cursor-not-allowed" : ""}`}
            onClick={() => !isLoading && setDialogOpen(true)}
          >
            {isLoading ? (
              <ReloadIcon className="animate-spin" />
            ) : (
              <Cross1Icon />
            )}

            <div className="select-none">Truncate table</div>
          </div>
        </div>
      </header>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Clear Data</DialogTitle>
          </DialogHeader>
          <DialogDescription>
            Are you sure you want to clear all data?
          </DialogDescription>
          <DialogFooter>
            <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleClearData}>Confirm</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
