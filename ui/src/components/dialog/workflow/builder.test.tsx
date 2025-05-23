import { createWorkflow, getTables, TableInfo } from "@/actions";
import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import WorkflowBuilderDialog from "./builder";

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
});
