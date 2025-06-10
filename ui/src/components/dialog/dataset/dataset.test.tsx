import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { CreateDatasetDialog, CreateDatasetDialogProps } from './dataset'; // Adjust path as needed
import { DatasetInfo } from '@/actions'; // Adjust path as needed

// Mock the GenerateOptionsDialog as it's not the focus of these tests
jest.mock('../generate-options-dialog', () => ({
  GenerateOptionsDialog: jest.fn(() => null),
}));

// Mock DnD hooks and context providers as they are complex and not the focus
jest.mock('@dnd-kit/core', () => ({
  ...jest.requireActual('@dnd-kit/core'), // Import and retain default exports
  DndContext: jest.fn(({ children }) => <div>{children}</div>), // Simple pass-through
  useSensor: jest.fn(),
  useSensors: jest.fn(),
  PointerSensor: jest.fn(),
  KeyboardSensor: jest.fn(),
  closestCenter: jest.fn(),
}));

jest.mock('@dnd-kit/sortable', () => ({
  ...jest.requireActual('@dnd-kit/sortable'), // Import and retain default exports
  SortableContext: jest.fn(({ children }) => <div>{children}</div>), // Simple pass-through
  useSortable: jest.fn(() => ({
    attributes: {},
    listeners: {},
    setNodeRef: jest.fn(),
    transform: null,
    transition: null,
    isDragging: false,
  })),
  verticalListSortingStrategy: jest.fn(),
  sortableKeyboardCoordinates: jest.fn(),
  arrayMove: jest.fn((arr, from, to) => {
    const newArray = [...arr];
    const [element] = newArray.splice(from, 1);
    newArray.splice(to, 0, element);
    return newArray;
  }),
}));


const mockOnClose = jest.fn();
const mockOnCreate = jest.fn();
const mockOnUpdate = jest.fn();

const defaultProps: CreateDatasetDialogProps = {
  isOpen: true,
  onClose: mockOnClose,
  onCreate: mockOnCreate,
  onUpdate: mockOnUpdate,
};

const renderDialog = (props: Partial<CreateDatasetDialogProps> = {}) => {
  return render(<CreateDatasetDialog {...defaultProps} {...props} />);
};

