import { TestProvider } from "@/test/helpers/test-provider";
import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react"; // Added waitFor
import userEvent from "@testing-library/user-event";
import { useNavigate } from "react-router-dom";
import { beforeEach, describe, expect, it, MockedFunction, vi } from "vitest";
import type { CreateDatasetDialogProps } from "./dataset";
import { CreateDatasetDialog } from "./dataset";

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
