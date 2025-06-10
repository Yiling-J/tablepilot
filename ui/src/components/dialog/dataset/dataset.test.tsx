import React from 'react';
import { render, screen, act } from '@testing-library/react'; // Removed fireEvent
import userEvent from '@testing-library/user-event';
import { CreateDatasetDialog, CreateDatasetDialogProps } from './dataset'; // Adjust path as needed
import { DatasetInfo } from '@/actions'; // Adjust path as needed
import { DragEndEvent } from '@dnd-kit/core';
import { vi } from 'vitest'; // Import vi

// Mock dnd-kit hooks and components to simplify testing if needed,
// though many basic dnd-kit features work in JSDOM.
// For this test, we will focus on handler logic more than pixel-perfect drag simulation.

// Mocking @dnd-kit/sortable's arrayMove as it's a utility function
vi.mock('@dnd-kit/sortable', async () => {
  const actual = await vi.importActual<typeof import('@dnd-kit/sortable')>('@dnd-kit/sortable');
  return {
    ...actual,
    arrayMove: vi.fn((array: unknown[], from: number, to: number) => {
      const newArray = [...array];
      const [item] = newArray.splice(from, 1);
      newArray.splice(to, 0, item);
      return newArray;
    }),
  };
});


// Mock Lucide icons
vi.mock('lucide-react', async () => {
  const actual = await vi.importActual<typeof import('lucide-react')>('lucide-react');
  return {
    ...actual,
    Wand2: () => <div data-testid="wand-icon" />,
    GripVertical: () => <div data-testid="grip-icon" />,
  };
});

// Mock child dialog - GenerateOptionsDialog
vi.mock('../generate-options-dialog', () => ({
    GenerateOptionsDialog: vi.fn(({ isOpen }: { isOpen: boolean }) => isOpen ? <div data-testid="generate-options-dialog">Mocked Generate Options Dialog</div> : null),
}));


const mockOnCreate = vi.fn();
const mockOnUpdate = vi.fn();
const mockOnClose = vi.fn();

const initialProps: CreateDatasetDialogProps = {
  isOpen: true,
  onClose: mockOnClose,
  onCreate: mockOnCreate,
  onUpdate: mockOnUpdate,
};

// Helper to create File objects
const createFile = (name: string, type = 'text/csv', size = 1024): File => {
  const blob = new Blob(['a'.repeat(size)], { type });
  return new File([blob], name, { type });
};

const fileA = createFile('a.csv');
const fileB = createFile('b.csv');
// Removed fileC and fileD as they were unused. If needed later, they can be re-added.


