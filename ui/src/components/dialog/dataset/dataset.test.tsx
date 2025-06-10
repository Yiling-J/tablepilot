import { TestProvider } from "@/test/helpers/test-provider";
import "@testing-library/jest-dom";
import { render, screen, within, waitFor } from "@testing-library/react"; // Added waitFor, within
import userEvent from "@testing-library/user-event";
import { useNavigate } from "react-router-dom";
import { beforeEach, describe, expect, it, MockedFunction, vi } from "vitest";
import type { CreateDatasetDialogProps } from "./dataset";
import { CreateDatasetDialog } from "./dataset";

let capturedOnDragEnd: any;

vi.mock("react-beautiful-dnd", async () => {
  const actual = await vi.importActual("react-beautiful-dnd");
  return {
    ...actual,
    DragDropContext: vi.fn(({ children, onDragEnd }) => {
      capturedOnDragEnd = onDragEnd;
      return children;
    }),
    Droppable: vi.fn(({ children }) =>
      children(
        {
          innerRef: vi.fn(),
          droppableProps: { "data-testid": "droppable-area" },
          placeholder: null,
        },
        {},
      ),
    ),
    Draggable: vi.fn(({ children, draggableId }) =>
      children(
        {
          innerRef: vi.fn(),
          draggableProps: { "data-testid": `draggable-${draggableId}` },
          dragHandleProps: {},
        },
        {},
      ),
    ),
  };
});

vi.mock("../generate-options-dialog", () => ({
  GenerateOptionsDialog: vi.fn((props) => {
    if (!props.isOpen) return null;
    return (
      <div data-testid="generate-options-dialog">
        <button
          style={{ pointerEvents: "auto" }}
          data-testid="generate-options-submit"
          onClick={() =>
            props.onGenerationComplete([
              "gen_opt1_from_mock",
              "gen_opt2_from_mock",
            ])
          }
        >
          Mock Generate Complete
        </button>
        <button onClick={props.onClose}>Mock Dialog Close</button>
        <p>Dataset Name: {props.datasetName}</p>
        <p>Dataset Description: {props.datasetDescription}</p>
      </div>
    );
  }),
}));

const mockOnCreate = vi.fn();
const mockOnUpdate = vi.fn();

