import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import { CreateDatasetDialog } from './dataset';
import { vi } from 'vitest';

// Mock UI components
vi.mock('@/components/ui/dialog', () => ({
  Dialog: (props: { children: React.ReactNode, open: boolean, onOpenChange: (open: boolean) => void }) => {
    // The component's onOpenChange prop is passed to the mock.
    // The mock itself doesn't add an onClick to also trigger it, to simplify.
    return props.open ? <div data-testid="dialog-mock" data-open={props.open}>{props.children}</div> : null;
  },
  DialogContent: ({ children, className }: { children: React.ReactNode, className?:string }) => <div className={className}>{children}</div>,
  DialogDescription: ({ children }: { children: React.ReactNode }) => <p>{children}</p>,
  DialogFooter: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogTitle: ({ children }: { children: React.ReactNode }) => <h5>{children}</h5>,
}));
vi.mock('@/components/ui/button', () => ({
  Button: ({ children, onClick, variant, size, 'aria-label': ariaLabel }: { children: React.ReactNode, onClick: (e?: any) => void, variant?: string, size?: string, 'aria-label'?: string }) => (
    <button onClick={onClick} data-variant={variant} data-size={size} aria-label={ariaLabel}>{children}</button>
  ),
}));
vi.mock('@/components/ui/input', () => ({
  Input: React.forwardRef(({ id, value, onChange, className, type, multiple, accept }: { id: string, value?: string, onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void, className?: string, type?: string, multiple?: boolean, accept?: string }, ref: React.Ref<HTMLInputElement>) => (
    <input ref={ref} id={id} value={value || ''} onChange={onChange} className={className} type={type} multiple={multiple} accept={accept} data-testid={id}/>
  )),
}));
vi.mock('@/components/ui/label', () => ({
  Label: ({ children, htmlFor }: { children: React.ReactNode, htmlFor: string }) => <label htmlFor={htmlFor}>{children}</label>,
}));
vi.mock('@/components/ui/textarea', () => ({
  Textarea: ({ id, value, onChange, placeholder, className }: { id: string, value: string, onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void, placeholder?: string, className?: string }) => (
    <textarea id={id} value={value} onChange={onChange} placeholder={placeholder} className={className} data-testid={id}/>
  ),
}));
vi.mock('@/components/ui/radio-group', () => ({
  RadioGroup: ({ children, value, onValueChange, className }: { children: React.ReactNode, value: string, onValueChange: (value: string) => void, className?: string }) => (
    <div className={className} data-testid="radio-group" data-value={value} onChange={(e: any) => onValueChange(e.target.value)}>
      {children}
    </div>
  ),
  RadioGroupItem: ({ value, id }: { value: string, id: string }) => <input type="radio" name="type" value={value} id={id} data-testid={id} />,
}));
vi.mock('@/components/ui/scroll-area', () => ({
  ScrollArea: ({ children, className }: { children: React.ReactNode, className?: string }) => <div className={className} data-testid="scroll-area">{children}</div>,
}));


