import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useLocation, useNavigate } from "react-router-dom";
import { Mock } from "vitest";
import {
    TableInfo,
    deleteTable,
    getModels,
    getTableSchema, // Added getTableSchema
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

  it("should show table list", async () => {
    expect(screen.getByText("users table")).toBeInTheDocument();
    expect(screen.getByText("recipes")).toBeInTheDocument();
    expect(screen.getByText("recipes table")).toBeInTheDocument();
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

    // Click the delete button in the confirmation dialog
    const confirmDeleteButton = screen.getByRole("button", {
      name: /^delete$/i,
    });
    await userEvent.click(confirmDeleteButton as HTMLElement); // This targets the button in the dialog
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
          name: "name", // id removed
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
          name: "job", // id removed
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
      // model: "" // model removed
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
    expect(await screen.findByText("Update Table")).toBeInTheDocument(); // Dialog title is "Update Table"
  });
});
