import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useLocation, useNavigate } from "react-router-dom";
import { vi, Mock } from "vitest";
import {
  WorkflowInfo,
  Workflow,
  getWorkflows,
  getWorkflow,
  deleteWorkflow,
} from "@/actions";
import { WorkflowListPage } from "./workflow-list-page";

// Mock actions directly
vi.mock("@/actions");
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useNavigate: vi.fn(),
    useLocation: vi.fn(),
  };
});

// Mock for WorkflowBuilderDialog and WorkflowExecutionDialog
vi.mock("@/components/dialog/workflow/builder", () => ({
  __esModule: true,
  default: ({ open, workflow, onOpenChange }: { open: boolean, workflow?: Workflow, onOpenChange: (open: boolean) => void }) =>
    open ? <div data-testid="workflow-builder-dialog">{workflow ? `Edit: ${workflow.name}` : "Create New Workflow"} <button onClick={() => onOpenChange(false)}>Close Builder</button></div> : null,
}));

vi.mock("@/components/dialog/workflow/workflow", () => ({
  __esModule: true,
  default: ({ open, workflow, onOpenChange }: { open: boolean, workflow?: Workflow, onOpenChange: (open: boolean) => void }) =>
    open && workflow ? <div data-testid="workflow-execution-dialog">{`Run: ${workflow.name}`} <button onClick={() => onOpenChange(false)}>Close Runner</button></div> : null,
}));

const mockNavigate = vi.fn();
const mockUseLocation = vi.mocked(useLocation);
const mockUseNavigate = vi.mocked(useNavigate);

// Declare mocked actions for typing
let mockedGetWorkflows: Mock<[], Promise<{ workflows: WorkflowInfo[]; total: number }>>;
let mockedGetWorkflow: Mock<[string], Promise<Workflow>>;
let mockedDeleteWorkflow: Mock<[string], Promise<number>>;

const sampleWorkflows: WorkflowInfo[] = [
  { id: "wf1", name: "Workflow One", description: "Description for one" },
  { id: "wf2", name: "Workflow Two", description: "Description for two" },
];

const sampleWorkflowDetail: Workflow = {
  id: "wf1",
  name: "Workflow One",
  description: "Detailed description for one",
  steps: [],
  variables: [],
  model_config: {},
};

describe("WorkflowListPage", () => {
  beforeEach(async () => {
    mockUseNavigate.mockReturnValue(mockNavigate);
    mockUseLocation.mockReturnValue({
      key: "testKey",
      pathname: "/workflows",
      search: "",
      hash: "",
      state: null,
    });

    // Initialize mocks from @/actions
    mockedGetWorkflows = vi.mocked(getWorkflows);
    mockedGetWorkflow = vi.mocked(getWorkflow);
    mockedDeleteWorkflow = vi.mocked(deleteWorkflow);

    mockedGetWorkflows.mockResolvedValue({
      workflows: sampleWorkflows,
      total: sampleWorkflows.length,
    });
    mockedGetWorkflow.mockResolvedValue(sampleWorkflowDetail);
    mockedDeleteWorkflow.mockResolvedValue(1); // Simulate successful deletion

    render(
      <TestProvider>
        <WorkflowListPage />
      </TestProvider>
    );
    await screen.findByText("Workflow One"); // Wait for initial data to load
  });

  it("should display a list of workflows", () => {
    expect(screen.getByText("Workflow One")).toBeInTheDocument();
    expect(screen.getByText("Description for one")).toBeInTheDocument();
    expect(screen.getByText("Workflow Two")).toBeInTheDocument();
    expect(screen.getByText("Description for two")).toBeInTheDocument();
  });

  it("should open workflow execution dialog when a workflow card is clicked", async () => {
    await userEvent.click(screen.getByText("Workflow One"));
    expect(mockedGetWorkflow).toHaveBeenCalledWith("wf1");
    await screen.findByTestId("workflow-execution-dialog");
    expect(screen.getByText("Run: Workflow One")).toBeInTheDocument();
  });

  it("should open workflow builder dialog when 'Add New Workflow' is clicked", async () => {
    await userEvent.click(screen.getByText("Add New Workflow"));
    await screen.findByTestId("workflow-builder-dialog");
    expect(screen.getByText("Create New Workflow")).toBeInTheDocument();
  });

  it("should open workflow builder dialog for editing when settings icon is clicked", async () => {
    // Find the card for "Workflow One"
    const workflowOneCard = screen.getByText("Workflow One").closest('div[class*="cursor-pointer"]');
    if (!workflowOneCard) throw new Error("Workflow card not found for 'Workflow One'");

    // The settings button is a Button component wrapping an SVG. It doesn't have a textual name.
    // We find all buttons within the card and assume the first one that isn't "Delete" is settings.
    // A more robust way would be a data-testid on the settings button in the component.
    const buttonsInCard = await within(workflowOneCard).findAllByRole('button');
    const settingsButton = buttonsInCard.find(button => !button.textContent?.includes("Delete"));

    if (!settingsButton) throw new Error("Settings button not found for Workflow One");
    
    await userEvent.click(settingsButton);
    
    expect(mockedGetWorkflow).toHaveBeenCalledWith("wf1");
    await screen.findByTestId("workflow-builder-dialog");
    expect(screen.getByText("Edit: Workflow One")).toBeInTheDocument();
  });

  it("should call deleteWorkflow and refresh list when delete button is clicked", async () => {
    mockedGetWorkflows.mockClear(); // Clear previous calls from beforeEach

    // Find the card for "Workflow One"
    const workflowOneCard = screen.getByText("Workflow One").closest('div[class*="cursor-pointer"]');
    if (!workflowOneCard) throw new Error("Workflow card not found for 'Workflow One'");
    
    const deleteButton = within(workflowOneCard).getByRole('button', { name: /delete/i });
    await userEvent.click(deleteButton);

    expect(mockedDeleteWorkflow).toHaveBeenCalledWith("wf1");
    
    // Check if getWorkflows was called again after delete (due to refreshWorkflows)
    await waitFor(() => {
      expect(mockedGetWorkflows).toHaveBeenCalledTimes(1);
    });
  });
});
