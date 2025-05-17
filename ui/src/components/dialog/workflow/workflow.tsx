import { Workflow, runWorkflow } from "@/actions";
import { DataCard } from "@/components/dialog/workflow/data-card";
import { WorkflowSteps } from "@/components/dialog/workflow/workflow-steps";
import { ModelSelector } from "@/components/model-selector";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogOverlay,
    DialogTitle,
} from "@/components/ui/dialog";
import { Slider } from "@/components/ui/slider";
import { Terminal } from "@/components/ui/terminal";
import { JSONObject } from "@/json";
import { cn } from "@/lib/utils";
import { useEffect, useRef, useState } from "react";
import { VariablesDialog } from "./variable-input";

interface WorkflowEvent {
  type: string;
  message: string;
  rows: JSONObject[];
}

interface TableButton {
  text: string;
  enabled: boolean;
  clickState: string;
  icon: string;
  color: string;
}

export default function WorkflowExecutionDialog({
  workflow,
  open,
  onOpenChange,
}: {
  workflow: Workflow | undefined;
  open: boolean;
  onOpenChange: (v: boolean) => void;
}) {
  const [currentStep, setCurrentStep] = useState(-1);
  const [failed, setFailed] = useState(false);
  const [running, setRunning] = useState(false);
  const [events, setEvents] = useState<WorkflowEvent[]>([]);
  const terminalRef = useRef<HTMLDivElement>(null);
  const abortControllerRef = useRef(new AbortController());
  const [button, setButton] = useState<TableButton>({
    text: "Start",
    enabled: true,
    clickState: "start",
    icon: "play_circle",
    color: "bg-green-600",
  });
  const [model, setModel] = useState("");
  const [imageModel, setImageModel] = useState("");
  const [temperature, setTemperature] = useState(0.6);
  const [varDialogOpen, setVarDialogOpen] = useState(false);

  // Auto-scroll to bottom when new content is added
  useEffect(() => {
    scroll();
  }, [events]);

  const scroll = async () => {
    if (terminalRef.current) {
      await new Promise((resolve) => setTimeout(resolve, 50));
      terminalRef.current.scrollTo({
        top: terminalRef.current.scrollHeight,
        behavior: "smooth",
      });
    }
  };

  if (!open || workflow === undefined) {
    return null;
  }

  const handleWorkflowRunEvent = (data: string): void => {
    if (data === "[DONE]") {
      setRunning(false);
      setButton({
        text: "Start",
        enabled: true,
        clickState: "start",
        icon: "play_circle",
        color: "bg-green-600",
      });
      return;
    }
    try {
      const msg: JSONObject = JSON.parse(data);
      switch (msg.type as string) {
        case "MESSAGE":
          setEvents((old) => [
            ...old,
            { type: "Message", message: msg.data as string, rows: [] },
          ]);
          break;
        case "ROWS":
          setEvents((old) => [
            ...old,
            { type: "Rows", message: "", rows: msg.data as JSONObject[] },
          ]);
          break;
        case "STEP_DONE":
          setCurrentStep((old) => old + 1);
          setEvents((old) => [
            ...old,
            { type: "NextStep", message: "", rows: [] },
          ]);
          break;
        case "WORKFLOW_DONE":
          setEvents((old) => [
            ...old,
            {
              type: "Message",
              message: "Workflow completed successfully!",
              rows: [],
            },
          ]);
          break;
        case "ERROR":
          setEvents((old) => [
            ...old,
            { type: "Error", message: msg.data as string, rows: [] },
          ]);
          setRunning(false);
          setFailed(true);
          setButton({
            text: "Start",
            enabled: true,
            clickState: "start",
            icon: "play_circle",
            color: "bg-green-600",
          });
          break;
      }
    } catch (error) {
      console.error("Error generating data:", error);
    }
  };

  const clickButton = (state: string) => {
    switch (state) {
      case "start": {
        // ask user input variables first
        if (workflow.variables.length > 0) {
          setVarDialogOpen(true);
          return;
        }
        setButton({
          text: "Stop",
          enabled: true,
          clickState: "stop",
          icon: "stop_circle",
          color: "bg-red-600",
        });
        setCurrentStep((old) => old + 1);
        setRunning(true);
        abortControllerRef.current = new AbortController();
        runWorkflow(
          workflow.id,
          abortControllerRef.current.signal,
          handleWorkflowRunEvent,
          temperature,
          model,
          imageModel,
          {},
        );
        break;
      }
      case "stop": {
        abortControllerRef.current.abort();
        break;
      }
    }
  };

  const reset = () => {
    setCurrentStep(-1);
    setFailed(false);
    setRunning(false);
    setEvents([]);
  };

  return (
    <Dialog
      open={open}
      onOpenChange={() => {
        reset();
      }}
    >
      <DialogOverlay />
      <DialogContent
        className="min-w-[80vw] h-[85vh] flex flex-column [&>button:last-child]:hidden"
        onInteractOutside={(e) => {
          e.preventDefault();
        }}
      >
        <VariablesDialog
          open={varDialogOpen}
          onOpenChange={setVarDialogOpen}
          variables={workflow.variables}
          onSave={(v) => {
            setButton({
              text: "Stop",
              enabled: true,
              clickState: "stop",
              icon: "stop_circle",
              color: "bg-red-600",
            });
            setVarDialogOpen(false);
            setCurrentStep((old) => old + 1);
            setRunning(true);
            abortControllerRef.current = new AbortController();
            runWorkflow(
              workflow.id,
              abortControllerRef.current.signal,
              handleWorkflowRunEvent,
              temperature,
              model,
              imageModel,
              v,
            );
          }}
        />

        <DialogTitle className="px-4 text-2xl tracking-wider">
          {workflow.name}
        </DialogTitle>

        <div className="flex px-4 items-center">
          <Button
            className={cn("mr-3 text-white rounded-sm", button.color)}
            onClick={() => {
              clickButton(button.clickState);
            }}
            disabled={!button.enabled || (model === "" && imageModel === "")}
          >
            <div className="flex pr-2 justify-center">
              <span className="cursor-pointer material-symbols-rounded">
                {button.icon}
              </span>
            </div>
            {button.text}
          </Button>

          <div>
            <ModelSelector
              hasImageColumn={true}
              generating={running}
              selectModel={(v) => {
                setModel(v);
              }}
              selectImageModel={(v) => {
                setImageModel(v);
              }}
            />
          </div>

          <div className="flex ml-4">Temperature</div>
          <Slider
            min={0}
            value={[temperature]}
            max={2}
            step={0.01}
            className="w-36 ml-4"
            onValueChange={(v: Array<number>) => {
              setTemperature(Number(v[0]));
            }}
          />
          <div className="w-[25px]">
            {temperature.toLocaleString("en-us", { minimumFractionDigits: 2 })}
          </div>
        </div>

        <div className="flex flex-col grow h-[calc(100%-150px)]">
          <div className="flex flex-row md:flex-row gap-6 h-full">
            {/* Left column - Workflow Steps */}
            <div className="w-[500px] rounded-lg shadow-sm overflow-auto">
              <WorkflowSteps
                steps={workflow.steps}
                currentStep={currentStep}
                failed={failed}
              />
            </div>

            {/* Right column - Terminal */}
            <div className="w-full bg-gray-900 h-full rounded-lg shadow-sm overflow-auto border-amber-100/50 border-2 border-solid">
              <Terminal ref={terminalRef} running={running}>
                {events.map((event, index) => (
                  <div key={index} className="mb-3">
                    {event.type === "Message" && (
                      <div className="text-cyan-400">{event.message}</div>
                    )}
                    {event.type === "Error" && (
                      <div className="text-red-400">{event.message}</div>
                    )}
                    {event.type === "Rows" &&
                      event.rows.map((row, i) => (
                        <DataCard data={row} key={500 * index + i} />
                      ))}
                  </div>
                ))}
                {events.length === 0 && (
                  <div className="text-gray-500">
                    Waiting for workflow to start...
                  </div>
                )}
              </Terminal>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button
            type="submit"
            disabled={running}
            onClick={() => {
              reset();
              onOpenChange(false);
            }}
          >
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
