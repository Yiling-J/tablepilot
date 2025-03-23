import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useNavigate } from "react-router-dom";
import { Mock } from "vitest";
import { TableInfo, deleteTable, getTables } from "../actions";
import { TableListPage } from "./tables";

describe("Tables", () => {
  beforeEach(async () => {
    vi.mock("react-router-dom");
    const m = vi.mocked(useNavigate);
    m.mockReturnValue(vi.fn());
    vi.mock("@/actions");
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
    await userEvent.click(screen.getByText("users"));
    expect((m.mock.results[0].value as Mock).mock.calls[0][0]).toBe(
      "/tables/abc",
    );
  });
  it("should call delete API and fetch again when delete a table", async () => {
    vi.mock("@/actions");
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
      tables: [recipeTable],
      total: 1,
    });
    const mockedDeleteTable = vi.mocked(deleteTable);
    expect(screen.getByText("users table")).toBeInTheDocument();
    await userEvent.click(screen.getAllByText("Delete")[0]);
    expect(mockedDeleteTable.mock.calls[0][0]).toBe("abc");
    expect(screen.queryByText("users table")).toBe(null);
  });
  it("should open create table form when click add new table", async () => {
    await userEvent.click(screen.getByText("Add New Table"));
    expect(screen.getByText("Create New Table")).toBeInTheDocument();
  });
  it("should open file selector when click import", async () => {
    await userEvent.click(screen.getByText("Import"));
    expect(screen.getByText("Click to select a CSV file")).toBeInTheDocument();
  });
});
