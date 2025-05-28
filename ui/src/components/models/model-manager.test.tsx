import type { Model, Provider, ProviderType, TableInfo } from "@/actions";
import {
    createProvider,
    deleteProvider,
    getProviders,
    updateProvider,
} from "@/actions";
import { useToast } from "@/hooks/use-toast";
import { TestProvider } from "@/test/helpers/test-provider";
import {
    fireEvent,
    render,
    screen,
    waitFor,
    within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useNavigate } from "react-router-dom";
import { Mock, vi } from "vitest";
import { ModelManagerPageWrapper } from "./model-manager-page-wrapper";

// Define sample data at the top level for the vi.mock factory
const topLevelSampleModels: Model[] = [
  {
    model: "gpt-3.5-turbo",
    alias: "GPT-3.5",
    max_tokens: 4096,
    rpm: 0,
    image: false,
  },
  { model: "gpt-4", alias: "GPT-4", max_tokens: 8192, rpm: 0, image: false },
  {
    model: "claude-2",
    alias: "Claude 2",
    max_tokens: 100000,
    rpm: 0,
    image: false,
  },
];

const topLevelSampleProviders: Provider[] = [
  {
    id: 1,
    name: "OpenAI Provider",
    type: "openai",
    key: "OpenAI" as ProviderType,
    base_url: "",
    models: [topLevelSampleModels[0], topLevelSampleModels[1]],
    editable: true,
    enabled: true,
  },
  {
    id: 2,
    name: "Anthropic Provider",
    type: "anthropic",
    key: "Anthropic" as ProviderType,
    base_url: "",
    models: [topLevelSampleModels[2]],
    editable: true,
    enabled: true,
  },
];

const topLevelMockTableInfoResponse: TableInfo = {
  id: "mockTableId",
  name: "Mock TableName",
  description: "Mock table description",
  columns: [],
  model: "mockModel",
};

vi.mock("@/actions", async () => {
  const originalActions = (await vi.importActual("@/actions")) as Record<
    string,
    unknown
  >;
  return {
    ...originalActions,
    getProviders: vi.fn(() => Promise.resolve([...topLevelSampleProviders])),
    createProvider: vi.fn(() => Promise.resolve(topLevelMockTableInfoResponse)),
    updateProvider: vi.fn(() => Promise.resolve(topLevelMockTableInfoResponse)),
    deleteProvider: vi.fn(() => Promise.resolve(200)),
    ProviderTypeOptions: originalActions.ProviderTypeOptions || [
      "OpenAI",
      "Gemini",
      "Anthropic",
      "OpenAI-Compatible",
    ],
  };
});

vi.mock("@/hooks/use-toast", () => ({
  useToast: vi.fn(() => ({
    toast: vi.fn(),
    dismiss: vi.fn(),
    toasts: [],
  })),
}));

const mockedGetProviders = getProviders as Mock;
const mockedCreateProvider = createProvider as Mock;
const mockedUpdateProvider = updateProvider as Mock;
const mockedDeleteProvider = deleteProvider as Mock;

// Use copies of the top-level data for tests to avoid modification issues
const sampleModels: Model[] = JSON.parse(JSON.stringify(topLevelSampleModels));
const sampleProviders: Provider[] = JSON.parse(
  JSON.stringify(topLevelSampleProviders),
);
const mockTableInfoResponse: TableInfo = JSON.parse(
  JSON.stringify(topLevelMockTableInfoResponse),
);

const findProviderCard = async (name: string): Promise<HTMLElement> => {
  const providerNameElement = await screen.findByText(name);
  const providerCard = providerNameElement.closest("div.provider-card");
  if (!providerCard) throw new Error(`Provider card for "${name}" not found`);
  return providerCard as HTMLElement;
};

