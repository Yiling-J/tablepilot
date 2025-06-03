// React import removed as it's not strictly needed for modern Vitest/RTL tests unless directly using React APIs like useState, etc.
import { render, screen, waitFor } from '@testing-library/react'; // within import removed as it's unused
// Explicitly import vitest functions to satisfy tsc, as global types might not be picked up by tsc alone.
import { vi, describe, it, expect, beforeEach, type Mock } from 'vitest'; // Added type Mock
import { DatasetPreviewDialog } from './preview'; // Adjust path as needed
import { previewDataset as mockPreviewDataset } from '@/actions';

// Mock the actions module
vi.mock('@/actions', async () => { // Corrected syntax: removed underscore
  const actual = await vi.importActual('@/actions');
  return {
    ...actual,
    previewDataset: vi.fn(),
  };
});

const mockOnClose = vi.fn();

const defaultProps = {
  isOpen: true,
  onClose: mockOnClose,
  datasetId: 'test-dataset-id',
};

// Helper to render the component
const renderDialog = (props = {}) => {
  return render(<DatasetPreviewDialog {...defaultProps} {...props} />);
};

describe('DatasetPreviewDialog', () => { // Corrected syntax
  beforeEach(() => { // Corrected syntax
    vi.resetAllMocks(); // Reset mocks before each test
    (mockPreviewDataset as Mock).mockClear(); // Changed vi.Mock to Mock
    mockOnClose.mockClear();
  });

  it('shows loading state initially', async () => { // Corrected syntax
    (mockPreviewDataset as Mock).mockReturnValue(new Promise(() => {})); // Changed vi.Mock to Mock, Corrected syntax, Promise that never resolves
    renderDialog();
    // DialogContent has min-h-[200px], Skeletons are rendered inside that.
    // Skeletons don't have implicit roles, so we check for their presence by structure or testId if added.
    // For now, let's assume the presence of multiple skeleton divs indicates loading.
    // A more robust way would be to add test-ids to Skeleton components or check for "Loading..." text if it existed.
    // The current implementation uses Skeleton components.
    // Let's check for the DialogTitle as a proxy for the dialog being open.
    expect(screen.getByText('Dataset Preview')).toBeInTheDocument();
    // And then check for skeleton elements (though this is a bit fragile)
    // Querying by class name is not ideal with testing-library, but given Skeleton structure:
    const skeletons = document.querySelectorAll('.animate-pulse'); // Default class for Skeleton
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('shows error state if fetching fails', async () => { // Corrected syntax
    const errorMessage = 'Failed to fetch dataset preview.';
    (mockPreviewDataset as Mock).mockRejectedValue(new Error(errorMessage)); // Changed vi.Mock to Mock
    renderDialog();
    await waitFor(() => { // Corrected syntax
      expect(screen.getByText(`Error: ${errorMessage}`)).toBeInTheDocument();
    });
  });

  describe('List View', () => { // Corrected syntax
    const listData = {
      type: 'list' as const,
      data: ['item1', 'item2', 'item3'],
      // rows: undefined, // or [] depending on actual API response for list type
    };

    it('renders list items as badges and no table', async () => { // Corrected syntax
      (mockPreviewDataset as Mock).mockResolvedValue(listData); // Changed vi.Mock to Mock
      renderDialog();

      await waitFor(() => { // Corrected syntax
        expect(screen.getByText('Dataset Preview')).toBeInTheDocument();
      });

      // Check for badges
      expect(screen.getByText('item1')).toBeInTheDocument();
      expect(screen.getByText('item2')).toBeInTheDocument();
      expect(screen.getByText('item3')).toBeInTheDocument();
      // Check parent of item1 has badge classes (example)
      expect(screen.getByText('item1').className).toContain('bg-blue-100');


      // Check no table elements are rendered
      expect(screen.queryByRole('table')).toBeNull();
    });
  });

  describe('CSV View', () => { // Corrected syntax
    const csvData = {
      type: 'csv' as const,
      rows: [
        { col1: 'val1a', col2: 'val2a' },
        { col1: 'val1b', col2: 'val2b' },
      ],
      // data: undefined, // or []
    };

    it('renders table with headers and cells, and no badges', async () => { // Corrected syntax
      (mockPreviewDataset as Mock).mockResolvedValue(csvData); // Changed vi.Mock to Mock
      renderDialog();

      await waitFor(() => { // Corrected syntax
        expect(screen.getByText('Dataset Preview')).toBeInTheDocument();
      });

      // Check for table headers
      expect(screen.getByRole('columnheader', { name: 'col1' })).toBeInTheDocument();
      expect(screen.getByRole('columnheader', { name: 'col2' })).toBeInTheDocument();

      // Check for table cells
      expect(screen.getByRole('cell', { name: 'val1a' })).toBeInTheDocument();
      expect(screen.getByRole('cell', { name: 'val2a' })).toBeInTheDocument();
      expect(screen.getByRole('cell', { name: 'val1b' })).toBeInTheDocument();
      expect(screen.getByRole('cell', { name: 'val2b' })).toBeInTheDocument();

      // Check no list items (badges) are rendered
      // Assuming list items would have specific text not present in CSV data or unique badge role/styling
      expect(screen.queryByText('item1')).toBeNull(); // Example, if list items were named 'itemX'
      const badges = document.querySelectorAll('.bg-blue-100'); // Badge specific class
      expect(badges.length).toBe(0);
    });

    it('renders message if CSV has no rows', async () => { // Corrected syntax
      const emptyCsvData = {
        type: 'csv' as const,
        rows: [],
        // data: undefined,
      };
      (mockPreviewDataset as Mock).mockResolvedValue(emptyCsvData); // Changed vi.Mock to Mock
      renderDialog();

      await waitFor(() => { // Corrected syntax
        expect(screen.getByText('CSV has no rows to display.')).toBeInTheDocument();
      });
    });
  });

  it('calls onClose when Close button is clicked', async () => { // Corrected syntax
    // Mock a successful load (e.g., list type) to ensure dialog content is stable
    (mockPreviewDataset as Mock).mockResolvedValue({ type: 'list', data: ['test'] }); // Changed vi.Mock to Mock
    renderDialog();

    await waitFor(() => { // Corrected syntax
      expect(screen.getByText('test')).toBeInTheDocument(); // Wait for content
    });

    // Get all buttons named "Close"
    const closeButtons = screen.getAllByRole('button', { name: 'Close' });
    // The button in the footer does not have an SVG child, unlike the Radix 'X' button.
    // This is a bit fragile; a data-testid on the footer button would be better.
    const footerCloseButton = closeButtons.find(
      (button) => !button.querySelector('svg') && button.textContent === "Close"
    );

    if (!footerCloseButton) {
      throw new Error("Could not find the footer close button.");
    }
    footerCloseButton.click();

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });
});
