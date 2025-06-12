import { render, fireEvent, screen, act } from '@testing-library/react'; // Removed waitFor
import { vi, describe, test, expect, beforeEach, Mock, afterEach } from 'vitest';
import { ReactNode } from 'react';
import { CreateDatasetDialog, CreateDatasetDialogProps } from './dataset';
import { DatasetInfo } from '@/actions';
import { DragEndEvent, Active, Over, UniqueIdentifier } from '@dnd-kit/core';

vi.mock('@/actions', () => ({}));
vi.mock('@/urls', () => ({
  imageUrl: (path: string) => `mock://${path}`,
}));

type MockDatasetInfo = {
  id: string;
  name: string;
  description: string;
  type: "list" | "csv" | "image";
  data: string[];
  columns: string[];
};

let dndOnDragEnd: ((event: DragEndEvent) => void) | undefined = undefined;

vi.mock('@dnd-kit/core', async () => {
  const actual = await vi.importActual('@dnd-kit/core');
  return {
    ...actual,
    DndContext: ({ children, onDragEnd }: { children: ReactNode, onDragEnd?: (event: DragEndEvent) => void }) => {
      dndOnDragEnd = onDragEnd;
      return <div data-testid="dnd-context">{children}</div>;
    },
    useSensor: vi.fn(),
    useSensors: vi.fn(),
    PointerSensor: vi.fn(),
    KeyboardSensor: vi.fn(),
    closestCenter: vi.fn(),
  };
});

vi.mock('@dnd-kit/sortable', async () => {
  const actual = await vi.importActual('@dnd-kit/sortable');
  return {
    ...actual,
    SortableContext: ({ children }: { children: ReactNode }) => <div data-testid="sortable-context">{children}</div>,
    useSortable: ({ id }: { id: string }) => ({
      attributes: { role: 'button', 'aria-roledescription': 'sortable', 'data-sortable-id': id },
      listeners: { onMouseDown: vi.fn(), onKeyDown: vi.fn() },
      setNodeRef: vi.fn(),
      transform: null,
      transition: null,
      isDragging: false,
    }),
    arrayMove: vi.fn((arr, from, to) => {
      const newArray = [...arr];
      const element = newArray.splice(from, 1)[0];
      newArray.splice(to, 0, element);
      return newArray;
    }),
    verticalListSortingStrategy: vi.fn(),
    sortableKeyboardCoordinates: vi.fn(),
  };
});

const mockFile = (name: string, type: string, content: string = '', size?: number): File => {
  const blob = new Blob([content], { type });
  const file = new File([blob], name, { type, lastModified: Date.now() });
  if (size !== undefined) {
    Object.defineProperty(file, 'size', { value: size, writable: false, configurable: true });
  }
  return file;
};


