import { TableCreateRequest } from "@/actions";
import { JSONObject } from "@/json";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ImportCSVDialog } from "./import-csv";

describe("ImportCSV", () => {
  it("should parse the csv file", async () => {
    let nextRun = false;
    const onNext = (form: TableCreateRequest, rows: JSONObject[]) => {
      expect(form).toMatchObject({
        name: "users",
        description: "",
        sources: [],
        columns: [
          {
            name: "name",
            description: "",
            type: "string",
            fill_mode: "ai",
            random: true,
            replacement: false,
            repeat: 1,
            linked_column: "",
            linked_context_columns: [],
          },
          {
            name: "job",
            description: "",
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
      expect(rows).toMatchObject([
        { name: "Alice", job: "Engineer" },
        { name: "Bob", job: "Designer" },
        { name: "Charlie", job: "Manager" },
      ]);
      nextRun = true;
    };
    render(
      <ImportCSVDialog isOpen={true} setIsOpen={vi.fn()} onNext={onNext} />,
    );
    const csvContent =
      "name,job\nAlice,Engineer\nBob,Designer\nCharlie,Manager";
    const blob = new Blob([csvContent], { type: "text/csv" });
    const file = new File([blob], "users.csv", { type: "text/csv" });
    await userEvent.upload(screen.getByTestId("import-file-selector"), file);
    expect(screen.getByText("users.csv")).toBeInTheDocument();
    await userEvent.click(screen.getByText("Next"));
    expect(nextRun).toBe(true);
  });
});
