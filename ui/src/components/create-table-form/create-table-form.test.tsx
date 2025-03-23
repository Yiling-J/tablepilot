import {
    AiSource,
    LinkedSource,
    TableInfo,
    createTable,
    getSources,
    getTables,
} from "@/actions";
import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useNavigate } from "react-router-dom";
import CreateTableForm from "./create-table-form";

describe("CreateTableForm", () => {
  beforeEach(() => {
    vi.mock("react-router-dom");
    vi.mocked(useNavigate);
  });

  it("should render first tab", async () => {
    render(
      <TestProvider>
        <CreateTableForm close={() => {}} />
      </TestProvider>,
    );
    await screen.findByText("Table Configuration");
    expect(screen.getByText("Basic")).toBeInTheDocument();
    expect(screen.getByText("Sources")).toBeDisabled();
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
    expect(screen.getByText("Sources")).toBeEnabled();
    expect(screen.getByText("Columns")).toBeEnabled();
    await userEvent.clear(input);
    expect(screen.getByText("Sources")).toBeDisabled();
    expect(screen.getByText("Columns")).toBeDisabled();
    await userEvent.keyboard("****");
    expect(screen.getByText("Sources")).toBeDisabled();
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
    await screen.findByText("Data Sources");
  });

  describe("AddSource", () => {
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
      const form = {
        name: "foo",
        description: "bar",
        sources: [],
        columns: [],
      };
      render(
        <TestProvider>
          <CreateTableForm close={() => {}} form={form} />
        </TestProvider>,
      );
      await screen.findByText("Table Configuration");
      await userEvent.click(screen.getByText("Next"));
      await userEvent.click(screen.getByText("Add Source"));
      await screen.findByText("Add New Source");
    });
    it("should add a new ai type source, and can edit/delete", async () => {
      await userEvent.click(
        screen.getByPlaceholderText("e.g., cuisines, meals, customer"),
      );
      await userEvent.keyboard("s1");
      await userEvent.click(
        screen.getByPlaceholderText("e.g., Generate 20 recipe cuisines."),
      );
      await userEvent.keyboard("20 tags");
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "sources": [ { "name": "s1", "type": "ai", "prompt": "20 tags" } ], "columns": [] }`,
      );
      expect(screen.getByText("s1")).toBeInTheDocument();
      expect(screen.getByText("AI Generated")).toBeInTheDocument();

      // edit
      await userEvent.click(screen.getByTestId("source-ops").children.item(0)!);
      await screen.findByText("Table Configuration");
      expect(screen.getByDisplayValue("s1")).toBeInTheDocument();
      await userEvent.click(screen.getByText("Update"));

      // delete
      await userEvent.click(screen.getByTestId("source-ops").children.item(1)!);
      expect(
        screen.getByText(
          `No sources added yet. Click the "Add Source" button to create one.`,
        ),
      ).toBeInTheDocument();
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "sources": [], "columns": [] }`,
      );
    });

    it("should add a new list type source", async () => {
      await userEvent.click(
        screen.getByPlaceholderText("e.g., cuisines, meals, customer"),
      );
      await userEvent.keyboard("s1");
      await userEvent.click(screen.getByText("AI Generated").parentElement!);
      await userEvent.click(screen.getByText("List of Options"));
      await userEvent.click(screen.getByPlaceholderText(/Dinner/i));
      await userEvent.keyboard(`foo
bar`);
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "sources": [ { "name": "s1", "type": "list", "options": [ "foo", "bar" ] } ], "columns": [] }`,
      );
      expect(screen.getByText("s1")).toBeInTheDocument();
      expect(screen.getByText("List of Options")).toBeInTheDocument();
    });

    it("should add a new linked table type source", async () => {
      await userEvent.click(
        screen.getByPlaceholderText("e.g., cuisines, meals, customer"),
      );
      await userEvent.keyboard("s1");
      await userEvent.click(screen.getByText("AI Generated").parentElement!);
      await userEvent.click(screen.getByText("Linked Table"));
      await userEvent.click(screen.getByText("Select a table").parentElement!);
      await userEvent.click(screen.getByText("users"));
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "sources": [ { "name": "s1", "type": "linked", "table": "users" } ], "columns": [] }`,
      );
      expect(screen.getByText("s1")).toBeInTheDocument();
      expect(screen.getByText("Linked Table")).toBeInTheDocument();
    });
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
      const mockedGetSources = vi.mocked(getSources);
      mockedGetSources.mockResolvedValue([
        {
          name: "s1",
          data: {},
          columns: [],
        },
        {
          name: "s3",
          data: { name: "s3", type: "ai", prompt: "foo" },
          columns: [],
        },
      ]);
      const form = {
        name: "foo",
        description: "bar",
        sources: [
          { name: "s1", type: "ai", prompt: "" } as AiSource,
          { name: "s2", type: "linked", table: "users" } as LinkedSource,
        ],
        columns: [],
      };
      render(
        <TestProvider>
          <CreateTableForm close={() => {}} form={form} />
        </TestProvider>,
      );
      await screen.findByText("Table Configuration");
      await userEvent.click(screen.getByText("Next"));
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
    it("should add a new default type source and can edit/delete", async () => {
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "sources": [ { "name": "s1", "type": "ai", "prompt": "" }, { "name": "s2", "type": "linked", "table": "users" } ], "columns": [ { "name": "c1", "description": "recipe name", "type": "string", "fill_mode": "ai", "random": true, "replacement": false, "repeat": 1, "linked_column": "", "linked_context_columns": [] } ] }`,
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
        `{ "name": "foo", "description": "bar", "sources": [ { "name": "s1", "type": "ai", "prompt": "" }, { "name": "s2", "type": "linked", "table": "users" } ], "columns": [] }`,
      );
    });
    it("should add a new integer type source with options", async () => {
      await userEvent.click(screen.getByText("String").parentElement!);
      await userEvent.click(screen.getByText("Integer"));
      await userEvent.click(screen.getByPlaceholderText("e.g., 5"));
      await userEvent.keyboard("12");
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "sources": [ { "name": "s1", "type": "ai", "prompt": "" }, { "name": "s2", "type": "linked", "table": "users" } ], "columns": [ { "name": "c1", "description": "recipe name", "type": "integer", "fill_mode": "ai", "random": true, "replacement": false, "repeat": 1, "linked_column": "", "linked_context_columns": [], "context_length": 12 } ] }`,
      );
      expect(screen.getByText("c1")).toBeInTheDocument();
    });
    it("should show sources and shared sources in source list", async () => {
      await userEvent.click(screen.getByText("AI Generated").parentElement!);
      await userEvent.click(screen.getByText("Pick from Source"));
      await userEvent.click(screen.getByText("Select source").parentElement!);
      expect(screen.getAllByText("s1").length).toBe(1);
      expect(screen.getByText("s1")).toBeInTheDocument();
      expect(screen.getByText("s2")).toBeInTheDocument();
      expect(screen.getByText("s3")).toBeInTheDocument();
    });
    it("should create pick from ai list column", async () => {
      await userEvent.click(screen.getByText("AI Generated").parentElement!);
      await userEvent.click(screen.getByText("Pick from Source"));
      await userEvent.click(screen.getByText("Select source").parentElement!);
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
        `{ "name": "foo", "description": "bar", "sources": [ { "name": "s1", "type": "ai", "prompt": "" }, { "name": "s2", "type": "linked", "table": "users" } ], "columns": [ { "name": "c1", "description": "recipe name", "type": "string", "fill_mode": "pick", "random": false, "replacement": true, "repeat": 12, "linked_column": "", "linked_context_columns": [], "source": "s1" } ] }`,
      );
    });
    it("should create pick from table column", async () => {
      await userEvent.click(screen.getByText("AI Generated").parentElement!);
      await userEvent.click(screen.getByText("Pick from Source"));
      await userEvent.click(screen.getByText("Select source").parentElement!);
      await userEvent.click(screen.getByText("s2"));
      await userEvent.click(screen.getByText("Select a column").parentElement!);
      await userEvent.click(screen.getByText("name"));
      await userEvent.click(screen.getByText("Select context columns"));
      await userEvent.click(screen.getByText("job"));
      await userEvent.click(screen.getByText("Add"));
      await userEvent.click(screen.getByText("Show JSON"));
      await screen.findByText("JSON Preview");
      expect(screen.getByTestId("json-preview")).toHaveTextContent(
        `{ "name": "foo", "description": "bar", "sources": [ { "name": "s1", "type": "ai", "prompt": "" }, { "name": "s2", "type": "linked", "table": "users" } ], "columns": [ { "name": "c1", "description": "recipe name", "type": "string", "fill_mode": "pick", "random": true, "replacement": false, "repeat": 1, "linked_column": "name", "linked_context_columns": [ "job" ], "source": "s2" } ] }`,
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
        sources: [
          { name: "s1", type: "ai", prompt: "" },
          { name: "s2", type: "linked", table: "users" },
        ],
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
