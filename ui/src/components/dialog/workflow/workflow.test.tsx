import { getModels, getProviders, runWorkflow } from "@/actions";
import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen, within, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
// Removed act import
import WorkflowExecutionDialog from "./workflow";

describe("Workflow Run", () => {
  beforeEach(async () => {
    vi.mock("@/actions");
    const mockedGetModels = vi.mocked(getModels);
    mockedGetModels.mockResolvedValue({
      default_model: "gpt-3.5-turbo",
      default_image_model: "dall-e-3",
      models: [
        { name: "gpt-3.5-turbo", image: false },
        { name: "dall-e-3", image: true },
      ],
    });
    const mockedGetProviders = vi.mocked(getProviders);
    mockedGetProviders.mockResolvedValue([]);

    // Removed act wrapper around render
    render(
      <TestProvider>
        <WorkflowExecutionDialog
          workflow={{
          id: "w1",
            name: "foo",
            description: "bar",
            variables: [
              { name: "v1", type: "string", default_value: "a", options: [] },
              {
                name: "v2",
                type: "string",
                default_value: "b",
                options: ["b", "c"],
              },
            ],
            steps: [
              { type: "DeleteColumn", payload: { table: "t1", column: "c1" } },
              {
                type: "Generate",
                payload: { table: "t1", count: 1, batch: 1 },
              },
              { type: "DeleteTable", payload: { table: "t1" } },
            ],
          }}
          open={true}
          onOpenChange={() => {}}
        />
      </TestProvider>,
    );
    // Ensure any promises like getModels/getProviders resolve and component updates before tests run
    // A simple way is to wait for a known element that appears after these resolve.
    // For example, if the dialog title or first step name appears due to workflow prop.
    // Or, if not easily identifiable, a small flush of promises can sometimes help,
    // though usually findBy* queries in tests handle this.
    // For now, direct render is usually sufficient as findBy* in tests will wait.
  });

  it("should display workflow steps and icons in initial state", () => {
    expect(screen.getByText("DeleteColumn")).toBeInTheDocument();
    expect(screen.getByText("Generate")).toBeInTheDocument();
    expect(screen.getByText("DeleteTable")).toBeInTheDocument();

    const pendingBadges = screen.getAllByText("Pending");
    expect(pendingBadges).toHaveLength(3);

    // Check for Circle icons. Note: This is a simplified check.
    const iconElements = document.querySelectorAll("svg.lucide-circle");
    expect(iconElements.length).toBe(3);
  });

  it("should ask for variables and run workflow", async () => {
    const user = userEvent.setup();
    const mockedRunWorkflow = vi.mocked(runWorkflow);

    // Wait for the main "Start" button to become enabled due to auto-selection of default model.
    let startButton: HTMLElement | null = null; 
    await waitFor(async () => {
      startButton = await screen.findByRole("button", { name: /start/i });
      expect(startButton).not.toBeDisabled();
    });

    if (!startButton) throw new Error("Start button not found after waitFor");
    await user.click(startButton);

    expect(await screen.findByText("Input Variables")).toBeInTheDocument();

    const v1Input = screen.getByDisplayValue("a");
    await user.clear(v1Input);
    await user.type(v1Input, "new_v1");

    const v2SelectTrigger = screen.getByRole('combobox', { name: /v2/i });
    await user.click(v2SelectTrigger); 
    await user.click(await screen.findByText("c")); 

    const variableDialog = screen.getByRole("dialog", { name: "Input Variables"});
    const dialogStartButton = within(variableDialog).getByRole("button", { name: "Start" });
    await user.click(dialogStartButton);

    expect(mockedRunWorkflow).toHaveBeenCalledTimes(1);
    expect(mockedRunWorkflow).toHaveBeenCalledWith(
      "w1", // workflow.id from beforeEach
      expect.anything(), // AbortSignal
      expect.any(Function), // handleWorkflowRunEvent
      0.6, // Default temperature
      "gpt-3.5-turbo", // Selected model
      "dall-e-3", 
      { v1: "new_v1", v2: "c" }, 
    );
  });

  it("should run workflow to completion and update UI", async () => {
    const user = userEvent.setup();
    const workflowSteps = ["DeleteColumn", "Generate", "DeleteTable"];

    const mockedRunWorkflow = vi.mocked(runWorkflow);
    mockedRunWorkflow.mockImplementation(
      async (
        _workflowId: string,
        _signal: AbortSignal,
        eventCallback: (data: string) => void,
        // Unused params omitted for brevity in mock
      ) => {
        // This mock simulates the backend sending events during a workflow run.
        // Delays are added to mimic real-world async behavior and allow UI to update.
        eventCallback(JSON.stringify({ type: "MESSAGE", data: "Workflow starting..." }));
        await new Promise(r => setTimeout(r, 150)); 

        eventCallback(JSON.stringify({ type: "STEP_DONE" })); 
        await new Promise(r => setTimeout(r, 150));

        eventCallback(JSON.stringify({ type: "STEP_DONE" })); 
        await new Promise(r => setTimeout(r, 150));

        eventCallback(JSON.stringify({ type: "STEP_DONE" })); 
        await new Promise(r => setTimeout(r, 150));

        eventCallback(JSON.stringify({ type: "WORKFLOW_DONE" }));
        await new Promise(r => setTimeout(r, 150));
        
        eventCallback("[DONE]");
      },
    );

    let mainStartButton: HTMLElement | null = null;
    await waitFor(async () => {
      mainStartButton = await screen.findByRole("button", { name: /start/i });
      expect(mainStartButton).not.toBeDisabled();
    });
    if (!mainStartButton) throw new Error("Main Start button not found after waitFor");
    await user.click(mainStartButton);

    expect(await screen.findByText("Input Variables")).toBeInTheDocument();
    const v1Input = screen.getByDisplayValue("a"); 
    await user.clear(v1Input);
    await user.type(v1Input, "test_v1");
    const v2SelectTrigger = screen.getByRole('combobox', { name: /v2/i });
    await user.click(v2SelectTrigger);
    await user.click(await screen.findByText("c")); 
    const dialogStartButton = within(screen.getByRole("dialog", { name: "Input Variables"})).getByRole("button", { name: "Start" });
    await user.click(dialogStartButton);
    
    expect(await screen.findByText("Workflow starting...")).toBeInTheDocument();
    expect(await screen.findByText("Workflow completed successfully!", {}, { timeout: 2000 })).toBeInTheDocument();

    expect(await screen.findByRole("button", { name: /stop/i })).toBeInTheDocument();

    // Check initial "Running" state for Step 0. This assertion is known to be flaky due to timing.
    await waitFor(() => {
      const stepElement = screen.getByText(workflowSteps[0]);
      expect(within(stepElement.closest("li")!).getByText("Running")).toBeInTheDocument();
    }, { timeout: 200 }); 

    // Assert step progression
    await waitFor(async () => {
      const step0Row = await screen.findByText(workflowSteps[0]); 
      expect(within(step0Row.closest("li")!).getByText("Completed")).toBeInTheDocument();
      const step1Row = await screen.findByText(workflowSteps[1]);
      expect(within(step1Row.closest("li")!).getByText("Running")).toBeInTheDocument();
    });
    
    await waitFor(async () => {
      const step1Row = await screen.findByText(workflowSteps[1]);
      expect(within(step1Row.closest("li")!).getByText("Completed")).toBeInTheDocument();
      const step2Row = await screen.findByText(workflowSteps[2]);
      expect(within(step2Row.closest("li")!).getByText("Running")).toBeInTheDocument();
    });

    await waitFor(async () => {
      const step2Row = await screen.findByText(workflowSteps[2]);
      expect(within(step2Row.closest("li")!).getByText("Completed")).toBeInTheDocument();
    });

    for (const stepName of workflowSteps) {
      const stepRow = await screen.findByText(stepName);
      expect(within(stepRow.closest("li")!).getByText("Completed")).toBeInTheDocument();
    }
    
    await waitFor(() => {
        const completedIcons = document.querySelectorAll("svg.lucide-circle-check-big"); 
        expect(completedIcons.length).toBe(workflowSteps.length);
    });

    expect(await screen.findByRole("button", { name: /start/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /stop/i })).not.toBeInTheDocument();
  });

  it("should handle workflow error and update UI accordingly", async () => {
    const user = userEvent.setup();
    const workflowSteps = ["DeleteColumn", "Generate", "DeleteTable"];
    const errorMessage = "Generate step failed!";

    const mockedRunWorkflow = vi.mocked(runWorkflow);
    mockedRunWorkflow.mockImplementation(
      async (
        _workflowId: string,
        _signal: AbortSignal,
        eventCallback: (data: string) => void,
        // Unused params omitted
      ) => {
        eventCallback(JSON.stringify({ type: "MESSAGE", data: "Workflow starting..." }));
        await new Promise(r => setTimeout(r, 150)); 

        eventCallback(JSON.stringify({ type: "STEP_DONE" })); 
        await new Promise(r => setTimeout(r, 150)); 

        eventCallback(JSON.stringify({ type: "ERROR", data: errorMessage }));
      },
    );

    let mainStartButton: HTMLElement | null = null;
    await waitFor(async () => {
      mainStartButton = await screen.findByRole("button", { name: /start/i });
      expect(mainStartButton).not.toBeDisabled();
    });
    if (!mainStartButton) throw new Error("Main Start button not found after waitFor");
    await user.click(mainStartButton);

    expect(await screen.findByText("Input Variables")).toBeInTheDocument();
    const v1Input = screen.getByDisplayValue("a");
    await user.clear(v1Input);
    await user.type(v1Input, "error_test_v1");
    const v2SelectTrigger = screen.getByRole('combobox', { name: /v2/i });
    await user.click(v2SelectTrigger);
    await user.click(await screen.findByText("c"));
    const dialogStartButton = within(screen.getByRole("dialog", { name: "Input Variables"})).getByRole("button", { name: "Start" });
    await user.click(dialogStartButton);

    expect(await screen.findByText("Workflow starting...")).toBeInTheDocument();
    expect(await screen.findByText(errorMessage)).toBeInTheDocument();

    await waitFor(async () => {
      const step0Row = await screen.findByText(workflowSteps[0]);
      const step0ListItem = step0Row.closest("li")!;
      expect(within(step0ListItem).getByText("Completed")).toBeInTheDocument();
      expect(step0ListItem.querySelector("svg.text-green-500")).toBeInTheDocument(); 
    });

    await waitFor(async () => {
      const step1Row = await screen.findByText(workflowSteps[1]);
      const step1ListItem = step1Row.closest("li")!;
      expect(within(step1ListItem).getByText("Failed")).toBeInTheDocument();
      expect(step1ListItem.querySelector("svg.text-red-500")).toBeInTheDocument(); 
    });

    await waitFor(async () => {
      const step2Row = await screen.findByText(workflowSteps[2]);
      const step2ListItem = step2Row.closest("li")!;
      expect(within(step2ListItem).getByText("Pending")).toBeInTheDocument();
      expect(step2ListItem.querySelector("svg.text-gray-400")).toBeInTheDocument(); 
    });
    
    expect(await screen.findByRole("button", { name: /start/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /stop/i })).not.toBeInTheDocument();
  });

  it("should abort workflow when Stop button is clicked", async () => {
    const user = userEvent.setup();
    const workflowSteps = ["DeleteColumn", "Generate", "DeleteTable"];
    const userAbortMessage = "Workflow aborted by user.";
    let signalFromMock: AbortSignal | undefined;

    const mockedRunWorkflow = vi.mocked(runWorkflow);
    mockedRunWorkflow.mockImplementation(
      async (
        _workflowId: string,
        signal: AbortSignal, 
        eventCallback: (data: string) => void,
        // Unused params omitted
      ) => {
        signalFromMock = signal; 

        eventCallback(JSON.stringify({ type: "MESSAGE", data: "Workflow starting..." }));
        await new Promise(r => setTimeout(r, 50));

        if (signal.aborted) { eventCallback("[DONE]"); return; }
        eventCallback(JSON.stringify({ type: "STEP_DONE" })); 
        
        await new Promise(r => setTimeout(r, 200)); // Wait for user to click "Stop"

        if (signal.aborted) {
          eventCallback(JSON.stringify({ type: "MESSAGE", data: userAbortMessage }));
          eventCallback("[DONE]"); 
          return;
        }

        // Should not be reached if test clicks stop
        eventCallback(JSON.stringify({ type: "STEP_DONE" })); 
        await new Promise(r => setTimeout(r, 50));
        eventCallback(JSON.stringify({ type: "STEP_DONE" })); 
        await new Promise(r => setTimeout(r, 50));
        eventCallback(JSON.stringify({ type: "WORKFLOW_DONE" }));
        eventCallback("[DONE]");
      },
    );

    let mainStartButton: HTMLElement | null = null;
    await waitFor(async () => {
      mainStartButton = await screen.findByRole("button", { name: /start/i });
      expect(mainStartButton).not.toBeDisabled();
    });
    if (!mainStartButton) throw new Error("Main Start button not found after waitFor");
    await user.click(mainStartButton);

    expect(await screen.findByText("Input Variables")).toBeInTheDocument();
    const dialogStartButton = within(screen.getByRole("dialog", { name: "Input Variables"})).getByRole("button", { name: "Start" });
    await user.click(dialogStartButton); // Use default variable values

    const stopButton = await screen.findByRole("button", { name: /stop/i });

    // Check initial step states before stopping
    await waitFor(async () => {
      const step0Row = await screen.findByText(workflowSteps[0]);
      expect(within(step0Row.closest("li")!).getByText("Completed")).toBeInTheDocument();
      const step1Row = await screen.findByText(workflowSteps[1]);
      expect(within(step1Row.closest("li")!).getByText("Running")).toBeInTheDocument();
    });

    await user.click(stopButton);

    expect(signalFromMock).not.toBeUndefined();
    expect(signalFromMock!.aborted).toBe(true);
    expect(await screen.findByText(userAbortMessage)).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: /start/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /stop/i })).not.toBeInTheDocument();

    // Assert final step states
    const step0Row = await screen.findByText(workflowSteps[0]);
    expect(within(step0Row.closest("li")!).getByText("Completed")).toBeInTheDocument();
    const step1Row = await screen.findByText(workflowSteps[1]);
    expect(within(step1Row.closest("li")!).getByText("Running")).toBeInTheDocument(); 
    expect(within(step1Row.closest("li")!).queryByText("Failed")).not.toBeInTheDocument();
    const step2Row = await screen.findByText(workflowSteps[2]);
    expect(within(step2Row.closest("li")!).getByText("Pending")).toBeInTheDocument();
    
    expect(screen.queryByText("Workflow completed successfully!")).not.toBeInTheDocument();
  });
});
