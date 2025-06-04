import { TestProvider } from "@/test/helpers/test-provider";
import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useNavigate } from "react-router-dom";
import { vi } from "vitest";
import { CreateDatasetDialog } from "./dataset";

const mockOnCreate = vi.fn();
const mockOnUpdate = vi.fn();

describe("CreateDatasetDialog", () => {
  beforeEach(async () => {
    vi.mock("react-router-dom");
    vi.mocked(useNavigate).mockReturnValue(vi.fn());
    render(
      <TestProvider>
        <CreateDatasetDialog
          isOpen={true}
          onClose={() => {}}
          onCreate={mockOnCreate}
          onUpdate={mockOnUpdate}
        />
      </TestProvider>,
    );
  });

  it("should render", () => {
    expect(true).toBe(true);
  });

  it("should enable Create button only when name is provided", async () => {
    const createButton = screen.getByRole("button", { name: "Create" });
    expect(createButton).toBeDisabled();

    const nameInput = screen.getByLabelText("Name");
    await userEvent.type(nameInput, "test-dataset");

    expect(createButton).toBeEnabled();
  });

  it("should call onCreate with correct data for list type dataset", async () => {
    const nameInput = screen.getByLabelText("Name");
    await userEvent.type(nameInput, "test-list-dataset");

    const descriptionInput = screen.getByLabelText("Description");
    await userEvent.type(descriptionInput, "This is a test list dataset.");

    const listTypeRadio = screen.getByLabelText("List");
    await userEvent.click(listTypeRadio);

    const optionsInput = screen.getByLabelText("Options");
    await userEvent.type(optionsInput, "Option 1\nOption 2\nOption 3");

    const createButton = screen.getByRole("button", { name: "Create" });
    await userEvent.click(createButton);

    expect(mockOnCreate).toHaveBeenCalledWith({
      name: "test-list-dataset",
      description: "This is a test list dataset.",
      type: "list",
      options: ["Option 1", "Option 2", "Option 3"],
    });
  });

  it("should call onCreate with correct data for csv type dataset", async () => {
    const nameInput = screen.getByLabelText("Name");
    await userEvent.type(nameInput, "test-csv-dataset");

    const descriptionInput = screen.getByLabelText("Description");
    await userEvent.type(descriptionInput, "This is a test csv dataset.");

    const csvTypeRadio = screen.getByLabelText("CSV");
    await userEvent.click(csvTypeRadio);

    const fileInput = screen.getByLabelText("CSV Files") as HTMLInputElement;
    const testFile1 = new File(["col1,col2\nval1,val2"], "test1.csv", {
      type: "text/csv",
    });
    const testFile2 = new File(["h1,h2\ndata1,data2"], "test2.csv", {
      type: "text/csv",
    });
    await userEvent.upload(fileInput, [testFile1, testFile2]);

    // Check if files are listed (optional, good for debugging)
    expect(screen.getByText(/test1.csv \(\d+\.\d{2} KB\)/)).toBeInTheDocument();
    expect(screen.getByText(/test2.csv \(\d+\.\d{2} KB\)/)).toBeInTheDocument();

    const createButton = screen.getByRole("button", { name: "Create" });
    await userEvent.click(createButton);

    expect(mockOnCreate).toHaveBeenCalledWith({
      name: "test-csv-dataset",
      description: "This is a test csv dataset.",
      type: "csv",
      files: [testFile1, testFile2],
    });
  });
});
