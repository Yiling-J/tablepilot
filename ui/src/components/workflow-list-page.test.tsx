import {
    Provider,
    Workflow,
    WorkflowInfo,
    deleteWorkflow,
    getModels,
    getProviders,
    getTables,
    getWorkflow,
    getWorkflows,
} from "@/actions";
import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useNavigate } from "react-router-dom";
import { vi } from "vitest";
import { WorkflowListPage } from "./workflow-list-page";

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
};

describe("WorkflowListPage", () => {
  beforeEach(async () => {
    vi.mock("react-router-dom");
    const m = vi.mocked(useNavigate);
    m.mockReturnValue(vi.fn());
    vi.mock("@/actions");
    const mockedGetModels = vi.mocked(getModels);
    mockedGetModels.mockResolvedValue({
      default_model: "ai",
      default_image_model: "",
      models: [
        { name: "ai", image: false },
        { name: "bi", image: false },
        { name: "ci", image: true },
      ],
    });
    const mockedGetProviders = vi.mocked(getProviders);
    mockedGetProviders.mockResolvedValue([
      {
        id: 1,
        name: "p",
        type: "openai",
        models: [{ model: "ai" }],
      } as Provider,
    ]);

    const mockedGetWorkflows = vi.mocked(getWorkflows);
    mockedGetWorkflows.mockResolvedValue({
      workflows: sampleWorkflows,
      total: 2,
    });

    render(
      <TestProvider>
        <WorkflowListPage />
      </TestProvider>,
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
    const mockedGetWorkflow = vi.mocked(getWorkflow);
    mockedGetWorkflow.mockResolvedValue(sampleWorkflowDetail);
    await userEvent.click(screen.getByText("Workflow One"));
    expect(mockedGetWorkflow).toHaveBeenCalledWith("wf1");
    await screen.findByText("Steps");
    const dg = screen.getByRole("dialog");
    expect(within(dg).getByText("Workflow One")).toBeInTheDocument();
  });

  it("should open workflow builder dialog when 'Add New Workflow' is clicked", async () => {
    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({
      tables: [],
      total: 0,
    });
    await userEvent.click(screen.getByText("Add New Workflow"));
    await screen.findByText("New Workflow");
    expect(screen.getByText("New Workflow")).toBeInTheDocument();
  });

  it("should open workflow builder dialog for editing when settings icon is clicked", async () => {
    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({
      tables: [],
      total: 0,
    });
    const mockedGetWorkflow = vi.mocked(getWorkflow);
    mockedGetWorkflow.mockResolvedValue(sampleWorkflowDetail);
    // Find the card for "Workflow One"
    const workflowOneCard = screen
      .getByText("Workflow One")
      .closest('div[class*="cursor-pointer"]');
    if (!workflowOneCard)
      throw new Error("Workflow card not found for 'Workflow One'");

    const buttonsInCard = await within(
      workflowOneCard as HTMLElement,
    ).findAllByRole("button");
    const settingsButton = buttonsInCard.find(
      (button) => !button.textContent?.includes("Delete"),
    );

    if (!settingsButton)
      throw new Error("Settings button not found for Workflow One");

    await userEvent.click(settingsButton);

    expect(mockedGetWorkflow).toHaveBeenCalledWith("wf1");
    await screen.findByText("Workflow Steps");
    const dg = screen.getByRole("dialog");
    expect(within(dg).getByText("Workflow One")).toBeInTheDocument();
  });

  it("should call deleteWorkflow and refresh list when delete button is clicked", async () => {
    const mockedDeleteWorkflow = vi.mocked(deleteWorkflow);
    mockedDeleteWorkflow.mockResolvedValue();

    // Find the card for "Workflow One"
    const workflowOneCard = screen
      .getByText("Workflow One")
      .closest('div[class*="cursor-pointer"]');
    if (!workflowOneCard)
      throw new Error("Workflow card not found for 'Workflow One'");

    const mockedGetWorkflows = vi.mocked(getWorkflows);
    mockedGetWorkflows.mockReset();
    mockedGetWorkflows.mockResolvedValue({
      workflows: [],
      total: 0,
    });
    const deleteButton = within(workflowOneCard as HTMLElement).getByRole(
      "button",
      {
        name: /delete/i,
      },
    );
    await userEvent.click(deleteButton);

    expect(mockedDeleteWorkflow).toHaveBeenCalledWith("wf1");

    // Check if getWorkflows was called again after delete (due to refreshWorkflows)
    await waitFor(() => {
      expect(mockedGetWorkflows).toHaveBeenCalledTimes(1);
    });
  });
});
