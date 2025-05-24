import {
    AutofillStepPayload,
    CreateColumnStepPayload,
    CreateTableStepPayload,
    createWorkflow,
    DeleteColumnStepPayload,
    DeleteTableStepPayload,
    ExportStepPayload,
    GenerateStepPayload,
    getTables,
    ImportDataStepPayload,
    TableInfo,
} from "@/actions";
import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import WorkflowBuilderDialog from "./builder";

vi.mock(import("@/components/ui/var-input"), () => ({
  MentionInput: (prop) => {
    return (
      <input
        id={prop.id}
        placeholder={prop.placeholder ?? "Type @ to mention a variable..."}
        value={prop.value}
        onChange={prop.onChange}
      />
    );
  },
}));

describe("Workflow Builder", () => {
  beforeEach(async () => {
    const recipeTable = {
      id: "abd",
      name: "recipes",
      description: "recipes table new",
      columns: [
        {
          id: "col1",
          name: "name",
          description: "recipe name",
          type: "string",
          fill_mode: "ai",
        },
        {
          id: "col2",
          name: "description",
          description: "recipe description",
          type: "string",
          fill_mode: "ai",
        },
      ],
      model: "",
    } as TableInfo;
    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({
      tables: [recipeTable],
      total: 1,
    });
    vi.mock("@/actions");
    render(
      <TestProvider>
        <WorkflowBuilderDialog
          open={true}
          onOpenChange={() => {}}
          onSave={async () => {}}
        />
      </TestProvider>,
    );
    await screen.findByText("Workflow Steps");
  });

  it("UserInput should add variables", async () => {
    await screen.findAllByText("UserInput");
    await userEvent.click(screen.getAllByText("UserInput")[0]);
    await userEvent.click(screen.getByText("Add Variable"));
    expect(screen.getByText("Variable 1")).toBeInTheDocument();
    await userEvent.type(screen.getByPlaceholderText("e.g. tableName"), "var1");
    await userEvent.click(
      screen.getByText("String").parentElement as HTMLElement,
    );
    await userEvent.click(screen.getByText("Number"));
    await userEvent.type(screen.getByPlaceholderText("e.g. 3.14"), "1.23");

    await userEvent.click(screen.getByText("Add Variable"));
    expect(screen.getByText("Variable 2")).toBeInTheDocument();
    await userEvent.type(
      screen.getAllByPlaceholderText("e.g. tableName")[1],
      "var2",
    );
    await userEvent.click(
      screen.getByText("String").parentElement as HTMLElement,
    );
    await userEvent.click(screen.getByText("Integer"));
    await userEvent.type(screen.getByPlaceholderText("e.g. 20"), "123");
    await userEvent.type(
      screen.getAllByPlaceholderText(
        "Enter options, one per line, will select one option when workflow start instead input value manually.",
      )[1],
      "foobar",
    );
    await userEvent.click(screen.getByText("Save Workflow"));
    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables.length).toBe(2);
    expect(wf.variables[0]).toStrictEqual({
      default_value: "1.23",
      name: "var1",
      options: [],
      type: "number",
    });
    expect(wf.variables[1]).toStrictEqual({
      default_value: "123",
      name: "var2",
      options: ["foobar"],
      type: "integer",
    });
  });

  it("UserInput file variable should has no default value and option", async () => {
    await screen.findAllByText("UserInput");
    await userEvent.click(screen.getAllByText("UserInput")[0]);
    await userEvent.click(screen.getByText("Add Variable"));
    expect(screen.getByText("Variable 1")).toBeInTheDocument();
    expect(screen.getByText("Default Value")).toBeInTheDocument();
    expect(
      screen.getByText("Options (optional, one per line)"),
    ).toBeInTheDocument();
    await userEvent.type(screen.getByPlaceholderText("e.g. tableName"), "file");
    await userEvent.click(
      screen.getByText("String").parentElement as HTMLElement,
    );
    await userEvent.click(screen.getByText("File"));
    expect(screen.queryByText("Default Value")).toBeNull();
    expect(screen.queryByText("Options (optional, one per line)")).toBeNull();
  });

  it("CreateTable action should save with default name/description and default on_exists", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Create Table"));
    expect(screen.getByText("On Exists")).toBeInTheDocument();
    const onExistsSelectTrigger = screen.getByRole("combobox", {
      name: /on exists/i,
    });
    await userEvent.click(onExistsSelectTrigger);
    const listbox = await screen.findByRole("listbox");
    await userEvent.click(within(listbox).getByText("Stop workflow"));
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("CreateTable");
    const payload = step.payload as CreateTableStepPayload;
    expect(payload.request.name).toBe("");
    expect(payload.request.description).toBe("");
    expect(payload.on_exists).toBe("Stop");
  });

  it("DeleteTable action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Delete Table"));
    await screen.findByText("Table Name");
    await userEvent.click(screen.getByText("Select a table").parentElement!);
    const tableOptionsListbox = await screen.findByRole("listbox");
    await userEvent.click(within(tableOptionsListbox).getByText("recipes"));
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("DeleteTable");
    const payload = step.payload as DeleteTableStepPayload;
    expect(payload.table).toBe("abd");
  });

  it("CreateColumn action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Create Column"));
    await userEvent.click(screen.getByText("Select a table").parentElement!);
    const tableListbox = await screen.findByRole("listbox");
    await userEvent.click(within(tableListbox).getByText("recipes"));
    const columnNameInput = screen.getAllByPlaceholderText(
      "Type @ to mention a variable...",
    )[0];
    await userEvent.type(columnNameInput, "NewColumn");
    const columnDescInput = screen.getAllByPlaceholderText(
      "Type @ to mention a variable...",
    )[1];
    await userEvent.type(columnDescInput, "A new test column");
    const dataTypeSelect = screen.getByRole("combobox", { name: /data type/i });
    await userEvent.click(dataTypeSelect);
    const dataTypeListbox = await screen.findByRole("listbox");
    await userEvent.click(within(dataTypeListbox).getByText("String"));
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("CreateColumn");
    const payload = step.payload as CreateColumnStepPayload;
    expect(payload.table).toBe("abd");
    expect(payload.name).toBe("NewColumn");
    expect(payload.description).toBe("A new test column");
    expect(payload.type).toBe("string");
  });

  it("DeleteColumn action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Delete Column"));
    await userEvent.click(screen.getByText("Select a table").parentElement!);
    const tableListbox = await screen.findByRole("listbox");
    await userEvent.click(within(tableListbox).getByText("recipes"));
    await userEvent.click(screen.getByText("Select a column").parentElement!);
    const columnListbox = await screen.findByRole("listbox");
    await userEvent.click(within(columnListbox).getByText("name"));
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("DeleteColumn");
    const payload = step.payload as DeleteColumnStepPayload;
    expect(payload.table).toBe("abd");
    expect(payload.column).toBe("col1");
  });

  it("Import action should correctly save workflow with new table", async () => {
    await screen.findAllByText("UserInput");
    await userEvent.click(screen.getAllByText("UserInput")[0]);
    await userEvent.click(screen.getByText("Add Variable"));
    await userEvent.type(
      screen.getByPlaceholderText("e.g. tableName"),
      "importFileVar",
    );
    await userEvent.click(
      screen.getByText("String").parentElement as HTMLElement,
    );
    await userEvent.click(screen.getByText("File"));
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Import Data"));
    const fileVarSelect = screen.getByRole("combobox", { name: /file/i });
    await userEvent.click(fileVarSelect);
    const fileVarListbox = await screen.findByRole("listbox");
    await userEvent.click(within(fileVarListbox).getByText("importFileVar"));
    // Updated placeholder for New Table Name in ImportStep
    const input = screen.getByPlaceholderText(
      "Enter new table name or use @ for variables",
    );
    await userEvent.type(input, "ImportedTable");
    const promptInput = screen.getByPlaceholderText(
      "Used only when importing images, as AI is required to extract data from them.",
    );
    await userEvent.type(promptInput, "Importing data");
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables).toHaveLength(1);
    expect(wf.variables[0].name).toBe("importFileVar");
    expect(wf.variables[0].type).toBe("file");
    expect(wf.steps).toHaveLength(1);
    const importStep = wf.steps[0];
    expect(importStep.type).toBe("Import");
    const payload = importStep.payload as ImportDataStepPayload;
    expect(payload.file).toBe("importFileVar");
    expect(payload.name).toBe("ImportedTable");
    expect(payload.table).toBe("");
    expect(payload.truncate).toBe(false);
    expect(payload.prompt).toBe("Importing data");

    // next step should be able to select the table created
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Delete Column"));
    await userEvent.click(screen.getByText("Select a table").parentElement!);
    const tableListbox = await screen.findByRole("listbox");
    const options = within(tableListbox).getAllByRole("option");
    expect(options.length).toBe(2);
    await userEvent.click(within(tableListbox).getByText("ImportedTable"));
  });

  it("Import action should correctly save workflow with selected table", async () => {
    await screen.findAllByText("UserInput");
    await userEvent.click(screen.getAllByText("UserInput")[0]);
    await userEvent.click(screen.getByText("Add Variable"));
    await userEvent.type(
      screen.getByPlaceholderText("e.g. tableName"),
      "importFileVar",
    );
    await userEvent.click(
      screen.getByText("String").parentElement as HTMLElement,
    );
    await userEvent.click(screen.getByText("File"));
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Import Data"));
    const fileVarSelect = screen.getByRole("combobox", { name: /file/i });
    await userEvent.click(fileVarSelect);
    const fileVarListbox = await screen.findByRole("listbox");
    await userEvent.click(within(fileVarListbox).getByText("importFileVar"));
    await userEvent.click(screen.getByText("Create new table").parentElement!);
    const tableOptionsListbox = await screen.findByRole("listbox");
    await userEvent.click(
      within(tableOptionsListbox).getByText("Import into existing table"),
    );
    await userEvent.click(
      within(screen.getByText("Select Table").parentElement!).getByRole(
        "combobox",
      ),
    );
    const tablesBox = await screen.findByRole("listbox");
    await userEvent.click(within(tablesBox).getByText("recipes"));
    const promptInput = screen.getByPlaceholderText(
      "Used only when importing images, as AI is required to extract data from them.",
    );
    await userEvent.type(promptInput, "Importing data");
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables).toHaveLength(1);
    expect(wf.variables[0].name).toBe("importFileVar");
    expect(wf.variables[0].type).toBe("file");
    expect(wf.steps).toHaveLength(1);
    const importStep = wf.steps[0];
    expect(importStep.type).toBe("Import");
    const payload = importStep.payload as ImportDataStepPayload;
    expect(payload.file).toBe("importFileVar");
    expect(payload.name).toBe("");
    expect(payload.table).toBe("abd");
    expect(payload.truncate).toBe(false);
    expect(payload.prompt).toBe("Importing data");

    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Delete Column"));
    await userEvent.click(screen.getByText("Select a table").parentElement!);
    const tableListbox = await screen.findByRole("listbox");
    const options = within(tableListbox).getAllByRole("option");
    expect(options.length).toBe(1);
    await userEvent.click(within(tableListbox).getByText("recipes"));
  });

  it("ExportTable action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Export Data"));
    // Updated placeholder text for ExportTableStep
    await userEvent.click(
      screen.getByText("Select a table to export").parentElement!,
    );
    const tablesBox = await screen.findByRole("listbox");
    await userEvent.click(within(tablesBox).getByText("recipes"));
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("ExportTable");
    const payload = step.payload as ExportStepPayload;
    expect(payload.table).toBe("abd");
  });

  it("Generate action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Generate"));
    await userEvent.click(screen.getByText("Select a table").parentElement!);
    const tablesBox = await screen.findByRole("listbox");
    await userEvent.click(within(tablesBox).getByText("recipes"));
    const countInput = screen.getByLabelText("Count");
    await userEvent.clear(countInput);
    await userEvent.type(countInput, "50");
    const batchInput = screen.getByLabelText("Batch");
    await userEvent.clear(batchInput);
    await userEvent.type(batchInput, "10");
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("Generate");
    const payload = step.payload as GenerateStepPayload;
    expect(payload.table).toBe("abd");
    expect(payload.count).toBe(50);
    expect(payload.batch).toBe(10);
  });

  it("Autofill action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Autofill"));
    await userEvent.click(screen.getByText("Select a table").parentElement!);
    const tablesBox = await screen.findByRole("listbox");
    await userEvent.click(within(tablesBox).getByText("recipes"));
    await userEvent.click(screen.getByText("Select columns..."));
    const columnsListbox = await screen.findByRole("listbox");
    await userEvent.click(within(columnsListbox).getByText("name"));
    await userEvent.keyboard("{Escape}");
    await userEvent.type(
      screen.getByLabelText("Prompt"),
      "Autofill recipe names",
    );
    const countInput = screen.getByLabelText("Count");
    await userEvent.clear(countInput);
    await userEvent.type(countInput, "30");
    const batchInput = screen.getByLabelText("Batch");
    await userEvent.clear(batchInput);
    await userEvent.type(batchInput, "3");
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("Autofill");
    const payload = step.payload as AutofillStepPayload;
    expect(payload.table).toBe("abd");
    expect(payload.columns).toEqual(["col1"]);
    expect(payload.context_columns).toEqual([]);
    expect(payload.prompt).toBe("Autofill recipe names");
    expect(payload.count).toBe(30);
    expect(payload.batch).toBe(3);
  });
});

