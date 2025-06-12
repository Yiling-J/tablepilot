import { previewDataset as mockPreviewDataset } from "@/actions";
import { render, screen, waitFor } from "@testing-library/react"; // within import removed as it's unused
import { beforeEach, describe, expect, it, vi, type Mock } from "vitest"; // Added type Mock
import { DatasetPreviewDialog } from "./preview"; // Adjust path as needed

vi.mock("@/actions", async () => {
  const actual = await vi.importActual("@/actions");
  return {
    ...actual,
    previewDataset: vi.fn(),
  };
});

const mockOnClose = vi.fn();

const defaultProps = {
  isOpen: true,
  onClose: mockOnClose,
  datasetId: "test-dataset-id",
};

const renderDialog = (props = {}) => {
  return render(<DatasetPreviewDialog {...defaultProps} {...props} />);
};

describe("DatasetPreviewDialog", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    (mockPreviewDataset as Mock).mockClear();
    mockOnClose.mockClear();
  });

  it("shows loading state initially", async () => {
    (mockPreviewDataset as Mock).mockReturnValue(new Promise(() => {}));
    renderDialog();
    expect(screen.getByText("Dataset Preview")).toBeInTheDocument();
    const skeletons = document.querySelectorAll(".animate-pulse");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("shows error state if fetching fails", async () => {
    const errorMessage = "Failed to fetch dataset preview.";
    (mockPreviewDataset as Mock).mockRejectedValue(new Error(errorMessage));
    renderDialog();
    await waitFor(() => {
      // Corrected syntax
      expect(screen.getByText(`Error: ${errorMessage}`)).toBeInTheDocument();
    });
  });

  describe("List View", () => {
    const listData = {
      type: "list" as const,
      data: ["item1", "item2", "item3"],
    };

    it("renders list items as badges and no table", async () => {
      (mockPreviewDataset as Mock).mockResolvedValue(listData);
      renderDialog();
      await screen.findByText("Dataset Preview");
      await screen.findByText("item1");

      expect(screen.getByText("item1")).toBeInTheDocument();
      expect(screen.getByText("item2")).toBeInTheDocument();
      expect(screen.getByText("item3")).toBeInTheDocument();
      expect(screen.getByText("item1").className).toContain("bg-blue-100");

      expect(screen.queryByRole("table")).toBeNull();
    });
  });

  describe("CSV View", () => {
    const csvData = {
      type: "csv" as const,
      rows: [
        { col1: "val1a", col2: "val2a" },
        { col1: "val1b", col2: "val2b" },
      ],
    };

    it("renders table with headers and cells, and no badges", async () => {
      (mockPreviewDataset as Mock).mockResolvedValue(csvData);
      renderDialog();

      await screen.findByText("Dataset Preview");
      await screen.findByRole("columnheader", { name: "col1" });
      expect(
        screen.getByRole("columnheader", { name: "col2" }),
      ).toBeInTheDocument();

      expect(screen.getByRole("cell", { name: "val1a" })).toBeInTheDocument();
      expect(screen.getByRole("cell", { name: "val2a" })).toBeInTheDocument();
      expect(screen.getByRole("cell", { name: "val1b" })).toBeInTheDocument();
      expect(screen.getByRole("cell", { name: "val2b" })).toBeInTheDocument();

      expect(screen.queryByText("item1")).toBeNull();
      const badges = document.querySelectorAll(".bg-blue-100");
      expect(badges.length).toBe(0);
    });

    it("renders message if CSV has no rows", async () => {
      const emptyCsvData = {
        type: "csv" as const,
        rows: [],
      };
      (mockPreviewDataset as Mock).mockResolvedValue(emptyCsvData);
      renderDialog();

      await waitFor(() => {
        // Corrected syntax
        expect(
          screen.getByText("CSV has no rows to display."),
        ).toBeInTheDocument();
      });
    });
  });

  it("calls onClose when Close button is clicked", async () => {
    (mockPreviewDataset as Mock).mockResolvedValue({
      type: "list",
      data: ["test"],
    });
    renderDialog();

    await waitFor(() => {
      expect(screen.getByText("test")).toBeInTheDocument();
    });

    const closeButtons = screen.getAllByRole("button", { name: "Close" });
    const footerCloseButton = closeButtons.find(
      (button) =>
        !button.querySelector("svg") && button.textContent === "Close",
    );

    if (!footerCloseButton) {
      throw new Error("Could not find the footer close button.");
    }
    footerCloseButton.click();

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  describe("Image Dataset Preview", () => {
    const imageBase64 =
      "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA AAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO 9TXL0Y4OHwAAAABJRU5ErkJggg==";
    const imageUrl = "https://example.com/image1.jpg";

    it("renders images correctly from URLs and base64 strings", async () => {
      const imageData = {
        type: "image" as const,
        data: [imageUrl, imageBase64],
      };
      (mockPreviewDataset as Mock).mockResolvedValue(imageData);
      renderDialog();

      await waitFor(() => {
        expect(screen.getAllByRole("img")).toHaveLength(2);
      });

      const images = screen.getAllByRole("img");
      expect(images[0]).toHaveAttribute("src", imageUrl);
      expect(images[0]).toHaveAttribute("alt", "Preview 1");
      expect(screen.getByText("image1.jpg")).toBeInTheDocument();

      expect(images[1]).toHaveAttribute("src", imageBase64);
      expect(images[1]).toHaveAttribute("alt", "Preview 2");
      expect(screen.getByText("Image 2")).toBeInTheDocument(); // Generic name for base64
    });

    it("handles empty image list", async () => {
      const emptyImageData = {
        type: "image" as const,
        data: [],
      };
      (mockPreviewDataset as Mock).mockResolvedValue(emptyImageData);
      renderDialog();

      await waitFor(() => {
        expect(
          screen.getByText(
            "Preview for this data type is not yet implemented or the data is empty/invalid.",
          ),
        ).toBeInTheDocument();
      });
      expect(screen.queryAllByRole("img")).toHaveLength(0);
    });

    it("shows loading state for image preview", async () => {
      (mockPreviewDataset as Mock).mockReturnValue(new Promise(() => {})); // Never resolves
      renderDialog();
      // Check for DialogTitle to ensure dialog is attempting to render
      expect(screen.getByText("Dataset Preview")).toBeInTheDocument();
      // Check for presence of skeleton loaders (assuming they use 'animate-pulse')
      const skeletons = document.querySelectorAll(".animate-pulse");
      expect(skeletons.length).toBeGreaterThan(0); // Or a more specific count if known
    });

    it("shows error state if fetching image data fails", async () => {
      const errorMessage = "Failed to fetch image dataset.";
      (mockPreviewDataset as Mock).mockRejectedValue(new Error(errorMessage));
      renderDialog();
      await waitFor(() => {
        expect(screen.getByText(`Error: ${errorMessage}`)).toBeInTheDocument();
      });
    });
  });
});
