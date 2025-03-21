import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CsvOutput, download } from "export-to-csv";
import { beforeEach } from "vitest";
import {
    GenerateRequest,
    TableInfo,
    generate,
    getModels,
    getRows,
    getTable,
    getTables,
    truncateTable,
} from "../actions";
import { Table } from "./table";

/**
 * JSDOM doesn't implement PointerEvent so we need to mock our own implementation
 * Default to mouse left click interaction
 * https://github.com/radix-ui/primitives/issues/1822
 * https://github.com/jsdom/jsdom/pull/2666
 */
class MockPointerEvent extends Event {
  button: number;
  ctrlKey: boolean;
  pointerType: string;

  constructor(type: string, props: PointerEventInit) {
    super(type, props);
    this.button = props.button || 0;
    this.ctrlKey = props.ctrlKey || false;
    this.pointerType = props.pointerType || "mouse";
  }
}

/* eslint-disable-next-line  @typescript-eslint/no-explicit-any */
window.PointerEvent = MockPointerEvent as any;
window.HTMLElement.prototype.scrollIntoView = vi.fn();
window.HTMLElement.prototype.setPointerCapture = vi.fn();
window.HTMLElement.prototype.releasePointerCapture = vi.fn();
window.HTMLElement.prototype.hasPointerCapture = vi.fn();

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
    expect(b).toBeInTheDocument();
    expect((b as HTMLButtonElement).disabled).toBe(false);

    expect(screen.getByText("name")).toBeInTheDocument();
    expect(screen.getByText("job")).toBeInTheDocument();
    expect(screen.getByText("v1")).toBeInTheDocument();
    expect(screen.getByText("v2")).toBeInTheDocument();
  });

  it("should call generate API with default params", async () => {
    render(
      <TestProvider>
        <Table id="foo" />
      </TestProvider>,
    );
    const b = screen.getByRole("button", { name: /Start/i });
    expect(b).toBeInTheDocument();
    expect((b as HTMLButtonElement).disabled).toBe(true);

    await screen.findByText("users");
    const mockedGenerate = vi.mocked(generate);
    await userEvent.click(screen.getByRole("button", { name: /Start/i }));
    expect(mockedGenerate).toHaveBeenCalledWith(
      "foo",
      expect.anything(),
      expect.anything(),
      {
        batch: 10,
        count: 50,
        temperature: 0.6,
        model: "ai",
      },
    );
  });

  it("should call generate API with given params", async () => {
    render(
      <TestProvider>
        <Table id="foo" />
      </TestProvider>,
    );
    const b = screen.getByRole("button", { name: /Start/i });
    expect(b).toBeInTheDocument();
    expect((b as HTMLButtonElement).disabled).toBe(true);

    await screen.findByText("users");
    const mockedGenerate = vi.mocked(generate);
    await userEvent.dblClick(screen.getByDisplayValue("10"));
    await userEvent.keyboard("35");
    await userEvent.dblClick(screen.getByDisplayValue("50"));
    await userEvent.keyboard("100");
    await userEvent.click(screen.getByText("ai").closest("button")!);
    await userEvent.click(screen.getByText("bi"));
    await userEvent.click(screen.getByRole("button", { name: /Start/i }));
    expect(mockedGenerate).toHaveBeenCalledWith(
      "foo",
      expect.anything(),
      expect.anything(),
      {
        batch: 35,
        count: 100,
        temperature: 0.6,
        model: "bi",
      },
    );
  });

  it("should update table when receiving data from API", async () => {
    render(
      <TestProvider>
        <Table id="foo" />
      </TestProvider>,
    );
    const b = screen.getByRole("button", { name: /Start/i });
    expect(b).toBeInTheDocument();
    expect((b as HTMLButtonElement).disabled).toBe(true);

    await screen.findByText("users");
    const mockedGenerate = vi.mocked(generate);
    mockedGenerate.mockImplementation(
      async (
        _table: string,
        _signal: AbortSignal,
        callback: (data: string) => void,
        _genreq: GenerateRequest,
      ) => {
        callback(`{"data":[{"col1":"Alice","col2":"Software Engineer"}]}`);
        await new Promise((f) => setTimeout(f, 100));
        callback(`{"data":[{"col1":"Marco","col2":"Chef"}]}`);
        callback("[DONE]");
      },
    );
    await userEvent.click(screen.getByRole("button", { name: /Start/i }));
    await screen.findByRole("button", { name: /Stop/i });
    await screen.findByText("Marco");
    ["Alice", "Software Engineer", "Marco", "Chef"].forEach((v) =>
      expect(screen.getByText(v)).toBeInTheDocument(),
    );
    expect(screen.getByRole("button", { name: /Start/i })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /output.csv/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("Rows: 3")).toBeInTheDocument();
  });

  it("should call truncate API and fetch data again when click truncate", async () => {
    render(
      <TestProvider>
        <Table id="foo" />
      </TestProvider>,
    );

    await screen.findByText("users");
    const mockedTruncate = vi.mocked(truncateTable);
    const mockedGetRows = vi.mocked(getRows);
    mockedGetRows.mockReset();
    mockedGetRows.mockResolvedValue([]);
    await userEvent.click(screen.getByText(/Truncate table/i));
    await screen.findByText("Confirm");
    await userEvent.click(screen.getByRole("button", { name: "Confirm" }));

    expect(mockedTruncate.mock.calls.length).toBe(1);
    expect(mockedTruncate.mock.calls[0][0]).toBe("foo");
    expect(mockedGetRows.mock.calls.length).toBe(1);
    expect(screen.getByText("Rows: 0")).toBeInTheDocument();
  });

  it("should down the table in csv format", async () => {
    vi.mock("export-to-csv", async () => {
      const originalModule = await vi.importActual("export-to-csv");
      return { ...originalModule, download: vi.fn() };
    });
    render(
      <TestProvider>
        <Table id="foo" />
      </TestProvider>,
    );

    await screen.findByText("users");
    const mockedCSVDownload = vi.mocked(download);
    let called = false;
    mockedCSVDownload.mockReturnValue((v: CsvOutput) => {
      called = true;
      expect(String(v)).toBe(`\ufeff"name","job"\r\n"v1","v2"\r\n`);
    });
    await userEvent.click(screen.getByText(/output.csv/i));
    expect(called).toBe(true);
  });

  it("array display as list", async () => {
    const mockedGetTable = vi.mocked(getTable);
    const table = {
      id: "abc",
      name: "users",
      description: "users table",
      columns: [
        {
          id: "col1",
          name: "names",
          description: "user names",
          type: "array",
          fill_mode: "ai",
        },
      ],
      model: "",
    } as TableInfo;
    mockedGetTable.mockResolvedValue(table);
    const mockedGetRows = vi.mocked(getRows);
    mockedGetRows.mockResolvedValue([{ col1: ["ll0", "ll1", "ll2"] }]);

    render(
      <TestProvider>
        <Table id="foo" />
      </TestProvider>,
    );

    await screen.findByText("users");
    const e = screen.getByText(/ll0/i);
    expect(e.outerHTML).toBe(`<div class="max-h-80 line-clamp-6">• ll0
• ll1
• ll2</div>`);
  });

  it("text cell should expand when click", async () => {
    render(
      <TestProvider>
        <Table id="foo" />
      </TestProvider>,
    );

    await screen.findByText("users");
    await userEvent.hover(screen.getByText("v1"));
    await userEvent.click(
      screen.getByText("v1").parentNode?.parentNode?.children.item(1)
        ?.lastElementChild as HTMLElement,
    );
    await screen.findByRole("dialog");
    expect((await screen.findAllByText("v1")).length).toBe(2);
  });
});
