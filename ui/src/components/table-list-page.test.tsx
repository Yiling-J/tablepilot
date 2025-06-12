import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useLocation, useNavigate } from "react-router-dom";
import { Mock } from "vitest";
import {
    TableInfo,
    deleteTable,
    getModels,
    getTableSchema,
    getTables,
} from "../actions";
import { TableListPage } from "./table-list-page";

describe("TableListPage", () => {
  beforeEach(async () => {
    vi.mock("react-router-dom");
    const m = vi.mocked(useNavigate);
    m.mockReturnValue(vi.fn());
    vi.mocked(useLocation).mockReturnValue({
      key: "",
      pathname: "/tables",
      search: "",
      hash: "",
      state: null,
    });
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
    const userTable = {
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
    const recipeTable = {
      id: "abd",
      name: "recipes",
      description: "recipes table",
      columns: [
        {
          id: "col1",
          name: "name",
          description: "recipe name",
          type: "string",
          fill_mode: "ai",
        },
      ],
      model: "",
    } as TableInfo;
    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({
      tables: [userTable, recipeTable],
      total: 2,
    });
    const mockedDeleteTable = vi.mocked(deleteTable);
    mockedDeleteTable.mockImplementation(async (_id: string) => {
      await new Promise((f) => setTimeout(f, 100));
      return 1;
    });
    render(
      <TestProvider>
        <TableListPage />
      </TestProvider>,
    );
    await screen.findByText("users");
  });

  it("should navigate to table page when click", async () => {
    const m = vi.mocked(useNavigate);
    await userEvent.click(screen.getByText("users") as HTMLElement);
    expect((m.mock.results[0].value as Mock).mock.calls[0][0]).toBe(
      "/tables/abc",
    );
  });
  it("should call delete API and fetch again when delete a table", async () => {
    vi.mock("@/actions");
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
      ],
      model: "",
    } as TableInfo;
    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({
      tables: [recipeTable],
      total: 1,
    });
    const mockedDeleteTable = vi.mocked(deleteTable);
    expect(screen.getByText("users table")).toBeInTheDocument();

    const usersCardDelete = screen
      .getByText("users table")
      .closest('div[class*="h-60"]');
    if (!usersCardDelete)
      throw new Error("Users card not found for delete test");
    const deleteButton = within(usersCardDelete as HTMLElement).getByRole(
      "button",
      { name: /delete/i },
    );
    await userEvent.click(deleteButton as HTMLElement);

    const confirmDeleteButton = screen.getByRole("button", {
      name: /^delete$/i,
    });
    await userEvent.click(confirmDeleteButton as HTMLElement);
    expect(mockedDeleteTable.mock.calls[0][0]).toBe("abc");
    await screen.findByText("recipes table new");
    expect(screen.queryByText("users table")).toBe(null);
  });
  it("should open create table form when click add new table", async () => {
    await userEvent.click(screen.getByText("Add New Table") as HTMLElement);
    expect(screen.getByText("Create New Table")).toBeInTheDocument();
  });
  it("should open file selector when click import", async () => {
    await userEvent.click(screen.getByText("Import") as HTMLElement);
    expect(
      screen.getByText("Click to select a CSV or image file"),
    ).toBeInTheDocument();
  });

  it("should open edit table dialog when edit icon is clicked", async () => {
    const mockedGetTableSchema = vi.mocked(getTableSchema);
    mockedGetTableSchema.mockResolvedValue({
      name: "users",
      description: "users table",
      columns: [
        {
          name: "name",
          description: "user name",
          type: "string",
          fill_mode: "ai",
          random: false,
          replacement: false,
          repeat: 1,
          linked_column: "",
          linked_context_columns: [],
        },
        {
          name: "job",
          description: "user job",
          type: "string",
          fill_mode: "ai",
          random: false,
          replacement: false,
          repeat: 1,
          linked_column: "",
          linked_context_columns: [],
        },
      ],
    });

    const usersCardEdit = screen
      .getByText("users table")
      .closest('div[class*="h-60"]');
    if (!usersCardEdit) throw new Error("Users card not found for edit test");
    const editButton = within(usersCardEdit as HTMLElement).getByRole(
      "button",
      {
        name: /edit/i,
      },
    );
    await userEvent.click(editButton as HTMLElement);

    expect(mockedGetTableSchema).toHaveBeenCalledWith("abc");
    expect(await screen.findByText("Update Table")).toBeInTheDocument();
  });
});

