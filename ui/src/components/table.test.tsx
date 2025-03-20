import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach } from "vitest";
import { TableInfo, getModels, getRows, getTable, getTables } from "../actions";
import { Table } from "./table";

describe("Table", () => {
  beforeEach(() => {
    vi.mock("@/actions");
    const mockedGetTable = vi.mocked(getTable);
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
    mockedGetTable.mockResolvedValue(table);
    const mockedGetRows = vi.mocked(getRows);
    mockedGetRows.mockResolvedValue([{ col1: "v1", col2: "v2" }]);

    const mockedGetTables = vi.mocked(getTables);
    mockedGetTables.mockResolvedValue({
      tables: [table],
      total: 1,
    });

    const mockedGetModels = vi.mocked(getModels);
    mockedGetModels.mockResolvedValue({ default: "ai", models: ["ai", "bi"] });
  });

  it("should render Table component", async () => {
    render(
      <TestProvider>
        <Table id="foo" />
      </TestProvider>,
    );
    let b = screen.getByRole("button", { name: /Start/i });
    expect(b).toBeDefined();
    expect((b as HTMLButtonElement).disabled).toBe(true);

    await screen.findByText("users");
    b = screen.getByRole("button", { name: /Start/i });
    expect(b).toBeDefined();
    expect((b as HTMLButtonElement).disabled).toBe(false);

    expect(screen.getByText("name")).toBeDefined();
    expect(screen.getByText("job")).toBeDefined();
    expect(screen.getByText("v1")).toBeDefined();
    expect(screen.getByText("v2")).toBeDefined();
  });

  it("should call generate API with default params", async () => {
    render(
      <TestProvider>
        <Table id="foo" />
      </TestProvider>,
    );
    let b = screen.getByRole("button", { name: /Start/i });
    expect(b).toBeDefined();
    expect((b as HTMLButtonElement).disabled).toBe(true);

    await screen.findByText("users");
    await userEvent.click(screen.getByRole("button", { name: /Start/i }));
  });
});
