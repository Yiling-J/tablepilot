import {
    getModels,
    getProviders,
    importImage,
    Provider,
    TableCreateRequest,
} from "@/actions";
import { JSONObject } from "@/json";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useNavigate } from "react-router-dom";
import { Mock } from "vitest";
import { ImportFileDialog } from "./import-file";

describe("ImportFile", () => {
  beforeEach(() => {
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
    const mockedGetProviders = vi.mocked(getProviders);
    mockedGetProviders.mockResolvedValue([
      {
        id: 1,
        name: "p",
        type: "openai",
        models: [{ model: "ai" }],
      } as Provider,
    ]);
  });
  it("should parse the csv file", async () => {
    vi.mock("react-router-dom");
    vi.mocked(useNavigate).mockReturnValue(vi.fn());
    let nextRun = false;
    const onNext = (form: TableCreateRequest, rows: JSONObject[]) => {
      expect(form).toMatchObject({
        name: "users",
        description: "",
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
      <ImportFileDialog
        isOpen={true}
        setIsOpen={vi.fn()}
        onNext={onNext}
        tables={[]}
      />,
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

  it("should import image file", async () => {
    vi.mock("react-router-dom");
    const nv = vi.mocked(useNavigate);
    nv.mockReturnValue(vi.fn());
    let nextRun = false;
    const onNext = () => {
      nextRun = true;
    };
    const mockedImportImage = vi.mocked(importImage);
    mockedImportImage.mockImplementation(async (req) => {
      expect(req.model).toBe("bi");
      expect(req.prompt).toBe("test");
      expect(req.data).toBe("Zm9vYmFy");
      return "foobar";
    });
    render(
      <ImportFileDialog
        isOpen={true}
        setIsOpen={vi.fn()}
        onNext={onNext}
        tables={[]}
      />,
    );
    const fileContent = "foobar";
    const blob = new Blob([fileContent], { type: "image/png" });
    const file = new File([blob], "users.png", { type: "image/png" });
    await userEvent.upload(screen.getByTestId("import-file-selector"), file);
    expect(screen.getByText("users.png")).toBeInTheDocument();
    await userEvent.click(screen.getByText("ai").closest("button")!);
    await userEvent.click(screen.getByText("bi"));
    expect(screen.getByText("Prompt")).toBeInTheDocument();
    const input = screen.getByPlaceholderText(
      "Provide specific instructions to guide the AI in extracting the right table from your image.",
    );
    await userEvent.click(input);
    await userEvent.keyboard("test");
    await userEvent.click(screen.getByText("Next"));
    expect(nextRun).toBe(false);
    expect(mockedImportImage).toBeCalled();
    expect(nv).toBeCalled();
    expect((nv.mock.results[0].value as Mock).mock.calls[0][0]).toBe(
      "/tables/foobar",
    );
  });
});