describe('CreateDatasetDialog Management', () => {
  let mockOnClose: Mock<() => void>;
  let mockOnCreate: Mock<CreateDatasetDialogProps['onCreate']>;
  let mockOnUpdate: Mock<CreateDatasetDialogProps['onUpdate']>;
  const OriginalImage = window.Image;

  const initialProps: Omit<CreateDatasetDialogProps, 'onCreate' | 'onUpdate' | 'onClose' | 'dataset'> = {
    isOpen: true,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    dndOnDragEnd = undefined;
    mockOnClose = vi.fn();
    mockOnCreate = vi.fn();
    mockOnUpdate = vi.fn();

    HTMLCanvasElement.prototype.getContext = vi.fn(() => ({
        drawImage: vi.fn(),
        toDataURL: vi.fn(() => 'mock-data-url-canvas'),
    })) as unknown as () => Partial<CanvasRenderingContext2D> | null;

    const MockedFileReader = vi.fn((): FileReader => {
      const self = {
        // Properties
        error: null as DOMException | null,
        readyState: 0 as 0 | 1 | 2, // EMPTY
        result: null as string | ArrayBuffer | null,

        // Event handlers
        onabort: null as (((this: FileReader, ev: ProgressEvent<FileReader>) => void) | null),
        onerror: null as (((this: FileReader, ev: ProgressEvent<FileReader>) => void) | null),
        onload: null as (((this: FileReader, ev: ProgressEvent<FileReader>) => void) | null),
        onloadend: null as (((this: FileReader, ev: ProgressEvent<FileReader>) => void) | null),
        onloadstart: null as (((this: FileReader, ev: ProgressEvent<FileReader>) => void) | null),
        onprogress: null as (((this: FileReader, ev: ProgressEvent<FileReader>) => void) | null),

        // Methods
        abort: vi.fn<() => void>(),
        readAsArrayBuffer: vi.fn<(blob: Blob) => void>(),
        readAsBinaryString: vi.fn<(blob: Blob) => void>(),
        readAsDataURL: vi.fn((_blob: Blob): void => { // Ensure explicit void return for readAsDataURL
          const useFake = vi.isMockFunction(setTimeout) && (setTimeout as unknown as { clock: unknown }).clock;
          const delayFn = useFake ? setTimeout : (fn: () => void) => Promise.resolve().then(fn);

          self.readyState = 1; // LOADING
          delayFn(() => {
            self.result = 'mock-data-url-filereader';
            self.readyState = 2; // DONE
            if (self.onload) {
              self.onload.call(self as FileReader, { target: self } as unknown as ProgressEvent<FileReader>);
            }
          }, 0);
        }),
        readAsText: vi.fn<(blob: Blob, encoding?: string) => void>(),

        // EventTarget methods
        addEventListener: vi.fn(), // Simpler typing
        removeEventListener: vi.fn(), // Simpler typing
        dispatchEvent: vi.fn<(event: Event) => boolean>(),
      };
      return self as unknown as FileReader;
    });
    // Attach static properties to the mock constructor
    Object.defineProperty(MockedFileReader, 'EMPTY', { value: 0, writable: false });
    Object.defineProperty(MockedFileReader, 'LOADING', { value: 1, writable: false });
    Object.defineProperty(MockedFileReader, 'DONE', { value: 2, writable: false });

    vi.spyOn(window, 'FileReader').mockImplementation(MockedFileReader as unknown as typeof FileReader);

    // @ts-expect-error window.Image is not available in a test environment
    window.Image = vi.fn(function() {
      const img = new OriginalImage();
      let _src = '';
      Object.defineProperty(img, 'src', {
        get: () => _src,
        set(value) {
            _src = value;
            img.width = 100;
            img.height = 100;
            const useFake = vi.isMockFunction(setTimeout) && (setTimeout as unknown as { clock: unknown }).clock;
            const delayFn = useFake ? setTimeout : (fn: () => void) => Promise.resolve().then(fn);
            delayFn(() => {
              if (img.onload) { // Null check before calling
                img.onload({} as Event);
              }
            }, 0);
        }
      });
      return img;
    });
  });

  afterEach(() => {
    window.Image = OriginalImage;
    // Ensure fake timers are restored if a test block used them
    // Check if clock exists on setTimeout to determine if it's a Vitest fake timer
    if (vi.isMockFunction(setTimeout) && (setTimeout as unknown as { clock: unknown }).clock) {
        vi.useRealTimers();
    }
  });

  const findFileItemByName = (name: string) => screen.findByText((content, element) => {
    return element?.tagName.toLowerCase() === 'span' && content.startsWith(name);
  }, {}, {timeout: 5000});


  describe('Dataset Creation', () => {
    test('should create a "list" type dataset', async () => {
        render(<CreateDatasetDialog {...initialProps} onClose={mockOnClose} onCreate={mockOnCreate} onUpdate={mockOnUpdate} />);
        await act(async () => {
            fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'My List Dataset' } });
            fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'A desc' } });
            fireEvent.change(screen.getByPlaceholderText('Enter each option on a new line'), { target: { value: 'Option 1\nOption 2' } });
        });
        await act(async () => { fireEvent.click(screen.getByRole('button', { name: 'Create' }))});
        expect(mockOnCreate).toHaveBeenCalledWith({ name: 'My List Dataset', description: 'A desc', type: 'list', data: ['Option 1', 'Option 2'] });
    });

    test('should create a "csv" type dataset with a file', async () => {
        render(<CreateDatasetDialog {...initialProps} onClose={mockOnClose} onCreate={mockOnCreate} onUpdate={mockOnUpdate} />);
        await act(async () => {
            fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'My CSV Dataset' } });
            fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'A CSV desc' } });
            fireEvent.click(screen.getByLabelText('CSV'));
        });
        const csvFile = mockFile('test.csv', 'text/csv', 'h1,h2\nv1,v2', 10); // Provide size
        const fileInput = screen.getByLabelText('CSV Files') as HTMLInputElement;
        await act(async () => {
            fireEvent.change(fileInput, { target: { files: [csvFile] } });
            await new Promise(r => setTimeout(r, 50));
        });
        await findFileItemByName(csvFile.name); // Will look for "test.csv (0.01 KB)"
        await act(async () => { fireEvent.click(screen.getByRole('button', { name: 'Create' }))});
        expect(mockOnCreate).toHaveBeenCalledWith({ name: 'My CSV Dataset', description: 'A CSV desc', type: 'csv', data: [csvFile.name], files: [csvFile] });
    });

    test.skip('should create an "image" type dataset with an image file', async () => { /* Kept skipped */ });

    test('should replace a CSV file if a new file with the same name is uploaded', async () => {
      render(
        <CreateDatasetDialog
          {...initialProps}
          onClose={mockOnClose}
          onCreate={mockOnCreate}
          onUpdate={mockOnUpdate}
        />
      );

      await act(async () => {
        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'CSV Replacement Test' } });
        fireEvent.click(screen.getByLabelText('CSV'));
      });

      const fileInput = screen.getByLabelText('CSV Files') as HTMLInputElement;
      const fileA_v1 = mockFile('fileA.csv', 'text/csv', 'version1', 1024); // 1.00 KB
      const fileA_v2 = mockFile('fileA.csv', 'text/csv', 'version2_new_content', 2048); // 2.00 KB

      await act(async () => {
        fireEvent.change(fileInput, { target: { files: [fileA_v1] } });
        await new Promise(r => setTimeout(r, 50));
      });
      await screen.findByText((content) => content.startsWith('fileA.csv') && content.includes('1.00 KB'));

      await act(async () => {
        fireEvent.change(fileInput, { target: { files: [fileA_v2] } });
        await new Promise(r => setTimeout(r, 50));
      });

      await screen.findByText((content) => content.startsWith('fileA.csv') && content.includes('2.00 KB'));
      expect(screen.queryByText((content) => content.startsWith('fileA.csv') && content.includes('1.00 KB'))).toBeNull();

      await act(async () => { fireEvent.click(screen.getByRole('button', { name: 'Create' }))});
      expect(mockOnCreate).toHaveBeenCalledWith(expect.objectContaining({ data: [fileA_v2.name], files: [fileA_v2] }));
    });
  });

  describe('Dataset Update', () => {
    const existingListDataset: MockDatasetInfo = { id: 'list1', name: 'Existing List', description: 'Old list description', type: 'list', data: ['Old Option 1'], columns: []};
    test('should update a "list" type dataset', async () => {
        render(<CreateDatasetDialog {...initialProps} dataset={existingListDataset} onClose={mockOnClose} onCreate={mockOnCreate} onUpdate={mockOnUpdate} />);
        await act(async () => { fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Updated List Name' } });});
        await act(async () => { fireEvent.click(screen.getByRole('button', { name: 'Update' }))});
        expect(mockOnUpdate).toHaveBeenCalledWith(existingListDataset.id, expect.objectContaining({ name: 'Updated List Name' }));
    });

    const existingCsvDataset: MockDatasetInfo = { id: 'csv1', name: 'Existing CSV', description: 'Old CSV file', type: 'csv', data: ['file1.csv', 'file2.csv', 'file3.csv'], columns: []};
    test('should update a "csv" type dataset by adding a new file', async () => {
        render(<CreateDatasetDialog {...initialProps} dataset={existingCsvDataset} onClose={mockOnClose} onCreate={mockOnCreate} onUpdate={mockOnUpdate} />);
        await findFileItemByName(existingCsvDataset.data[0]);
        const newCsvFile = mockFile('new_upload.csv', 'text/csv');
        const fileInput = screen.getByLabelText('CSV Files') as HTMLInputElement;
        await act(async () => {
            fireEvent.change(fileInput, { target: { files: [newCsvFile] } });
            await new Promise(r => setTimeout(r, 50));
        });
        await findFileItemByName(newCsvFile.name);
        await act(async () => { fireEvent.click(screen.getByRole('button', { name: 'Update' }))});
        expect(mockOnUpdate).toHaveBeenCalledWith(existingCsvDataset.id, expect.objectContaining({ files: [newCsvFile] }));
    });

    test.skip('should update an "image" type dataset by adding a new image', async () => { /* Kept skipped */ });

    describe('DND Reordering', () => {
        const existingCsvDatasetForDND: MockDatasetInfo = { id: 'csvDND', name: 'CSV DND', description: 'CSV DND test', type: 'csv', data: ['fileA.csv', 'fileB.csv', 'fileC.csv'], columns: []};
        test('should reorder files for a "csv" dataset via DND and call onUpdate', async () => {
          // This test does not need fake timers as CSV file item rendering is synchronous after initial load
          render(
            <CreateDatasetDialog
              {...initialProps}
              dataset={existingCsvDatasetForDND as unknown as DatasetInfo}
              onClose={mockOnClose}
              onCreate={mockOnCreate}
              onUpdate={mockOnUpdate}
            />
          );

          await findFileItemByName(existingCsvDatasetForDND.data[0]);
          await findFileItemByName(existingCsvDatasetForDND.data[1]);
          await findFileItemByName(existingCsvDatasetForDND.data[2]);

          // IDs are: dataset.id + '-' + fileName + '-' + index
          const activeItemId = `csvDND-${existingCsvDatasetForDND.data[2]}-2`; // fileC.csv (index 2)
          const overItemId = `csvDND-${existingCsvDatasetForDND.data[0]}-0`;   // fileA.csv (index 0)

          expect(dndOnDragEnd).toBeDefined();
          if (dndOnDragEnd) {
            const dragEndEvent: DragEndEvent = {
              active: { id: activeItemId as UniqueIdentifier } as Active,
              over: { id: overItemId as UniqueIdentifier } as Over,
            } as DragEndEvent;

            await act(async () => { // Wrap state update in act
                dndOnDragEnd!(dragEndEvent);
            });
          }

          // Optional: Verify DOM order change if needed, though payload is primary check
          // const fileItemsAfterDrag = await screen.findAllByRole('button', { name: /Drag to reorder/i });
          // expect(fileItemsAfterDrag[0].closest('[data-sortable-id]')?.getAttribute('data-sortable-id')).toBe(activeItemId);


          await act(async () => {
            fireEvent.click(screen.getByRole('button', { name: 'Update' }));
          });

          expect(mockOnUpdate).toHaveBeenCalledTimes(1);
          expect(mockOnUpdate).toHaveBeenCalledWith(
            existingCsvDatasetForDND.id,
            expect.objectContaining({
              name: existingCsvDatasetForDND.name,
              type: 'csv',
              data: [existingCsvDatasetForDND.data[2], existingCsvDatasetForDND.data[0], existingCsvDatasetForDND.data[1]],
              files: undefined,
            })
          );
          expect(mockOnClose).toHaveBeenCalledTimes(1);
        });
    });
  });
});