describe('CreateDatasetDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnCreate = vi.fn();

  const initialProps = {
    isOpen: true,
    onClose: mockOnClose,
    onCreate: mockOnCreate,
  };

  beforeEach(() => {
    vi.clearAllMocks(); // Use vi.clearAllMocks() for Vitest
  });

  it('renders common fields and default type (list) correctly when open', () => {
    render(<CreateDatasetDialog {...initialProps} />);
    expect(screen.getByText('Create New Dataset')).toBeInTheDocument();
    expect(screen.getByLabelText('Name*')).toBeInTheDocument();
    expect(screen.getByLabelText('Description')).toBeInTheDocument();
    expect(screen.getByLabelText('Type*')).toBeInTheDocument();
    // Check that radio group defaults to list by checking the group's data-value attribute
    expect(screen.getByTestId('radio-group')).toHaveAttribute('data-value', 'list');
    // Check that list-specific fields are visible
    expect(screen.getByLabelText('Options*')).toBeInTheDocument();
    expect(screen.getByTestId('list-options')).toBeInTheDocument();
    // Check that csv-specific fields are not visible
    expect(screen.queryByLabelText('CSV Files*')).not.toBeInTheDocument();
    expect(screen.getByText('Cancel')).toBeInTheDocument();
    expect(screen.getByText('Create')).toBeInTheDocument();
  });

  it('does not render when not open', () => {
    render(<CreateDatasetDialog {...initialProps} isOpen={false} />);
    expect(screen.queryByText('Create New Dataset')).not.toBeInTheDocument();
  });

  it('calls onClose and resets form when cancel button is clicked', () => {
    const { rerender } = render(<CreateDatasetDialog {...initialProps} />);
    fireEvent.change(screen.getByLabelText('Name*'), { target: { value: 'Test Data' } });
    fireEvent.change(screen.getByTestId('list-options'), { target: { value: 'Test Options' } });

    fireEvent.click(screen.getByText('Cancel'));
    expect(mockOnClose).toHaveBeenCalledTimes(1);

    // Simulate reopening the dialog with fresh state (as parent would control isOpen)
    // And initialProps already has mockOnClose and mockOnCreate which are cleared in beforeEach
    rerender(<CreateDatasetDialog {...initialProps} isOpen={true} />);

    expect(screen.getByLabelText('Name*')).toHaveValue('');
    expect(screen.getByTestId('list-options')).toHaveValue('');
    // Check type is reset to 'list'
    expect(screen.getByTestId('radio-group')).toHaveAttribute('data-value', 'list');
  });

  it('switches to CSV type and shows relevant fields', () => {
    render(<CreateDatasetDialog {...initialProps} />);
    const csvRadio = screen.getByTestId('type-csv');
    fireEvent.click(csvRadio); // Click the actual radio input
    expect(csvRadio).toBeChecked();
    expect(screen.getByLabelText('CSV Files*')).toBeInTheDocument();
    expect(screen.getByTestId('csv-files')).toBeInTheDocument();
    expect(screen.queryByLabelText('Options*')).not.toBeInTheDocument();
  });

  describe('Form Submission and Validation', () => {
    it('submits with "list" type and calls onCreate with correct data', () => {
      render(<CreateDatasetDialog {...initialProps} />);
      fireEvent.change(screen.getByLabelText('Name*'), { target: { value: 'My List Dataset' } });
      fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'List Description' } });
      fireEvent.change(screen.getByTestId('list-options'), { target: { value: 'Option1\nOption2\n Option3 ' } });
      fireEvent.click(screen.getByText('Create'));

      expect(mockOnCreate).toHaveBeenCalledWith({
        name: 'My List Dataset',
        description: 'List Description',
        type: 'list',
        options: ['Option1', 'Option2', 'Option3'],
      });
      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('submits with "csv" type and calls onCreate with files', async () => {
      render(<CreateDatasetDialog {...initialProps} />);
      fireEvent.click(screen.getByTestId('type-csv')); // Switch to CSV

      fireEvent.change(screen.getByLabelText('Name*'), { target: { value: 'My CSV Dataset' } });
      fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'CSV Description' } });

      const file1 = new File(['col1,col2\nval1,val2'], 'data1.csv', { type: 'text/csv' });
      const file2 = new File(['colA,colB\ncolA,colB'], 'data2.csv', { type: 'text/csv' });
      const fileInput = screen.getByTestId('csv-files');

      await act(async () => {
        fireEvent.change(fileInput, { target: { files: [file1, file2] } });
      });

      expect(screen.getByText(/data1.csv/)).toBeInTheDocument();
      expect(screen.getByText(/data2.csv/)).toBeInTheDocument();

      fireEvent.click(screen.getByText('Create'));

      expect(mockOnCreate).toHaveBeenCalledWith({
        name: 'My CSV Dataset',
        description: 'CSV Description',
        type: 'csv',
        files: [file1, file2],
      });
      expect(mockOnClose).toHaveBeenCalledTimes(1); // This is the assertion that fails
    });

    it('shows validation error if name is empty', async () => {
      render(<CreateDatasetDialog {...initialProps} />);
      await act(async () => {
        fireEvent.click(screen.getByText('Create'));
      });
      // After act, DOM should be updated
      expect(screen.getByLabelText('Name*')).toHaveClass('border-red-500');
      expect(screen.getByText('Name cannot be empty')).toBeInTheDocument();
      expect(mockOnCreate).not.toHaveBeenCalled();
    });

    it('shows validation error if list options are empty for "list" type', async () => {
      render(<CreateDatasetDialog {...initialProps} />); // Defaults to list type
      fireEvent.change(screen.getByLabelText('Name*'), { target: { value: 'Test List' } });
      await act(async () => {
        fireEvent.click(screen.getByText('Create'));
      });
      expect(screen.getByTestId('list-options')).toHaveClass('border-red-500');
      expect(screen.getByText('List options cannot be empty')).toBeInTheDocument();
      expect(mockOnCreate).not.toHaveBeenCalled();
    });

    it('shows validation error if no files are selected for "csv" type', async () => {
      render(<CreateDatasetDialog {...initialProps} />);
      fireEvent.click(screen.getByTestId('type-csv')); // Switch to CSV
      fireEvent.change(screen.getByLabelText('Name*'), { target: { value: 'Test CSV' } });
      await act(async () => {
        fireEvent.click(screen.getByText('Create'));
      });
      expect(screen.getByText('Please select at least one CSV file')).toBeInTheDocument();
      expect(mockOnCreate).not.toHaveBeenCalled();
    });
  });

  describe('CSV File Handling', () => {
    it('allows selecting multiple CSV files and displays them', async () => {
      render(<CreateDatasetDialog {...initialProps} />);
      fireEvent.click(screen.getByTestId('type-csv'));
      const fileInput = screen.getByTestId('csv-files');
      const file1 = new File(['content1'], 'test1.csv', {type: 'text/csv'});
      const file2 = new File(['content2'], 'test2.csv', {type: 'text/csv'});

      await act(async () => {
        fireEvent.change(fileInput, { target: { files: [file1, file2] } });
      });

      expect(screen.getByText(/test1.csv/)).toBeInTheDocument();
      expect(screen.getByText(/test2.csv/)).toBeInTheDocument();
      expect(screen.queryByText('Only CSV files are allowed.')).not.toBeInTheDocument();
    });

    it('shows error and does not add non-CSV files', async () => {
        render(<CreateDatasetDialog {...initialProps} />);
        fireEvent.click(screen.getByTestId('type-csv'));
        const fileInput = screen.getByTestId('csv-files');
        const txtFile = new File(['content'], 'test.txt', {type: 'text/plain'});

        await act(async () => {
            fireEvent.change(fileInput, { target: { files: [txtFile] } });
        });

        expect(screen.getByText('Only CSV files are allowed.')).toBeInTheDocument();
        expect(screen.queryByText(/test.txt/)).not.toBeInTheDocument(); // Assuming we don't add invalid files
    });

    it('allows removing a selected CSV file', async () => {
      render(<CreateDatasetDialog {...initialProps} />);
      fireEvent.click(screen.getByTestId('type-csv'));
      const fileInput = screen.getByTestId('csv-files');
      const file1 = new File(['content1'], 'test1.csv', {type: 'text/csv'});

      await act(async () => {
        fireEvent.change(fileInput, { target: { files: [file1] } });
      });

      expect(screen.getByText(/test1.csv/)).toBeInTheDocument();
      fireEvent.click(screen.getByRole('button', { name: /Remove test1.csv/i }));
      expect(screen.queryByText(/test1.csv/)).not.toBeInTheDocument();
    });

    it('prevents duplicate files from being added', async () => {
        render(<CreateDatasetDialog {...initialProps} />);
        fireEvent.click(screen.getByTestId('type-csv'));
        const fileInput = screen.getByTestId('csv-files');
        const file1 = new File(['content1'], 'test1.csv', {type: 'text/csv'});

        await act(async () => {
            fireEvent.change(fileInput, { target: { files: [file1] } });
        });
        await act(async () => { // Try adding the same file again
            fireEvent.change(fileInput, { target: { files: [file1] } });
        });

        expect(screen.getAllByText(/test1.csv/).length).toBe(1); // Should only appear once
    });
  });

  it('clears form fields and errors when closed and reopened', async () => {
    const { rerender } = render(<CreateDatasetDialog {...initialProps} />);
    // Set some values and trigger errors
    fireEvent.change(screen.getByLabelText('Name*'), { target: { value: '' } });
    await act(async () => {
      fireEvent.click(screen.getByText('Create')); // Trigger name error
    });
    expect(screen.getByText('Name cannot be empty')).toBeInTheDocument(); // Check error is present
    fireEvent.change(screen.getByTestId('list-options'), { target: { value: 'some option' } });


    // Simulate closing the dialog by clicking cancel
    fireEvent.click(screen.getByText('Cancel'));
    expect(mockOnClose).toHaveBeenCalledTimes(1);

    // Reopen the dialog (props would change in a real app, here we re-render)
    // Ensure mocks are cleared for the new render's interactions
    mockOnClose.mockClear();
    mockOnCreate.mockClear();

    rerender(<CreateDatasetDialog {...initialProps} isOpen={true} />); // Use rerender

    // Check if fields are cleared
    expect(screen.getByLabelText('Name*')).toHaveValue('');
    expect(screen.getByLabelText('Description')).toHaveValue('');
    expect(screen.getByTestId('list-options')).toHaveValue('');
    // Check if errors are cleared
    expect(screen.queryByText('Name cannot be empty')).not.toBeInTheDocument();
    expect(screen.queryByText('List options cannot be empty')).not.toBeInTheDocument();
    // Check type is reset to 'list' by checking the group's data-value attribute
    expect(screen.getByTestId('radio-group')).toHaveAttribute('data-value', 'list');
  });

});