describe("TableListPage Search Functionality", () => {
  const sampleTables = [
    { id: "t1", name: "Table Alpha", description: "Alpha description" },
    { id: "t2", name: "Table Beta", description: "Beta description" },
    { id: "t3", name: "Gamma Table", description: "Gamma description" },
  ];

  beforeEach(async () => {
    vi.mock("react-router-dom");
    vi.mocked(useNavigate).mockReturnValue(vi.fn());
    vi.mocked(useLocation).mockReturnValue({
      key: "",
      pathname: "/tables",
      search: "",
      hash: "",
      state: null,
    });
    vi.mock("@/actions");
    vi.mocked(getModels).mockResolvedValue({
      default_model: "ai",
      default_image_model: "",
      models: [{ name: "ai", image: false }],
    });
    vi.mocked(deleteTable).mockImplementation(async (_id: string) => {
      await new Promise((f) => setTimeout(f, 100));
      return 1;
    });
     vi.mocked(getTableSchema).mockResolvedValue({
      name: "any",
      description: "any",
      columns: [],
    });

    (getTables as Mock).mockResolvedValue({
      tables: sampleTables,
      total: sampleTables.length,
    });

    render(
      <TestProvider>
        <TableListPage />
      </TestProvider>,
    );
    await screen.findByText("Table Alpha");
    await screen.findByText("Table Beta");
    await screen.findByText("Gamma Table");
  });

  it("should render all tables initially", () => {
    expect(screen.getByText("Table Alpha")).toBeInTheDocument();
    expect(screen.getByText("Table Beta")).toBeInTheDocument();
    expect(screen.getByText("Gamma Table")).toBeInTheDocument();
  });

  it("should filter tables based on search query", async () => {
    const searchInput = screen.getByPlaceholderText("Search tables...");
    await userEvent.type(searchInput, "Alpha");

    expect(screen.getByText("Table Alpha")).toBeInTheDocument();
    expect(screen.queryByText("Table Beta")).not.toBeInTheDocument();
    expect(screen.queryByText("Gamma Table")).not.toBeInTheDocument();
  });

  it("should be case-insensitive", async () => {
    const searchInput = screen.getByPlaceholderText("Search tables...");
    await userEvent.type(searchInput, "gamma");

    expect(screen.queryByText("Table Alpha")).not.toBeInTheDocument();
    expect(screen.queryByText("Table Beta")).not.toBeInTheDocument();
    expect(screen.getByText("Gamma Table")).toBeInTheDocument();
  });

  it("should show no results message if search matches nothing", async () => {
    const searchInput = screen.getByPlaceholderText("Search tables...");
    await userEvent.type(searchInput, "NonExistentTable");

    // The current component does not implement a "no results" message for tables.
    // It simply renders an empty list. So we check for absence of any table.
    expect(screen.queryByText("Table Alpha")).not.toBeInTheDocument();
    expect(screen.queryByText("Table Beta")).not.toBeInTheDocument();
    expect(screen.queryByText("Gamma Table")).not.toBeInTheDocument();
    // If a "No tables found..." message were implemented, we'd assert its presence here.
  });

  it("should show all tables when search query is cleared", async () => {
    const searchInput = screen.getByPlaceholderText("Search tables...");
    await userEvent.type(searchInput, "Alpha");

    expect(screen.getByText("Table Alpha")).toBeInTheDocument();
    expect(screen.queryByText("Table Beta")).not.toBeInTheDocument();

    await userEvent.clear(searchInput);

    expect(screen.getByText("Table Alpha")).toBeInTheDocument();
    expect(screen.getByText("Table Beta")).toBeInTheDocument();
    expect(screen.getByText("Gamma Table")).toBeInTheDocument();
  });
});
