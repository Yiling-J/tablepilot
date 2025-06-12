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
import { vi, Mock } from "vitest"; // Added Mock import
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
    await screen.findByText("Workflow One");
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
    const workflowOneCard = screen
      .getByText("Workflow One")
      .closest('div[class*="h-60"]');
    if (!workflowOneCard)
      throw new Error("Workflow card not found for 'Workflow One'");

    const settingsButton = await within(
      workflowOneCard as HTMLElement,
    ).findByTitle("Edit");

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

    const workflowOneCard = screen
      .getByText("Workflow One")
      .closest('div[class*="h-60"]');
    if (!workflowOneCard)
      throw new Error("Workflow card not found for 'Workflow One'");

    const mockedGetWorkflows = vi.mocked(getWorkflows);
    mockedGetWorkflows.mockReset();
    mockedGetWorkflows.mockResolvedValue({
      workflows: [],
      total: 0,
    });
    const deleteButton = within(workflowOneCard as HTMLElement).getByTitle(
      "Delete",
    );
    await userEvent.click(deleteButton);
    await userEvent.click(screen.getByText("Delete"));

    expect(mockedDeleteWorkflow).toHaveBeenCalledWith("wf1");

    await waitFor(() => {
      expect(mockedGetWorkflows).toHaveBeenCalledTimes(1);
    });
  });
});

describe("WorkflowListPage Search Functionality", () => {
  const searchSampleWorkflows: WorkflowInfo[] = [
    { id: "sw1", name: "Alpha Workflow", description: "Description Alpha" },
    { id: "sw2", name: "Beta Workflow", description: "Description Beta" },
    { id: "sw3", name: "Workflow Gamma", description: "Description Gamma" },
  ];

  beforeEach(async () => {
    vi.resetAllMocks();
    vi.mock("react-router-dom", async (importOriginal) => {
        const actual = await importOriginal();
        return {
            ...(typeof actual === 'object' && actual !== null ? actual : {}),
            useNavigate: vi.fn(() => vi.fn()),
            useLocation: vi.fn(() => ({ pathname: '/workflows', search: '', hash: '', state: null, key: 'testKey' })),
        };
    });
    vi.mock("@/actions");

    (getWorkflows as Mock).mockResolvedValue({
      workflows: searchSampleWorkflows,
      total: searchSampleWorkflows.length,
    });
    (getWorkflow as Mock).mockResolvedValue(sampleWorkflowDetail);
    (deleteWorkflow as Mock).mockResolvedValue({});
    (getModels as Mock).mockResolvedValue({ default_model: "test", models: [] });
    (getProviders as Mock).mockResolvedValue([]);
    (getTables as Mock).mockResolvedValue({ tables: [], total: 0 });


    render(
      <TestProvider>
        <WorkflowListPage />
      </TestProvider>
    );
    await screen.findByText("Alpha Workflow");
    await screen.findByText("Beta Workflow");
    await screen.findByText("Workflow Gamma");
  });

  it("should render all workflows initially", () => {
    expect(screen.getByText("Alpha Workflow")).toBeInTheDocument();
    expect(screen.getByText("Beta Workflow")).toBeInTheDocument();
    expect(screen.getByText("Workflow Gamma")).toBeInTheDocument();
  });

  it("should filter workflows based on search query", async () => {
    const searchInput = screen.getByPlaceholderText("Search workflows...");
    await userEvent.type(searchInput, "Alpha");

    await waitFor(() => {
      expect(screen.getByText("Alpha Workflow")).toBeInTheDocument();
      expect(screen.queryByText("Beta Workflow")).not.toBeInTheDocument();
      expect(screen.queryByText("Workflow Gamma")).not.toBeInTheDocument();
    });
  });

  it("should be case-insensitive", async () => {
    const searchInput = screen.getByPlaceholderText("Search workflows...");
    await userEvent.type(searchInput, "gamma");

    await waitFor(() => {
      expect(screen.queryByText("Alpha Workflow")).not.toBeInTheDocument();
      expect(screen.queryByText("Beta Workflow")).not.toBeInTheDocument();
      expect(screen.getByText("Workflow Gamma")).toBeInTheDocument();
    });
  });

  it("should show 'No workflows found' message if search matches nothing", async () => {
    const searchInput = screen.getByPlaceholderText("Search workflows...");
    await userEvent.type(searchInput, "NonExistentWorkflow");

    await waitFor(() => {
      expect(screen.getByText("No workflows found matching your search.")).toBeInTheDocument();
    });
    expect(screen.queryByText("Alpha Workflow")).not.toBeInTheDocument();
    expect(screen.queryByText("Beta Workflow")).not.toBeInTheDocument();
    expect(screen.queryByText("Workflow Gamma")).not.toBeInTheDocument();
  });

  it("should show all workflows when search query is cleared", async () => {
    const searchInput = screen.getByPlaceholderText("Search workflows...");
    await userEvent.type(searchInput, "Alpha");

    await waitFor(() => expect(screen.getByText("Alpha Workflow")).toBeInTheDocument());
    await waitFor(() => expect(screen.queryByText("Beta Workflow")).not.toBeInTheDocument());

    await userEvent.clear(searchInput);

    await waitFor(() => {
      expect(screen.getByText("Alpha Workflow")).toBeInTheDocument();
      expect(screen.getByText("Beta Workflow")).toBeInTheDocument();
      expect(screen.getByText("Workflow Gamma")).toBeInTheDocument();
    });
  });
});
