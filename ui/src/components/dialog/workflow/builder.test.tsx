import { createWorkflow, getTables, TableInfo } from "@/actions";
import { TestProvider } from "@/test/helpers/test-provider";
import {
  render,
  screen,
  within,
  waitFor, 
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import WorkflowBuilderDialog from "./builder";
import { MemoryRouter } from "react-router-dom";

// const consoleLogSpy = vi.spyOn(console, 'log');

describe("Workflow Builder", () => {
  beforeEach(async () => {
    const recipeTable = { // This is an existing table from "DB"
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
        }
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
        <MemoryRouter>
          <WorkflowBuilderDialog
            open={true}
            onOpenChange={() => {}}
            onSave={async () => {}}
          />
        </MemoryRouter>
      </TestProvider>,
    );
  });

  // afterEach(() => {
  //   consoleLogSpy.mockRestore();
  // });

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
    expect(step.payload.request.name).toBe("");
    expect(step.payload.request.description).toBe("");
    expect(step.payload.on_exists).toBe("Stop");
  });

  it("DeleteTable action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Delete Table"));
    const tableNameSelectTrigger = screen.getByRole("combobox", {
      name: /table name/i, 
    });
    await userEvent.click(tableNameSelectTrigger);
    const tableOptionsListbox = await screen.findByRole("listbox");
    await userEvent.click(within(tableOptionsListbox).getByText("recipes"));
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("DeleteTable");
    expect(step.payload.table).toBe("abd");
  });

  it("CreateColumn action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Create Column"));
    const tableNameSelect = screen.getByRole("combobox", { name: /table name/i });
    await userEvent.click(tableNameSelect);
    const tableListbox = await screen.findByRole("listbox");
    await userEvent.click(within(tableListbox).getByText("recipes"));
    const columnNameInput = screen.getByLabelText("Column Name");
    await userEvent.type(columnNameInput, "NewColumn");
    const columnDescInput = screen.getByLabelText("Column Description");
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
    expect(step.payload.table).toBe("abd");
    expect(step.payload.name).toBe("NewColumn");
    expect(step.payload.description).toBe("A new test column");
    expect(step.payload.type).toBe("string");
  });

  it("DeleteColumn action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Delete Column"));
    const tableNameSelect = screen.getByRole("combobox", { name: /table name/i });
    await userEvent.click(tableNameSelect);
    const tableListbox = await screen.findByRole("listbox");
    await userEvent.click(within(tableListbox).getByText("recipes"));
    const columnSelect = screen.getByRole("combobox", { name: /column/i });
    await userEvent.click(columnSelect);
    const columnListbox = await screen.findByRole("listbox");
    await userEvent.click(within(columnListbox).getByText("name"));
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("DeleteColumn");
    expect(step.payload.table).toBe("abd"); 
    expect(step.payload.column).toBe("col1"); 
  });

  it("Import action should correctly save workflow with new table", async () => {
    await userEvent.click(screen.getByText("Add Variable"));
    await userEvent.type(screen.getByPlaceholderText("e.g. tableName"), "importFileVar");
    await userEvent.click(screen.getByText("String").parentElement as HTMLElement);
    await userEvent.click(screen.getByText("File"));
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Import Data"));
    const fileVarSelect = screen.getByRole("combobox", { name: /file/i });
    await userEvent.click(fileVarSelect);
    const fileVarListbox = await screen.findByRole("listbox");
    await userEvent.click(within(fileVarListbox).getByText("importFileVar"));
    const newTableNameInput = screen.getByLabelText("New Table Name");
    await userEvent.type(newTableNameInput, "ImportedTable");
    const promptInput = screen.getByLabelText("Prompt (Import Image)");
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
    expect(importStep.payload.file).toBe("importFileVar"); 
    expect(importStep.payload.name).toBe("ImportedTable");
    expect(importStep.payload.table).toBe(""); 
    expect(importStep.payload.truncate).toBe(false); 
    expect(importStep.payload.prompt).toBe("Importing data");
  });

  it("ExportTable action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Export Data"));
    const tableNameSelect = screen.getByRole("combobox", { name: /table name/i });
    await userEvent.click(tableNameSelect);
    const tableListbox = await screen.findByRole("listbox");
    await userEvent.click(within(tableListbox).getByText("recipes"));
    await userEvent.click(screen.getByText("Save Workflow"));
    expect(createWorkflow).toHaveBeenCalled();
    const workflow = vi.mocked(createWorkflow).mock.calls[0][0];
    expect(workflow.steps).toHaveLength(1);
    const step = workflow.steps[0];
    expect(step.type).toBe("ExportTable");
    expect(step.payload.table).toBe("abd"); 
  });

  it("Generate action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Generate"));
    const tableNameSelect = screen.getByRole("combobox", { name: /table name/i });
    await userEvent.click(tableNameSelect);
    const tableListbox = await screen.findByRole("listbox");
    await userEvent.click(within(tableListbox).getByText("recipes"));
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
    expect(step.payload.table).toBe("abd");
    expect(step.payload.count).toBe(50);
    expect(step.payload.batch).toBe(10);
  });

  it("Autofill action should correctly save workflow", async () => {
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Autofill"));
    const tableNameSelect = screen.getByRole("combobox", { name: /table name/i });
    await userEvent.click(tableNameSelect);
    const tableListbox = await screen.findByRole("listbox");
    await userEvent.click(within(tableListbox).getByText("recipes"));
    const columnsToAutofillTrigger = screen.getByRole("button", { name: /columns to autofill/i });
    await userEvent.click(columnsToAutofillTrigger);
    const columnsListbox = await screen.findByRole("listbox"); 
    await userEvent.click(within(columnsListbox).getByText("name"));
    const promptInput = screen.getByLabelText(/prompt/i); 
    await userEvent.type(promptInput, "Autofill recipe names");
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
    expect(step.payload.table).toBe("abd");
    expect(step.payload.columns).toEqual(["col1"]);
    expect(step.payload.context_columns).toEqual([]);
    expect(step.payload.prompt).toBe("Autofill recipe names");
    expect(step.payload.count).toBe(30);
    expect(step.payload.batch).toBe(3);
  });

  it("should correctly save a complex multi-step workflow", async () => {
    // Initial UserInput Step
    await userEvent.click(screen.getByText("Add Variable"));
    await userEvent.type(screen.getByPlaceholderText("e.g. tableName"), "inputFile1");
    await userEvent.click(screen.getByText("String").parentElement as HTMLElement);
    await userEvent.click(screen.getByText("File"));

    // Step 1: CreateTable
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Create Table"));
    // "On Exists" is "Stop" by default, no interaction needed to verify.

    // Step 2: CreateColumn
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Create Column"));
    let tableNameSelect = screen.getByRole("combobox", { name: /table name/i });
    await userEvent.click(tableNameSelect);
    let listbox = await screen.findByRole("listbox");
    // The display name for "step0.table" might be generic like "Output of Create Table step"
    // or include the (empty) name. Assuming "Output of Create Table step" for now.
    await userEvent.click(within(listbox).getByText(/output of create table step/i)); 
    await userEvent.type(screen.getByLabelText("Column Name"), "AddedColumn1");
    await userEvent.type(screen.getByLabelText("Column Description"), "Test column for MultiStepTable");
    let dataTypeSelect = screen.getByRole("combobox", { name: /data type/i });
    await userEvent.click(dataTypeSelect);
    listbox = await screen.findByRole("listbox");
    await userEvent.click(within(listbox).getByText("String"));

    // Step 3: DeleteColumn
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Delete Column"));
    tableNameSelect = screen.getByRole("combobox", { name: /table name/i });
    await userEvent.click(tableNameSelect);
    listbox = await screen.findByRole("listbox");
    await userEvent.click(within(listbox).getByText(/output of create table step/i)); 
    let columnSelect = screen.getByRole("combobox", { name: /column/i });
    await userEvent.click(columnSelect);
    listbox = await screen.findByRole("listbox");
    await userEvent.click(within(listbox).getByText("AddedColumn1")); // Display name of "step1.column"

    // Step 4: Import
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Import Data"));
    const fileVarSelect = screen.getByRole("combobox", { name: /file/i });
    await userEvent.click(fileVarSelect);
    listbox = await screen.findByRole("listbox");
    await userEvent.click(within(listbox).getByText("inputFile1"));
    // "Create new table" is default.
    await userEvent.type(screen.getByLabelText("New Table Name"), "SecondMultiStepTable");
    await userEvent.type(screen.getByLabelText("Prompt (Import Image)"), "Importing for multi-step");

    // Step 5: Generate
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Generate"));
    tableNameSelect = screen.getByRole("combobox", { name: /table name/i });
    await userEvent.click(tableNameSelect);
    listbox = await screen.findByRole("listbox");
    await userEvent.click(within(listbox).getByText(/output of create table step/i)); 
    let countInput = screen.getByLabelText("Count");
    await userEvent.clear(countInput);
    await userEvent.type(countInput, "10");
    let batchInput = screen.getByLabelText("Batch");
    await userEvent.clear(batchInput);
    await userEvent.type(batchInput, "2");

    // Step 6: ExportTable
    await userEvent.click(screen.getByText("Add Step"));
    await userEvent.click(screen.getByText("Export Data"));
    tableNameSelect = screen.getByRole("combobox", { name: /table name/i });
    await userEvent.click(tableNameSelect);
    listbox = await screen.findByRole("listbox");
    // The display name for "step3.table" (from Import) will be "SecondMultiStepTable"
    await userEvent.click(within(listbox).getByText("SecondMultiStepTable"));

    // Save Workflow
    await userEvent.click(screen.getByText("Save Workflow"));

    // Assertions
    expect(createWorkflow).toHaveBeenCalled();
    const wf = vi.mocked(createWorkflow).mock.calls[0][0];

    // Variables Assertion
    expect(wf.variables).toHaveLength(1);
    expect(wf.variables[0]).toMatchObject({ name: "inputFile1", type: "file" });

    // Steps Assertion
    expect(wf.steps).toHaveLength(6);
    // Step 1: CreateTable
    expect(wf.steps[0]).toMatchObject({ type: "CreateTable", payload: { request: { name: "", description: "" }, on_exists: "Stop" } });
    // Step 2: CreateColumn
    expect(wf.steps[1]).toMatchObject({ type: "CreateColumn", payload: { table: "step0.table", name: "AddedColumn1", description: "Test column for MultiStepTable", type: "string" } });
    // Step 3: DeleteColumn
    expect(wf.steps[2]).toMatchObject({ type: "DeleteColumn", payload: { table: "step0.table", column: "step1.column" } });
    // Step 4: Import
    expect(wf.steps[3]).toMatchObject({ type: "Import", payload: { file: "inputFile1", name: "SecondMultiStepTable", table: "", truncate: false, prompt: "Importing for multi-step" } });
    // Step 5: Generate
    expect(wf.steps[4]).toMatchObject({ type: "Generate", payload: { table: "step0.table", count: 10, batch: 2 } });
    // Step 6: ExportTable
    expect(wf.steps[5]).toMatchObject({ type: "ExportTable", payload: { table: "step3.table" } });
  });
});