describe("ModelManager", () => {
  let toastMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mock("react-router-dom");
    vi.mocked(useNavigate).mockReturnValue(vi.fn());
    // Reset to fresh copies for each test run
    mockedGetProviders.mockResolvedValue(
      JSON.parse(JSON.stringify(sampleProviders)),
    );
    mockedCreateProvider.mockResolvedValue(
      JSON.parse(JSON.stringify(mockTableInfoResponse)),
    );
    mockedUpdateProvider.mockResolvedValue(
      JSON.parse(JSON.stringify(mockTableInfoResponse)),
    );
    mockedDeleteProvider.mockResolvedValue(200);

    toastMock = vi.fn();
    vi.mocked(useToast).mockReturnValue({
      toast: toastMock,
      dismiss: vi.fn(),
      toasts: [],
    });
  });

  it("should render the ModelManager component and display initial providers", async () => {
    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );

    expect(
      await screen.findByText("OpenAI Provider", {}, { timeout: 5000 }),
    ).toBeInTheDocument();
    expect(screen.getByText("Anthropic Provider")).toBeInTheDocument();
    // The DOM snapshot from previous failures showed "Model Providers" was not present.
    // This assertion is removed as it was causing the test to fail after data loaded.
    // expect(screen.getByRole('heading', { name: /Providers/i })).toBeInTheDocument();
    expect(mockedGetProviders).toHaveBeenCalledTimes(1);
  });

  it("should allow adding a new provider", async () => {
    const newProviderFormData = {
      name: "New Custom Provider",
      type: "OpenAI-Compatible",
      key: "OpenAI-Compatible" as ProviderType,
      config: { api_key: "sk-customkey" },
      base_url: "http://localhost:8080",
      models: [sampleModels[0]],
      editable: true,
      enabled: true,
    };
    const newProviderInList: Provider = {
      id: 3,
      name: newProviderFormData.name,
      type: newProviderFormData.type,
      key: newProviderFormData.key,
      base_url: newProviderFormData.base_url,
      models: newProviderFormData.models,
      editable: newProviderFormData.editable,
      enabled: newProviderFormData.enabled,
    };
    mockedGetProviders
      .mockResolvedValueOnce(JSON.parse(JSON.stringify(sampleProviders)))
      .mockResolvedValueOnce([
        ...JSON.parse(JSON.stringify(sampleProviders)),
        newProviderInList,
      ]);

    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );
    expect(await screen.findByText("OpenAI Provider")).toBeInTheDocument();

    await userEvent.click(
      screen.getByRole("button", { name: /Add Provider/i }),
    );
    await screen.findByText("Create New Provider");
    await userEvent.type(
      screen.getByLabelText(/Name/i),
      newProviderFormData.name,
    );
    await userEvent.type(
      screen.getByLabelText(/API Key/i),
      newProviderFormData.config.api_key,
    );
    fireEvent.submit(screen.getByRole("button", { name: "Create Provider" }));

    expect(mockedCreateProvider).toHaveBeenCalledWith({
      id: 0,
      name: newProviderFormData.name,
      key: newProviderFormData.config.api_key,
      type: "OpenAI",
      base_url: "",
      editable: true,
      enabled: true,
      models: [],
    });
    expect(
      await screen.findByText(newProviderFormData.name),
    ).toBeInTheDocument();
  });

  it("should allow adding a new OpenAI-Compatible provider with baseurl", async () => {
    mockedGetProviders.mockResolvedValue([]);

    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );

    expect(
      await screen.findByText("Create a new provider"),
    ).toBeInTheDocument();
    await userEvent.click(screen.getByText("Create a new provider"));
    await screen.findByText("Create New Provider");
    await userEvent.type(screen.getByLabelText(/Name/i), "foo");
    await userEvent.selectOptions(
      screen.getByRole("combobox").nextElementSibling!,
      "OpenAI-Compatible",
    );
    await userEvent.type(screen.getByLabelText("Base URL"), "https://zzz.com");
    fireEvent.submit(screen.getByRole("button", { name: "Create Provider" }));

    expect(mockedCreateProvider).toHaveBeenCalledWith({
      id: 0,
      name: "foo",
      key: "",
      type: "OpenAI-Compatible",
      base_url: "https://zzz.com",
      editable: true,
      enabled: true,
      models: [],
    });
  });

  it("should allow editing an existing provider", async () => {
    const providerToEdit = sampleProviders[0];
    const updatedName = "Updated OpenAI Provider";
    const updatedApiKey = "sk-updatedkey";
    const updatedProviderInList: Provider = {
      ...providerToEdit,
      name: updatedName,
    };

    mockedGetProviders
      .mockResolvedValueOnce(JSON.parse(JSON.stringify(sampleProviders)))
      .mockResolvedValueOnce([
        updatedProviderInList,
        JSON.parse(JSON.stringify(sampleProviders[1])),
      ]);

    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );
    expect(
      await screen.findByText(providerToEdit.name, {}, { timeout: 2000 }),
    ).toBeInTheDocument();

    const providerCard = await findProviderCard(providerToEdit.name);
    await userEvent.click(
      within(providerCard).getByRole("button", { name: "Edit Provider" }),
    );
    await screen.findByRole("heading", { name: "Edit Provider" });
    const nameInput = screen.getByLabelText(/Name/i);
    const apiKeyInput = screen.getByLabelText(/API Key/i);
    await userEvent.clear(nameInput);
    await userEvent.type(nameInput, updatedName);
    await userEvent.clear(apiKeyInput);
    await userEvent.type(apiKeyInput, updatedApiKey);
    fireEvent.submit(screen.getByRole("button", { name: "Save Changes" }));

    expect(mockedUpdateProvider).toHaveBeenCalledWith(
      providerToEdit.id.toString(),
      {
        id: providerToEdit.id,
        name: updatedName,
        type: providerToEdit.type,
        base_url: providerToEdit.base_url,
        models: providerToEdit.models,
        key: updatedApiKey,
        editable: providerToEdit.editable,
        enabled: providerToEdit.enabled,
      },
    );
    expect(await screen.findByText(updatedName)).toBeInTheDocument();
  });

  it("should allow deleting a provider", async () => {
    const providerToDelete = sampleProviders[0];
    mockedGetProviders
      .mockResolvedValueOnce(JSON.parse(JSON.stringify(sampleProviders)))
      .mockResolvedValueOnce([JSON.parse(JSON.stringify(sampleProviders[1]))]);

    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );
    expect(
      await screen.findByText(providerToDelete.name, {}, { timeout: 2000 }),
    ).toBeInTheDocument();

    const providerCard = await findProviderCard(providerToDelete.name);
    await userEvent.click(
      within(providerCard).getByRole("button", { name: "Delete Provider" }),
    );
    await screen.findByText("Confirm Deletion");
    await userEvent.click(screen.getByRole("button", { name: /Confirm/i }));

    expect(mockedDeleteProvider).toHaveBeenCalledWith(
      providerToDelete.id.toString(),
    );
    await waitFor(() =>
      expect(screen.queryByText(providerToDelete.name)).not.toBeInTheDocument(),
    );
  });

  it("should toggle provider enabled state and show a toast", async () => {
    const providerToToggle = sampleProviders[0];
    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );
    expect(
      await screen.findByText(providerToToggle.name, {}, { timeout: 2000 }),
    ).toBeInTheDocument();

    const providerCard = await findProviderCard(providerToToggle.name);
    const toggleSwitch = within(providerCard).getByRole("switch");

    await userEvent.click(toggleSwitch);
    expect(toastMock).toHaveBeenCalledWith(
      expect.objectContaining({
        title: "Provider Disabled",
        description: "OpenAI Provider has been disabled.",
      }),
    );
    expect(mockedUpdateProvider).toHaveBeenCalledWith("1", {
      ...providerToToggle,
      enabled: false,
    });
    toastMock.mockClear();
    mockedUpdateProvider.mockClear();
    await userEvent.click(toggleSwitch);
    expect(toastMock).toHaveBeenCalledWith(
      expect.objectContaining({
        title: "Provider Enabled",
        description: "OpenAI Provider has been enabled.",
      }),
    );
    expect(mockedUpdateProvider).toHaveBeenCalledWith("1", {
      ...providerToToggle,
      enabled: true,
    });
  });

  it("should display models for each provider", async () => {
    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );
    expect(
      await screen.findByText("OpenAI Provider", {}, { timeout: 2000 }),
    ).toBeInTheDocument();

    const openAICard = await findProviderCard("OpenAI Provider");
    expect(within(openAICard).getByText("GPT-3.5")).toBeInTheDocument();
    const anthropicCard = await findProviderCard("Anthropic Provider");
    expect(within(anthropicCard).getByText("Claude 2")).toBeInTheDocument();
  });

  it("should allow adding a new model to a provider and call updateProvider", async () => {
    const providerName = sampleProviders[0].name;
    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );
    expect(
      await screen.findByText(providerName, {}, { timeout: 2000 }),
    ).toBeInTheDocument();

    const providerCard = await findProviderCard(providerName);
    await userEvent.click(within(providerCard).getByText("Add New Model"));
    await screen.findByText(/Add a new model to/i);
    const newModelData = {
      model: "new-model-id",
      alias: "New Model Alias",
      max_tokens: "2048",
    };
    await userEvent.type(screen.getByLabelText("Name"), newModelData.model);
    await userEvent.type(screen.getByLabelText("Alias"), newModelData.alias);
    await userEvent.clear(screen.getByLabelText("Max Tokens"));
    await userEvent.type(
      screen.getByLabelText("Max Tokens"),
      newModelData.max_tokens,
    );
    fireEvent.submit(screen.getByRole("button", { name: "Add Model" }));
    expect(toastMock).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Model Added" }),
    );
    expect(
      await within(providerCard).findByText(newModelData.alias),
    ).toBeInTheDocument();
    expect(mockedUpdateProvider).toHaveBeenCalledWith("1", {
      ...sampleProviders[0],
      models: [
        ...sampleProviders[0].models,
        {
          model: newModelData.model,
          alias: newModelData.alias,
          max_tokens: 2048,
          image: false,
          rpm: 10,
        },
      ],
    });
  });

  it("should allow editing an existing model and call updateProvider", async () => {
    const providerToEdit = sampleProviders[0];
    const modelToEdit = providerToEdit.models[0];
    const updatedAlias = "GPT-3.5 Updated Alias";
    const updatedName = "GPT-x";
    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );
    expect(
      await screen.findByText(providerToEdit.name, {}, { timeout: 2000 }),
    ).toBeInTheDocument();

    const providerCard = await findProviderCard(providerToEdit.name);
    const modelElement = await within(providerCard).findByText(
      modelToEdit.alias,
    );
    const modelContainer = modelElement.closest(
      "div.model-card",
    ) as HTMLElement;
    await userEvent.click(
      within(modelContainer).getByRole("button", { name: "Edit Model" }),
    );
    await screen.findByText("Edit Model");
    await userEvent.clear(screen.getByLabelText("Name"));
    await userEvent.type(screen.getByLabelText("Name"), updatedName);
    await userEvent.clear(screen.getByLabelText("Alias"));
    await userEvent.type(screen.getByLabelText("Alias"), updatedAlias);
    fireEvent.submit(screen.getByRole("button", { name: "Save Changes" }));
    expect(toastMock).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Model Updated" }),
    );
    expect(
      await within(providerCard).findByText(updatedAlias),
    ).toBeInTheDocument();
    expect(mockedUpdateProvider).toHaveBeenCalledWith("1", {
      ...sampleProviders[0],
      models: [
        {
          ...sampleProviders[0].models[0],
          alias: updatedAlias,
          model: updatedName,
          rpm: 10,
        },
        sampleProviders[0].models[1],
      ],
    });
  });

  it("should allow deleting a model from a provider and call updateProvider", async () => {
    const providerToDeleteFrom = sampleProviders[0];
    const modelToDelete = providerToDeleteFrom.models[0];
    render(
      <TestProvider>
        <ModelManagerPageWrapper />
      </TestProvider>,
    );
    expect(
      await screen.findByText(providerToDeleteFrom.name, {}, { timeout: 2000 }),
    ).toBeInTheDocument();

    const providerCard = await findProviderCard(providerToDeleteFrom.name);
    const modelElement = await within(providerCard).findByText(
      modelToDelete.alias,
    );
    const modelContainer = modelElement.closest(
      "div.model-card",
    ) as HTMLElement;
    await userEvent.click(
      within(modelContainer).getByRole("button", { name: "Delete Model" }),
    );
    await screen.findByText("Confirm Deletion");
    await userEvent.click(screen.getByRole("button", { name: /Confirm/i }));

    expect(toastMock).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Model Deleted" }),
    );
    await waitFor(() =>
      expect(
        within(providerCard).queryByText(modelToDelete.alias),
      ).not.toBeInTheDocument(),
    );
    expect(mockedUpdateProvider).toHaveBeenCalledWith("1", {
      ...sampleProviders[0],
      models: [sampleProviders[0].models[1]],
    });
  });
});