describe("Workflow Builder - UserInput Variable Name Validation", () => {
  beforeEach(async () => {
    // Common setup for these validation tests
    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({ tables: [], total: 0 }); // No tables needed for these tests
    vi.mock("@/actions"); // Reset mocks

    render(
      <TestProvider>
        <WorkflowBuilderDialog
          open={true}
          onOpenChange={() => {}}
          onSave={async () => {}}
        />
      </TestProvider>,
    );
    await screen.findByText("Workflow Steps");
    await within(
      screen.getByText("Workflow Steps").parentElement!.parentElement!,
    ).findByText("UserInput");

    // Select UserInput step and add a variable
    await userEvent.click(screen.getAllByText("UserInput")[0]);
    await userEvent.click(screen.getByText("Add Variable"));
    await screen.findByText("Variable 1");
  });

  it("should allow valid alphanumeric name", async () => {
    const variableNameInput = screen.getByPlaceholderText("e.g. tableName");
    await userEvent.type(variableNameInput, "TestVar123");
    await userEvent.click(screen.getByText("Save Workflow"));

    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables[0].name).toBe("TestVar123");
  });

  it("should prevent spaces in name", async () => {
    const variableNameInput = screen.getByPlaceholderText("e.g. tableName");
    await userEvent.type(variableNameInput, "Test Var 123"); // Type with spaces
    await userEvent.click(screen.getByText("Save Workflow"));

    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables[0].name).toBe("TestVar123"); // Spaces should be ignored
  });

  it("should prevent special characters in name", async () => {
    const variableNameInput = screen.getByPlaceholderText("e.g. tableName");
    await userEvent.type(variableNameInput, "Test@Var!"); // Type with special chars
    await userEvent.click(screen.getByText("Save Workflow"));

    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables[0].name).toBe("TestVar"); // Special chars should be ignored
  });

  it("should handle mixed valid and invalid characters", async () => {
    const variableNameInput = screen.getByPlaceholderText("e.g. tableName");
    await userEvent.type(variableNameInput, "Alpha1-Beta2="); // Mixed
    await userEvent.click(screen.getByText("Save Workflow"));

    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables[0].name).toBe("Alpha1Beta2"); // Invalid chars ignored
  });

  it("should allow empty name after typing and deleting", async () => {
    const variableNameInput = screen.getByPlaceholderText("e.g. tableName");
    await userEvent.type(variableNameInput, "abc");
    await userEvent.clear(variableNameInput);
    await userEvent.click(screen.getByText("Save Workflow"));

    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables[0].name).toBe("");
  });

  it("should filter leading and trailing invalid characters", async () => {
    const variableNameInput = screen.getByPlaceholderText("e.g. tableName");
    await userEvent.type(variableNameInput, "!@#Valid$%\^");
    await userEvent.click(screen.getByText("Save Workflow"));

    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables[0].name).toBe("Valid");
  });

  it("should result in an empty name if only invalid characters are typed", async () => {
    const variableNameInput = screen.getByPlaceholderText("e.g. tableName");
    await userEvent.type(variableNameInput, "!@#$%^");
    await userEvent.click(screen.getByText("Save Workflow"));

    const wf = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(wf.variables[0].name).toBe("");
  });
});