describe('CreateDatasetDialog - Image Type', () => {
  beforeEach(() => {
    // Reset mocks before each test
    mockOnClose.mockClear();
    mockOnCreate.mockClear();
    mockOnUpdate.mockClear();
    // Mock FileReader
    global.FileReader = jest.fn(() => ({
      readAsDataURL: jest.fn(),
      onload: jest.fn(),
      onerror: jest.fn(),
      result: 'mock-data-url', // Default mock result
    })) as any;
    // Mock Image
    global.Image = jest.fn(() => ({
        onload: jest.fn(),
        onerror: jest.fn(),
        src: '',
        width: 100, // mock dimensions
        height: 100, // mock dimensions
    })) as any;
    // Mock canvas
    global.HTMLCanvasElement.prototype.getContext = jest.fn(() => ({
        drawImage: jest.fn(),
    })) as any;
    global.HTMLCanvasElement.prototype.toDataURL = jest.fn(() => 'mock-thumbnail-data-url') as any;

  });

  test('renders the "Image" radio button and allows selection', () => {
    renderDialog();
    const imageRadio = screen.getByLabelText('Image') as HTMLInputElement;
    expect(imageRadio).toBeInTheDocument();
    fireEvent.click(imageRadio);
    expect(imageRadio.checked).toBe(true);
  });

  test('displays correct file input when "Image" type is selected', () => {
    renderDialog();
    fireEvent.click(screen.getByLabelText('Image'));

    const imageInput = screen.getByLabelText('Image Files');
    expect(imageInput).toBeInTheDocument();
    expect(imageInput).toHaveAttribute('accept', 'image/png, image/jpeg, image/gif, .png, .jpg, .jpeg, .gif');

    expect(screen.queryByLabelText('CSV Files')).not.toBeInTheDocument();
  });

  test('shows error message for non-image file selection', async () => {
    renderDialog();
    fireEvent.click(screen.getByLabelText('Image'));
    const imageInput = screen.getByLabelText('Image Files');

    const nonImageFile = new File(['hello'], 'hello.txt', { type: 'text/plain' });
    fireEvent.change(imageInput, { target: { files: [nonImageFile] } });

    expect(await screen.findByText('Only image files (PNG, JPG, GIF) are allowed.')).toBeVisible();
  });

  test('simulates image file selection and expects FileItems to be processed (simplified)', async () => {
    renderDialog();
    fireEvent.click(screen.getByLabelText('Image'));
    const imageInput = screen.getByLabelText('Image Files');

    const imageFile = new File(['(⌐□_□)'], 'chucknorris.png', { type: 'image/png' });

    // For this test, we need our mocks to actually trigger onload
    (global.FileReader as jest.Mock).mockImplementationOnce(function (this: any) {
      this.readAsDataURL = jest.fn();
      this.result = 'mock-data-url-chuck';
      // Simulate async behavior: call onload when readAsDataURL is called
      jest.spyOn(this, 'readAsDataURL').mockImplementationOnce(() => {
        if (this.onload) {
          this.onload({ target: { result: this.result } });
        }
      });
      return this;
    } as any);

    (global.Image as jest.Mock).mockImplementationOnce(function (this: any) {
        this.width = 50;
        this.height = 50;
        // Simulate async behavior: call onload when src is set
        jest.spyOn(this, 'src', 'set').mockImplementationOnce(function(this: any, value) {
            this._src = value;
            if (this.onload) {
                this.onload();
            }
        });
        return this;
    } as any);

    fireEvent.change(imageInput, { target: { files: [imageFile] } });

    await waitFor(() => {
      expect(screen.getByText(imageFile.name)).toBeInTheDocument();
    });

     await waitFor(() => {
        const thumbnailImg = screen.getByAltText(`Image thumbnail for ${imageFile.name}`) as HTMLImageElement;
        expect(thumbnailImg).toBeInTheDocument();
        expect(thumbnailImg.src).toContain('mock-thumbnail-data-url');
     });
  });


  test('displays thumbnails for existing image dataset (simulated by adding files)', async () => {
    // This test simulates adding files that then get thumbnails
    renderDialog();
    fireEvent.click(screen.getByLabelText('Image'));
    const imageInput = screen.getByLabelText('Image Files');
    const imageFile1 = new File(['img1'], 'photo1.jpg', { type: 'image/jpeg' });
    const imageFile2 = new File(['img2'], 'photo2.png', { type: 'image/png' });

    // Setup mocks to call onload
    (global.FileReader as jest.Mock).mockImplementation(function (this: any) {
      this.readAsDataURL = jest.fn();
      // Simulate async behavior: call onload when readAsDataURL is called
      const originalReadAsDataURL = this.readAsDataURL;
      this.readAsDataURL = (file: File) => {
        this.result = `data-url-for-${file.name}`; // Set result before calling onload
        if (this.onload) {
            this.onload({ target: { result: this.result } });
        }
        originalReadAsDataURL.call(this, file);
      };
      return this;
    } as any);

    (global.Image as jest.Mock).mockImplementation(function (this: any) {
        this.width = 50;
        this.height = 50;
        const originalSrcSetter = Object.getOwnPropertyDescriptor(window.Image.prototype, 'src')?.set;
        Object.defineProperty(this, 'src', {
            set: function(value) {
                if (originalSrcSetter) originalSrcSetter.call(this, value);
                if (this.onload) {
                    this.onload();
                }
            }
        });
        return this;
    } as any);

    fireEvent.change(imageInput, { target: { files: [imageFile1, imageFile2] } });

    await waitFor(() => {
      expect(screen.getByAltText(`Image thumbnail for ${imageFile1.name}`)).toHaveAttribute('src', 'mock-thumbnail-data-url');
      expect(screen.getByAltText(`Image thumbnail for ${imageFile2.name}`)).toHaveAttribute('src', 'mock-thumbnail-data-url');
    });
  });

  test('submits correct data for a new image dataset', async () => {
    renderDialog();
    fireEvent.click(screen.getByLabelText('Image'));
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Image Dataset' } });
    fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'A test set of images' } });

    const imageFile1 = new File(['img1_content'], 'image1.png', { type: 'image/png' });
    const imageFile2 = new File(['img2_content'], 'image2.jpg', { type: 'image/jpeg' });

    // Setup mocks to call onload for file processing
     (global.FileReader as jest.Mock).mockImplementation(function (this: any) {
      this.readAsDataURL = jest.fn();
      const originalReadAsDataURL = this.readAsDataURL;
      this.readAsDataURL = (file: File) => {
        this.result = `data-url-for-${file.name}`;
        if (this.onload) { this.onload({ target: { result: this.result } });}
        originalReadAsDataURL.call(this, file);
      };
      return this;
    } as any);
    (global.Image as jest.Mock).mockImplementation(function (this: any) {
        this.width = 50; this.height = 50;
        const originalSrcSetter = Object.getOwnPropertyDescriptor(window.Image.prototype, 'src')?.set;
        Object.defineProperty(this, 'src', {
            set: function(value) {
                if (originalSrcSetter) originalSrcSetter.call(this, value);
                if (this.onload) { this.onload(); }
            }
        });
        return this;
    } as any);

    const imageInput = screen.getByLabelText('Image Files');
    fireEvent.change(imageInput, { target: { files: [imageFile1, imageFile2] } });

    await waitFor(() => {
        expect(screen.getByText(imageFile1.name)).toBeInTheDocument();
        expect(screen.getByText(imageFile2.name)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: 'Create' }));

    await waitFor(() => {
      expect(mockOnCreate).toHaveBeenCalledTimes(1);
      expect(mockOnCreate).toHaveBeenCalledWith({
        name: 'Test Image Dataset',
        description: 'A test set of images',
        type: 'image',
        data: [imageFile1.name, imageFile2.name],
        files: [imageFile1, imageFile2],
      });
    });
  });

  test('shows validation error if no image files are selected for a new dataset', async () => {
    renderDialog();
    fireEvent.click(screen.getByLabelText('Image'));
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Empty Image Dataset' } });

    fireEvent.click(screen.getByRole('button', { name: 'Create' }));

    expect(await screen.findByText('Please select at least one image file')).toBeVisible();
    expect(mockOnCreate).not.toHaveBeenCalled();
  });

  test('loads existing image dataset and populates file items (no files, just names)', () => {
    const existingDataset: DatasetInfo = {
        id: 'imgd1',
        name: 'My Old Images',
        description: 'Old collection',
        type: 'image',
        data: ['old_pic1.jpg', 'old_pic2.png'],
    };
    renderDialog({ dataset: existingDataset });

    expect(screen.getByDisplayValue(existingDataset.name)).toBeInTheDocument();
    expect(screen.getByDisplayValue(existingDataset.description)).toBeInTheDocument();
    expect(screen.getByLabelText('Image')).toBeChecked();
    expect(screen.getByLabelText('Image')).toBeDisabled();

    expect(screen.getByText('old_pic1.jpg')).toBeInTheDocument();
    expect(screen.getByText('old_pic2.png')).toBeInTheDocument();

    expect(screen.queryByAltText(`Image thumbnail for old_pic1.jpg`)).not.toBeInTheDocument();
    expect(screen.queryByAltText(`Image thumbnail for old_pic2.png`)).not.toBeInTheDocument();
  });

});
