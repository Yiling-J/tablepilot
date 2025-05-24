import { TypedWorkflowStep } from "@/actions";
import { Badge } from "@/components/ui/badge";
import { CheckCircle, Circle, Loader2, XCircle } from "lucide-react";

type WorkflowStepsProps = {
  steps: TypedWorkflowStep[];
  currentStep: number;
  failed: boolean;
};

export function WorkflowSteps({
  steps,
  currentStep,
  failed,
}: WorkflowStepsProps) {
  const getStatus = (
    index: number,
  ): "Completed" | "Running" | "Pending" | "Failed" => {
    if (index < currentStep) return "Completed";
    if (index === currentStep) return failed ? "Failed" : "Running";
    return "Pending";
  };

  const getIcon = (status: string) => {
    switch (status) {
      case "Completed":
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case "Running":
        return <Loader2 className="h-5 w-5 text-blue-500 animate-spin" />;
      case "Failed":
        return <XCircle className="h-5 w-5 text-red-500" />;
      default:
        return <Circle className="h-5 w-5 text-gray-400" />;
    }
  };

  const getBadgeVariant = (status: string) => {
    if (status === "Failed") return "destructive";
    if (status === "Running") return "default";
    return "secondary";
  };

  return (
    <div className="h-full flex flex-col">
      <div className="px-4 py-1 border-b">
        <h2 className="text-lg font-semibold">Steps</h2>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <ul className="space-y-3">
          {steps.map((step, index) => {
            const status = getStatus(index);
            return (
              <li
                key={index}
                className="flex items-center justify-between p-3 rounded-md border"
              >
                <div className="flex items-center gap-3">
                  {getIcon(status)}
                  <span className="font-medium">{step.type}</span>
                </div>
                <Badge variant={getBadgeVariant(status)}>{status}</Badge>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
