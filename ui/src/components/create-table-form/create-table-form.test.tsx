import {
    TableCreateRequest,
    TableInfo,
    createRows,
    createTable,
    getDatasets,
    getTables,
    updateTable,
} from "@/actions";
import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useNavigate } from "react-router-dom";
import CreateTableForm from "./create-table-form";

describe("CreateTableForm", () => {
  beforeEach(() => {
    vi.mock("react-router-dom");
    vi.mocked(useNavigate).mockReturnValue(vi.fn());
  });

  it("should render first tab", async () => {
    render(
      <TestProvider>
        <CreateTableForm close={() => {}} />
      </TestProvider>,
    );
    await screen.findByText("Table Configuration");
    expect(screen.getByText("Basic")).toBeInTheDocument();
    expect(screen.getByText("Columns")).toBeDisabled();
  });

  it("should validate table name", async () => {
    render(
      <TestProvider>
        <CreateTableForm close={() => {}} />
      </TestProvider>,
    );
    await screen.findByText("Table Configuration");
    const input = screen.getByPlaceholderText(
      "Only letters, numbers, and underscores, and start with a letter",
    );
    expect(screen.getByText("Table name cannot be empty.")).toBeInTheDocument();
    await userEvent.click(input);
    await userEvent.keyboard("test");
    expect(screen.getByText("Columns")).toBeEnabled();
    await userEvent.clear(input);
    expect(screen.getByText("Columns")).toBeDisabled();
    await userEvent.keyboard("****");
    expect(screen.getByText("Columns")).toBeDisabled();
    expect(
      screen.getByText(
        "Table name must start with a letter and contain only letters, numbers, or underscores.",
      ),
    ).toBeInTheDocument();
  });

  it("should update first tab form data", async () => {
    const form = { name: "", description: "", sources: [], columns: [] };
    render(
      <TestProvider>
        <CreateTableForm close={() => {}} form={form} />
      </TestProvider>,
    );
    await screen.findByText("Table Configuration");
    const input = screen.getByPlaceholderText(
      "Only letters, numbers, and underscores, and start with a letter",
    );
    await userEvent.click(input);
    await userEvent.keyboard("test");
    const descInput = screen.getByPlaceholderText("Enter table description");
    await userEvent.click(descInput);
    await userEvent.keyboard("foobar");

    // show json will load formdata
    await userEvent.click(screen.getByText("Show JSON"));
    await screen.findByText("JSON Preview");
    expect(screen.getByTestId("json-preview")).toHaveTextContent(
      `{ "name": "test", "description": "foobar", "sources": [], "columns": [] }`,
    );
  });

  it("show second tab when click next", async () => {
    vi.mock("@/actions");
    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({
      tables: [],
      total: 0,
    });
    const form = { name: "foo", description: "bar", sources: [], columns: [] };
    render(
      <TestProvider>
        <CreateTableForm close={() => {}} form={form} />
      </TestProvider>,
    );
    await screen.findByText("Table Configuration");
    await userEvent.click(screen.getByText("Next"));
    await screen.findByText("Columns");
  });

  describe("AddColumns", () => {
    beforeEach(async () => {
      vi.mock("@/actions");
      const table = {
        id: "abc",
        name: "users",
        description: "users table",
        columns: [
          {
            id: "col1",
            name: "name",
            description: "user name",
            type: "string",
            fill_mode: "ai",
          },
          {
            id: "col2",
            name: "job",
            description: "user job",
            type: "string",
            fill_mode: "ai",
          },
        ],
        model: "",
      } as TableInfo;
      const mockedGetTables = vi.mocked(getTables);
      mockedGetTables.mockResolvedValue({
        tables: [table],
        total: 1,
      });
      const mockedGetDatasets = vi.mocked(getDatasets);
      mockedGetDatasets.mockResolvedValue({
        datasets: [
          {
            id: "d1",
            name: "s1",
            description: "ds",
            type: "csv",
            data: [],
            columns: [],
          },
        ],
        total: 1,
      });
      const form = {
        name: "foo",
        description: "bar",
        columns: [],
      };
      render(
        <TestProvider>
          <CreateTableForm close={() => {}} form={form} />
        </TestProvider>,
      );
      await screen.findByText("Table Configuration");
      await userEvent.click(screen.getByText("Next"));
      await userEvent.click(screen.getByText("Add Column"));
      await screen.findByText("Add New Column");
      await userEvent.click(
        screen.getByPlaceholderText("e.g., Name, Ingredients"),
      );
      await userEvent.keyboard("c1");
      await userEvent.click(
        screen.getByPlaceholderText("e.g., recipe name, list of ingredients"),
      );
      await userEvent.keyboard("recipe name");
    });

    it("should add a new default type column and can edit/delete", async () => {
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "columns": [ { "name": "c1", "description": "recipe name", "type": "string", "fill_mode": "ai", "random": true, "replacement": false, "repeat": 1, "linked_column": "", "linked_context_columns": [], "options": [] } ] }`,
      );
      expect(screen.getByText("c1")).toBeInTheDocument();
      // edit
      await userEvent.click(screen.getByTestId("column-ops").children.item(0)!);
      await screen.findByText("Edit Column");
      expect(screen.getByDisplayValue("c1")).toBeInTheDocument();
      await userEvent.click(screen.getByText("Update"));

      // delete
      await userEvent.click(screen.getByTestId("column-ops").children.item(1)!);
      expect(
        screen.getByText(
          `No columns added yet. Click the "Add Column" button to create one.`,
        ),
      ).toBeInTheDocument();
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "columns": [] }`,
      );
    });

    it("should add a new integer type column with options", async () => {
      await userEvent.click(screen.getByText("String").parentElement!);
      await userEvent.click(screen.getByText("Integer"));
      await userEvent.click(screen.getByPlaceholderText("e.g., 5"));
      await userEvent.keyboard("12");
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "columns": [ { "name": "c1", "description": "recipe name", "type": "integer", "fill_mode": "ai", "random": true, "replacement": false, "repeat": 1, "linked_column": "", "linked_context_columns": [], "context_length": 12, "options": [] } ] }`,
      );
      expect(screen.getByText("c1")).toBeInTheDocument();
    });

    it("should show tables in source list", async () => {
      await userEvent.click(screen.getByText("AI Generated").parentElement!);
      await userEvent.click(screen.getByText("Select from Table"));
      await userEvent.click(screen.getByText("Select a table").parentElement!);
      expect(screen.getAllByText("users").length).toBe(1);
      expect(screen.getByText("users")).toBeInTheDocument();
    });

    it("should show datasets in source list", async () => {
      await userEvent.click(screen.getByText("AI Generated").parentElement!);
      await userEvent.click(screen.getByText("Select from Dataset"));
      await userEvent.click(
        screen.getByText("Select a dataset").parentElement!,
      );
      expect(screen.getAllByText("s1").length).toBe(1);
      expect(screen.getByText("s1")).toBeInTheDocument();
    });

    it("should create pick from dataset column", async () => {
      await userEvent.click(screen.getByText("AI Generated").parentElement!);
      await userEvent.click(screen.getByText("Select from Dataset"));
      await userEvent.click(
        screen.getByText("Select a dataset").parentElement!,
      );
      await userEvent.click(screen.getByText("s1"));
      await userEvent.click(
        screen.getByText("Random Selection").parentElement!
          .firstElementChild as HTMLElement,
      );
      await userEvent.click(
        screen.getByText("Selection with Replacement").parentElement!
          .firstElementChild as HTMLElement,
      );
      await userEvent.click(screen.getByDisplayValue("1"));
      await userEvent.keyboard("2");
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "columns": [ { "name": "c1", "description": "recipe name", "type": "string", "fill_mode": "pick", "random": false, "replacement": true, "repeat": 12, "linked_column": "", "linked_context_columns": [], "source_id": "d1", "source_type": "dataset", "options": [] } ] }`,
      );
    });

    it("should create pick from table column", async () => {
      await userEvent.click(screen.getByText("AI Generated").parentElement!);
      await userEvent.click(screen.getByText("Select from Table"));
      await userEvent.click(screen.getByText("Select a table").parentElement!);
      await userEvent.click(screen.getByText("users"));
      await userEvent.click(screen.getByText("Select a column").parentElement!);
      await userEvent.click(screen.getByText("name"));
      await userEvent.click(screen.getByText("Select context columns"));
      await userEvent.click(screen.getByText("job"));
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "columns": [ { "name": "c1", "description": "recipe name", "type": "string", "fill_mode": "pick", "random": true, "replacement": false, "repeat": 1, "linked_column": "name", "linked_context_columns": [ "job" ], "source_id": "abc", "source_type": "table", "options": [] } ] }`,
      );
    });

    it("should call create API when click complete", async () => {
      const mockedCreateTable = vi.mocked(createTable);
      mockedCreateTable.mockResolvedValue({
        id: "t1",
        name: "",
        description: "",
        model: "",
        columns: [],
      });
      await userEvent.click(screen.getByText("Add"));
      const mockedGetTables = vi.mocked(getTables);
      mockedGetTables.mockReset();
      expect(mockedGetTables.mock.calls.length).toBe(0);
      await userEvent.click(screen.getByText("Complete"));
      expect(mockedCreateTable.mock.calls[0][0]).toMatchObject({
        name: "foo",
        description: "bar",
        columns: [
          {
            name: "c1",
            description: "recipe name",
            type: "string",
            fill_mode: "ai",
            random: true,
            replacement: false,
            repeat: 1,
            linked_column: "",
            linked_context_columns: [],
          },
        ],
      });
      expect(mockedGetTables.mock.calls.length).toBe(1);
    });
  });
});

describe("CreateTableFormWithRows", () => {
  it("should call create rows API is rows is not empty", async () => {
    vi.mock("react-router-dom");
    const m = vi.mocked(useNavigate);
    m.mockReturnValue(vi.fn());
    vi.mock("@/actions");
    const table = {
      id: "abc",
      name: "users",
      description: "users table",
      columns: [],
      model: "",
    } as TableInfo;
    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({
      tables: [table],
      total: 1,
    });
    const mockedGetDatasets = vi.mocked(getDatasets);
    mockedGetDatasets.mockResolvedValue({
      datasets: [
        {
          id: "d1",
          name: "s1",
          description: "ds",
          type: "csv",
          data: [],
          columns: [],
        },
      ],
      total: 1,
    });
    const mockedCreateRows = vi.mocked(createRows);
    const mockedCreateTable = vi.mocked(createTable);
    mockedCreateTable.mockResolvedValue({
      id: "t1",
      name: "",
      description: "",
      model: "",
      columns: [],
    });

    const form = {
      name: "foo",
      description: "bar",
      sources: [],
      columns: [
        {
          name: "c",
          description: "",
          type: "string",
          fill_mode: "ai",
          random: false,
          replacement: false,
          repeat: 1,
          linked_column: "",
          linked_context_columns: [],
        },
      ],
    } as TableCreateRequest;
    render(
      <TestProvider>
        <CreateTableForm
          close={() => {}}
          form={form}
          rows={[{ name: "foo" }, { name: "bar" }]}
        />
      </TestProvider>,
    );
    await screen.findByText("Table Configuration");
    await userEvent.click(screen.getByText("Next"));
    await userEvent.click(screen.getByText("Complete"));
    expect(mockedCreateRows.mock.calls[0][0]).toBe("t1");
    expect(mockedCreateRows.mock.calls[0][1]).toMatchObject([
      { name: "foo" },
      { name: "bar" },
    ]);
  });
});

describe("UpdateTableForm", () => {
  it("should call update api when complete", async () => {
    vi.mock("react-router-dom");
    vi.mocked(useNavigate).mockReturnValue(vi.fn());
    vi.mock("@/actions");
    const table = {
      id: "abc",
      name: "users",
      description: "users table",
      columns: [
        {
          id: "col1",
          name: "name",
          description: "user name",
          type: "string",
          fill_mode: "ai",
        },
        {
          id: "col2",
          name: "job",
          description: "user job",
          type: "string",
          fill_mode: "ai",
        },
      ],
      model: "",
    } as TableInfo;
    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({
      tables: [table],
      total: 1,
    });
    const mockedGetDatasets = vi.mocked(getDatasets);
    mockedGetDatasets.mockResolvedValue({
      datasets: [
        {
          id: "d1",
          name: "s1",
          description: "ds",
          type: "csv",
          data: [],
          columns: [],
        },
      ],
      total: 1,
    });
    const form = {
      name: "foo",
      description: "bar",
      columns: [
        {
          name: "c1",
          description: "recipe name",
          type: "string",
          fill_mode: "ai",
          random: true,
          replacement: false,
          repeat: 1,
          linked_column: "",
          linked_context_columns: [],
        },
      ],
    };
    const mockedUpdateTable = vi.mocked(updateTable);
    mockedUpdateTable.mockResolvedValue({
      id: "tb1",
      name: "",
      description: "",
      model: "",
      columns: [],
    });
    let cb = 0;
    render(
      <TestProvider>
        <CreateTableForm
          close={() => {}}
          form={form}
          table={"tb1"}
          submitCallback={async () => {
            cb++;
          }}
        />
      </TestProvider>,
    );
    await screen.findByText("Update your table configuration or import JSON");
    await userEvent.click(screen.getByText("Next"));
    mockedGetTables.mockReset();
    await userEvent.click(screen.getByText("Complete"));
    expect(mockedUpdateTable.mock.calls[0][0]).toBe("tb1");
    expect(mockedUpdateTable.mock.calls[0][1]).toMatchObject(form);
    expect(mockedGetTables.mock.calls.length).toBe(1);
    expect(cb).toBe(1);
  });
});