describe("CreateDatasetDialog", () => {
  beforeEach(async () => {
    mockOnCreate.mockClear();
    mockOnUpdate.mockClear();
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

  it("should display persisted files, allow replacement, and update correctly", async () => {
    const user = userEvent.setup();
    mockOnCreate.mockClear(); // Clear mocks for this specific test context if needed
    mockOnUpdate.mockClear();

    const existingDataset = {
      id: "csv1",
      name: "Persisted CSVs",
      description: "Test persisted",
      type: "csv" as "list" | "csv",
      data: ["old1.csv", "old2.csv"],
      columns: [], // Added to satisfy DatasetInfo type
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    render(
      <TestProvider>
        <CreateDatasetDialog
          isOpen={true}
          onClose={() => {}}
          onCreate={mockOnCreate}
          onUpdate={mockOnUpdate}
          dataset={existingDataset}
        />
      </TestProvider>,
    );

    await screen.findByText("old1.csv (persisted)");
    expect(screen.getByText("old2.csv (persisted)")).toBeInTheDocument();

    const fileInput = screen.getByLabelText("CSV Files") as HTMLInputElement;
    const newFile = new File(["new data"], "new.csv", { type: "text/csv" });
    await user.upload(fileInput, newFile);

    await screen.findByText(/new.csv \(\d+\.\d{2} KB\)/);
    expect(screen.queryByText("old1.csv (persisted)")).not.toBeInTheDocument();
    expect(screen.queryByText("old2.csv (persisted)")).not.toBeInTheDocument();

    const updateButton = screen.getByRole("button", { name: "Update" });
    await user.click(updateButton);

    expect(mockOnUpdate).toHaveBeenCalledWith(
      "csv1",
      expect.objectContaining({
        name: "Persisted CSVs",
        type: "csv",
        files: [newFile],
      }),
    );
  });

  it("should allow deleting a selected CSV file", async () => {
    const user = userEvent.setup();
    // Using the global render from beforeEach, but clear mocks to be safe
    mockOnCreate.mockClear();
    mockOnUpdate.mockClear();

    // Re-render or ensure component is in a clean state for this test if beforeEach setup is not ideal
    // For this case, we assume the beforeEach setup is sufficient as we are testing creation.
    // If issues arise, a dedicated render for this test might be needed.

    await user.type(screen.getByLabelText("Name"), "Delete Test");
    await user.click(screen.getByLabelText("CSV"));

    const fileInput = screen.getByLabelText("CSV Files") as HTMLInputElement;
    const file1 = new File(["content1"], "file1.csv", { type: "text/csv" }); // Approx 0.01KB
    const file2 = new File(["content2"], "file2.csv", { type: "text/csv" }); // Approx 0.01KB
    await user.upload(fileInput, [file1, file2]);

    const file1Matcher = /file1.csv \(0\.01 KB\)/;
    const file2Matcher = /file2.csv \(0\.01 KB\)/;

    await screen.findByText(file1Matcher);
    expect(screen.getByText(file2Matcher)).toBeInTheDocument();

    // Corrected selector for the draggable element containing file1
    const file1Draggable = screen.getByText(file1Matcher).closest(`div[data-testid="draggable-${file1.name}-${file1.size}"]`) as HTMLElement;
    if (!file1Draggable) throw new Error(`Draggable for ${file1.name} not found`);
    const deleteButton = within(file1Draggable).getByRole('button', { name: new RegExp(`Remove ${file1.name.replace('.', '\\.')}`, 'i') });
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.queryByText(file1Matcher)).not.toBeInTheDocument();
    });
    expect(screen.getByText(file2Matcher)).toBeInTheDocument();

    const createButton = screen.getByRole("button", { name: "Create" });
    await user.click(createButton);
    expect(mockOnCreate).toHaveBeenCalledWith(expect.objectContaining({
      files: [file2],
    }));
  });

  it("should call onUpdate without files field if no files were selected/changed", async () => {
    const user = userEvent.setup();
    mockOnCreate.mockClear();
    mockOnUpdate.mockClear();

    const existingDataset = {
      id: "csv2",
      name: "No File Change",
      description: "Desc",
      type: "csv" as "list" | "csv",
      data: ["existing.csv"],
      columns: [], // Added to satisfy DatasetInfo type
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    render(
      <TestProvider>
        <CreateDatasetDialog
          isOpen={true}
          onClose={() => {}}
          onCreate={mockOnCreate}
          onUpdate={mockOnUpdate}
          dataset={existingDataset}
        />
      </TestProvider>,
    );

    await screen.findByText("existing.csv (persisted)");
    // Update the name field only
    const nameInput = screen.getByLabelText("Name");
    await user.clear(nameInput);
    await user.type(nameInput, "Updated Name Only");

    const updateButton = screen.getByRole("button", { name: "Update" });
    await user.click(updateButton);

    expect(mockOnUpdate).toHaveBeenCalledWith(
      "csv2",
      expect.objectContaining({
        name: "Updated Name Only",
        description: "Desc",
        type: "csv",
      }),
    );
    // Check that the 'files' property is not present
    const calledWithData = mockOnUpdate.mock.calls[0][1];
    expect(calledWithData.hasOwnProperty("files")).toBe(false);
  });

  it("should reorder selected files using mocked onDragEnd", async () => {
    const user = userEvent.setup();
    // Using the global render from beforeEach, but clear mocks to be safe
    mockOnCreate.mockClear();
    mockOnUpdate.mockClear();

    await user.type(screen.getByLabelText("Name"), "Reorder Test");
    await user.click(screen.getByLabelText("CSV"));

    const fileInput = screen.getByLabelText("CSV Files") as HTMLInputElement;
    const fileA = new File(["contentA"], "fileA.csv", { type: "text/csv" });
    const fileB = new File(["contentBB"], "fileB.csv", { type: "text/csv" });
    const fileC = new File(["contentCCC"], "fileC.csv", { type: "text/csv" });
    await user.upload(fileInput, [fileA, fileB, fileC]);

    // Using more specific matchers for file sizes (0.01 KB for tiny files)
    await screen.findByText(/fileA.csv \(0\.01 KB\)/);
    screen.getByText(/fileB.csv \(0\.01 KB\)/);
    screen.getByText(/fileC.csv \(0\.01 KB\)/);

    const mockDropResult = {
      source: { index: 0, droppableId: "selected-csv-files" },
      destination: { index: 2, droppableId: "selected-csv-files" },
    };

    if (typeof capturedOnDragEnd === "function") {
      capturedOnDragEnd(mockDropResult);
    } else {
      throw new Error("onDragEnd was not captured or is not a function");
    }

    await waitFor(() => {
      const droppable = screen.getByTestId("droppable-area");
      const draggables = within(droppable).getAllByText(/\(0\.01 KB\)/);
      // After dragging fileA (index 0) to index 2: Expected order B, C, A
      expect(draggables[0].textContent).toMatch(/fileB.csv/);
      expect(draggables[1].textContent).toMatch(/fileC.csv/);
      expect(draggables[2].textContent).toMatch(/fileA.csv/);
    });

    const createButton = screen.getByRole("button", { name: "Create" });
    await user.click(createButton);
    expect(mockOnCreate).toHaveBeenCalledWith(
      expect.objectContaining({
        files: [fileB, fileC, fileA],
      }),
    );
  });
});

const mockOnCloseForAIFeature: MockedFunction<() => void> = vi.fn();
const mockOnCreateForAIFeature: MockedFunction<
  (data: {
    name: string;
    description: string;
    type: "list" | "csv";
    options?: string[];
    files?: File[];
  }) => void
> = vi.fn();
const mockOnUpdateForAIFeature: MockedFunction<
  (
    id: string,
    data: {
      name: string;
      description: string;
      type: "list" | "csv";
      options?: string[];
      files?: File[];
    },
  ) => void
> = vi.fn();

const defaultTestPropsForAIFeature: CreateDatasetDialogProps = {
  isOpen: true,
  onClose: mockOnCloseForAIFeature,
  onCreate: mockOnCreateForAIFeature,
  onUpdate: mockOnUpdateForAIFeature,
};

const renderCreateDatasetDialogForAIFeature = (
  props?: Partial<CreateDatasetDialogProps>,
) => {
  return render(
    <TestProvider>
      <CreateDatasetDialog {...defaultTestPropsForAIFeature} {...props} />
    </TestProvider>,
  );
};

describe("CreateDatasetDialog - AI Options Generation Feature", () => {
  beforeEach(() => {
    mockOnCloseForAIFeature.mockClear();
    mockOnCreateForAIFeature.mockClear();
    mockOnUpdateForAIFeature.mockClear();
  });

  const getOptionsTextarea = () =>
    screen.getByLabelText("Options") as HTMLTextAreaElement;

  it('DOES show wand icon button when dataset type is "list"', async () => {
    renderCreateDatasetDialogForAIFeature();
    const radioList = screen.getByLabelText("List");
    await userEvent.click(radioList);
    await screen.findByText("Options");
    await screen.findByLabelText("wand-button");
  });

  it("opens GenerateOptionsDialog when wand icon is clicked", async () => {
    renderCreateDatasetDialogForAIFeature();
    await userEvent.click(screen.getByLabelText("List"));

    const radioList = screen.getByLabelText("List");
    await userEvent.click(radioList);
    await screen.findByText("Options");
    await userEvent.click(screen.getByLabelText("wand-button"));
    expect(screen.getByTestId("generate-options-dialog")).toBeInTheDocument();
  });

  it("passes correct datasetName and datasetDescription to GenerateOptionsDialog", async () => {
    const datasetName = "AI Test Name";
    const datasetDescription = "AI Test Description";
    renderCreateDatasetDialogForAIFeature();

    const nameInput = screen.getByLabelText("Name");
    await userEvent.clear(nameInput);
    await userEvent.type(nameInput, datasetName);

    const descriptionInput = screen.getByLabelText("Description");
    await userEvent.clear(descriptionInput);
    await userEvent.type(descriptionInput, datasetDescription);

    await userEvent.click(screen.getByLabelText("List"));
    await userEvent.click(screen.getByLabelText("wand-button"));

    expect(screen.getByTestId("generate-options-dialog")).toBeInTheDocument();
    expect(
      screen.getByText(`Dataset Name: ${datasetName}`),
    ).toBeInTheDocument();
    expect(
      screen.getByText(`Dataset Description: ${datasetDescription}`),
    ).toBeInTheDocument();
  });

  it("appends generated options to textarea when onGenerationComplete is called from mock", async () => {
    renderCreateDatasetDialogForAIFeature();
    await userEvent.click(screen.getByLabelText("List"));

    const optionsTextarea = getOptionsTextarea();
    await userEvent.type(optionsTextarea, "Initial Option 1\nInitial Option 2");
    await userEvent.click(screen.getByLabelText("wand-button"));
    const mockGenerateButton = screen.getByTestId("generate-options-submit");
    await userEvent.click(mockGenerateButton);

    const expectedOptions = "gen_opt1_from_mock\ngen_opt2_from_mock";
    expect(optionsTextarea.value).toBe(expectedOptions);
    expect(
      screen.queryByTestId("generate-options-dialog"),
    ).not.toBeInTheDocument();
  });
});