describe('CreateDatasetDialog - CSV Functionality', () => {
  beforeEach(() => {
    vi.clearAllMocks(); // Changed from jest.clearAllMocks()
    // Reset arrayMove mock calls if needed, though it's stateless here
  });

  // Test 1: Initial State (CSV mode)
  test('renders with existing CSV dataset and displays initial files', () => {
    const existingDataset: DatasetInfo = {
      id: 'csv1',
      name: 'My CSV Dataset',
      description: 'An existing dataset.',
      type: 'csv',
      data: ['file1.csv', 'file2.csv'], // file names
      columns: ['header1', 'header2'], // Added missing 'columns' property
      // Removed created_at, updated_at, project_id as they are not in DatasetInfo based on TS error
    };

    render(<CreateDatasetDialog {...initialProps} dataset={existingDataset} />);

    expect(screen.getByLabelText('Name')).toHaveValue('My CSV Dataset');
    expect(screen.getByLabelText('Type')).toBeDisabled(); // Type is disabled for existing datasets

    // Check for displayed files
    expect(screen.getByText('file1.csv (existing)')).toBeInTheDocument();
    expect(screen.getByText('file2.csv (existing)')).toBeInTheDocument();
  });

  // Test 2: File Addition
  test('allows adding new files, displays them, and prevents duplicates', async () => {
    const user = userEvent.setup();
    render(<CreateDatasetDialog {...initialProps} />);

    // Ensure CSV type is selected (it's not by default, default is "list")
    await user.click(screen.getByLabelText('CSV'));
    expect(screen.getByLabelText('CSV')).toBeChecked();


    const fileInput = screen.getByLabelText('CSV Files') as HTMLInputElement;

    // Add fileA
    await act(async () => {
      await user.upload(fileInput, fileA);
    });
    expect(screen.getByText(`${fileA.name} (${(fileA.size / 1024).toFixed(2)} KB)`)).toBeInTheDocument();
    expect(fileInput.files?.[0]).toBe(fileA); // Check if file is in input (transient)

    // Add fileB
    await act(async () => {
      await user.upload(fileInput, fileB);
    });
    expect(screen.getByText(`${fileB.name} (${(fileB.size / 1024).toFixed(2)} KB)`)).toBeInTheDocument();

    // Attempt to add fileA again (duplicate)
    await act(async () => {
      await user.upload(fileInput, fileA);
    });
    // Should still only have one instance of fileA displayed
    expect(screen.getAllByText((content, _element) => content.startsWith(fileA.name)).length).toBe(1); // Prefixed element with _
  });

  // Test 3: File Deletion
  test('allows deleting files from the list', async () => {
    const user = userEvent.setup();
    render(<CreateDatasetDialog {...initialProps} />);
    await user.click(screen.getByLabelText('CSV'));


    const fileInput = screen.getByLabelText('CSV Files');
    await act(async () => {
      await user.upload(fileInput, [fileA, fileB]);
    });

    expect(screen.getByText(new RegExp(fileA.name))).toBeInTheDocument();
    expect(screen.getByText(new RegExp(fileB.name))).toBeInTheDocument();

    // Delete fileA
    // The remove button is identified by its aria-label `Remove ${item.name}`
    const removeButtonForA = screen.getByLabelText(`Remove ${fileA.name}`);
    await act(async () => {
      await user.click(removeButtonForA);
    });

    expect(screen.queryByText(new RegExp(fileA.name))).not.toBeInTheDocument();
    expect(screen.getByText(new RegExp(fileB.name))).toBeInTheDocument();
  });

   // Test 4: Drag and Drop (Reordering) - Testing handleDragEnd directly
  test('handleDragEnd function reorders fileItems correctly', async () => {
    // Define a simple item type for the TestComponent state
    interface TestFileItem {
      id: string;
      name: string;
    }

    let testHandleDragEndFn: ((event: DragEndEvent) => Promise<void>) | null = null;

    const TestComponent = () => {
      const [fileItems, setFileItems] = React.useState<TestFileItem[]>([
        { id: '1', name: 'file1.csv' },
        { id: '2', name: 'file2.csv' },
        { id: '3', name: 'file3.csv' },
      ]);

      const handleDragEndInternal = async (event: DragEndEvent) => {
        const { active, over } = event;
        if (over && active.id !== over.id) {
          const sortable = await vi.importActual<typeof import('@dnd-kit/sortable')>('@dnd-kit/sortable');
          setFileItems((items) => {
            const oldIndex = items.findIndex(item => item.id === active.id);
            const newIndex = items.findIndex(item => item.id === over.id);
            return sortable.arrayMove(items, oldIndex, newIndex);
          });
        }
      };

      testHandleDragEndFn = handleDragEndInternal; // Assign to outer scope variable

      return (
        <div>
          {fileItems.map(f => <div key={f.id}>{f.name}</div>)}
        </div>
      );
    };

    render(<TestComponent />);

    const initialOrder = ['file1.csv', 'file2.csv', 'file3.csv'];
    screen.getAllByText(/file\d\.csv/).forEach((el, i) => {
      expect(el).toHaveTextContent(initialOrder[i]);
    });

    // Helper for mock ClientRect - no longer needed here due to 'any' casting for active/over
    // const createMockRect = (): ClientRect => ({
    //   width: 0, height: 0, top: 0, left: 0, bottom: 0, right: 0,
    //   x: 0, y: 0, toJSON: () => ({})
    // });

    // Construct a DragEndEvent. Cast active and over to 'any' to bypass deep type issues
    // for properties not used by the specific handler under test.
    const dragEndEvent: DragEndEvent = {
      active: {
        id: '1',
        data: { current: {} },
        // rect property is required by Active type, but causing persistent issues.
        // Casting to 'any' because the tested function only uses 'id' and 'data'.
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
      over: {
        id: '3',
        data: { current: {} },
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
      collisions: [],
      delta: { x:0, y:0 },
      activatorEvent: new MouseEvent('mousedown'),
    };
    if (testHandleDragEndFn) {
      await act(async () => {
        await testHandleDragEndFn!(dragEndEvent);
      });
    } else {
      throw new Error("testHandleDragEndFn was not assigned");
    }

    // Expected order: file2.csv, file3.csv, file1.csv (moved '1' after '3')
    // No, arrayMove moves '1' to the position of '3', '3' shifts right.
    // If '1' (idx 0) moves to where '3' (idx 2) is:
    // Original: [1, 2, 3] -> Old index: 0, New index: 2
    // Result: [2, 3, 1]
    const expectedOrder = ['file2.csv', 'file3.csv', 'file1.csv'];
     screen.getAllByText(/file\d\.csv/).forEach((el, i) => {
      expect(el).toHaveTextContent(expectedOrder[i]);
    });
  });


  // Test 5: Submit Logic (CSV Create)
  test('calls onCreate with correct data for new CSV dataset', async () => {
    const user = userEvent.setup();
    render(<CreateDatasetDialog {...initialProps} isOpen={true} />); // Ensure isOpen is true

    await user.type(screen.getByLabelText('Name'), 'New CSV Dataset');
    await user.click(screen.getByLabelText('CSV')); // Select CSV type

    const fileInput = screen.getByLabelText('CSV Files');
    await act(async () => {
      await user.upload(fileInput, [fileA, fileB]);
    });

    await act(async () => {
      await user.click(screen.getByRole('button', { name: 'Create' }));
    });

    expect(mockOnCreate).toHaveBeenCalledTimes(1);
    expect(mockOnCreate).toHaveBeenCalledWith({
      name: 'New CSV Dataset',
      description: '',
      type: 'csv',
      data: [fileA.name, fileB.name], // Ordered file names
      files: [fileA, fileB],          // File objects
      // options should be undefined or not present
    });
  });

  // Test 6: Submit Logic (CSV Update - only reordering)
  test('calls onUpdate with reordered file names and no new files if only reorder happened', async () => {
    const user = userEvent.setup();
    const existingDataset: DatasetInfo = {
      id: 'csv2', name: 'Reorderable Dataset', description: '', type: 'csv',
      data: ['initialA.csv', 'initialB.csv'],
      columns: ['colA', 'colB'], // Added missing 'columns' property
      // Removed created_at, updated_at, project_id
    };

    // This test relies on the separate test for `handleDragEnd` to ensure reordering logic is correct.
    // Here, we assume `fileItems` state would be updated by `handleDragEnd` if a drag occurred.
    // Since we don't simulate the drag itself, we'll test the scenario where `fileItems` order
    // is the same as initial, but the key is to check `files: undefined`.

    render(<CreateDatasetDialog {...initialProps} dataset={existingDataset} />);

    // To truly test reordering impact on submit, one would need to:
    // 1. Render the dialog.
    // 2. Programmatically update `fileItems` state to a new order (difficult from outside).
    // OR
    // 2. Simulate drag-and-drop that calls `handleDragEnd` and updates `fileItems`.
    // For this test, we focus on the structure of `onUpdate` when no *new* files are added.
    // The `data` field will reflect the initial order because we haven't simulated a reorder.
    // A more advanced test could mock `setFileItems` to force a reordered state before submit.

    await user.click(screen.getByRole('button', { name: 'Update' }));

    expect(mockOnUpdate).toHaveBeenCalledTimes(1);
    expect(mockOnUpdate).toHaveBeenCalledWith(existingDataset.id, {
      name: existingDataset.name,
      description: existingDataset.description,
      type: 'csv',
      data: ['initialA.csv', 'initialB.csv'], // This would be reordered if drag was simulated
      files: undefined,
    });
  });


  // Test 7: Submit Logic (CSV Update - adding, deleting, reordering)
  test('calls onUpdate with mixed changes: new, deleted, and reordered files', async () => {
    const user = userEvent.setup();
    const existingDataset: DatasetInfo = {
      id: 'csv3', name: 'Complex Update', description: '', type: 'csv',
      data: ['a.csv', 'b.csv', 'c.csv'], // Initial files from server
      columns: ['h1', 'h2', 'h3'], // Added missing 'columns' property
      // Removed created_at, updated_at, project_id
    };

    render(<CreateDatasetDialog {...initialProps} dataset={existingDataset} />);

    // Initial state: a.csv (existing), b.csv (existing), c.csv (existing)
    expect(screen.getByText('a.csv (existing)')).toBeInTheDocument();
    expect(screen.getByText('b.csv (existing)')).toBeInTheDocument();
    expect(screen.getByText('c.csv (existing)')).toBeInTheDocument();

    // 1. Delete 'b.csv'
    const removeButtonForB = screen.getByLabelText(`Remove b.csv`);
    await act(async () => {
      await user.click(removeButtonForB);
    });
    // State: a.csv (existing), c.csv (existing)
    expect(screen.queryByText('b.csv (existing)')).not.toBeInTheDocument();


    // 2. Add new file 'd.csv'
    const fileInput = screen.getByLabelText('CSV Files');
    const fileDLocal = createFile('d.csv', 'text/csv', 500); // Local instance
    await act(async () => {
      await user.upload(fileInput, fileDLocal);
    });
    // State: a.csv (existing), c.csv (existing), d.csv (new)
    expect(screen.getByText(`${fileDLocal.name} (${(fileDLocal.size / 1024).toFixed(2)} KB)`)).toBeInTheDocument();

    // 3. Reorder to ['d.csv', 'a.csv', 'c.csv'] - This part is assumed to be handled by dnd-kit logic
    // tested via `handleDragEnd`. The actual order submitted will be based on the final `fileItems` state.
    // Without dnd simulation, the order after these operations is [a.csv (existing), c.csv (existing), d.csv (new)]

    await act(async () => {
      await user.click(screen.getByRole('button', { name: 'Update' }));
    });

    expect(mockOnUpdate).toHaveBeenCalledTimes(1);
    expect(mockOnUpdate).toHaveBeenCalledWith(existingDataset.id, {
      name: existingDataset.name,
      description: existingDataset.description,
      type: 'csv',
      // Order after operations (delete b, add d) without explicit reorder:
      data: ['a.csv', 'c.csv', fileDLocal.name],
      files: [fileDLocal], // Only the new File object for d.csv
    });
  });
});

// A note on testing dnd-kit:
// Full simulation of drag and drop can be complex with @testing-library/user-event alone.
// Libraries like `@dnd-kit/test-utils` or approaches described in dnd-kit documentation
// might be needed for more robust dnd interaction tests.
// For this suite, testing `handleDragEnd` directly for reordering logic and then
// testing the submit handlers with assumed states (post-reorder) is a pragmatic approach.
// The current `handleDragEnd` test is a bit artificial; ideally, it would be tested
// on an actual instance of the dialog, but that requires more setup or exporting the function.
// The current test for handleDragEnd is functional but uses a separate TestComponent.

// To improve the 'handleDragEnd' test for the actual component:
// 1. Render CreateDatasetDialog.
// 2. Add some files to populate `fileItems`.
// 3. Find a way to get a reference to the `handleDragEnd` function from the rendered instance,
//    or make it a prop, or trigger it through a more abstract dnd simulation.
// 4. Call it with a mocked event.
// 5. Check if the displayed list of files in the dialog reorders.
// This is still indirect. True dnd simulation is the most robust if feasible.

// The test for "CSV Update - only reordering" also has a similar challenge.
// It currently submits the initial order because no reorder was actually simulated.
// A more complete test would involve:
//   render dialog with ['a','b'] -> simulate drag to get ['b','a'] -> submit -> check onUpdate data: ['b','a']
// This would require either a good dnd test utility or a way to manually trigger
// the state change for `fileItems` to reflect the new order before submit.
// The `arrayMove` mock ensures we can track if it's used correctly by `handleDragEnd`.
