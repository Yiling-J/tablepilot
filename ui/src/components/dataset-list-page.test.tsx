import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach, Mock } from 'vitest';
import { DatasetListPage } from './dataset-list-page';
import * as actions from '@/actions';
import { MemoryRouter, useLocation, useNavigate } from 'react-router-dom';
import { TestProvider } from '@/test/helpers/test-provider';

vi.mock('@/actions');
vi.mock('react-router-dom', async () => {
  const originalModule = await vi.importActual('react-router-dom');
  return {
    ...originalModule,
    useNavigate: vi.fn(),
    useLocation: vi.fn(),
  };
});

import * as ActualToastHook from '@/hooks/use-toast';
vi.mock('@/hooks/use-toast', async (importOriginal) => {
  const actual = await importOriginal<typeof ActualToastHook>();
  return {
    ...actual,
    useToast: vi.fn(() => ({
      toast: vi.fn(),
    })),
  };
});


describe('DatasetListPage Search Functionality', () => {
  const sampleDatasets = [
    { id: 'ds1', name: 'Dataset Alpha', description: 'First one', type: 'csv' },
    { id: 'ds2', name: 'Dataset Beta', description: 'Second one', type: 'list' },
    { id: 'ds3', name: 'Gamma Dataset', description: 'Third one', type: 'csv' },
  ];

  beforeEach(async () => {
    vi.resetAllMocks();

    (actions.getDatasets as Mock).mockResolvedValue({
      datasets: sampleDatasets,
    });
    (actions.deleteDataset as Mock).mockResolvedValue({});
    (actions.createDataset as Mock).mockResolvedValue({});
    (actions.updateDataset as Mock).mockResolvedValue({});


    (useNavigate as Mock).mockReturnValue(vi.fn());
    (useLocation as Mock).mockReturnValue({
        key: 'testKey',
        pathname: '/datasets',
        search: '',
        hash: '',
        state: null,
    });

    render(
      <MemoryRouter>
        <TestProvider>
          <DatasetListPage />
        </TestProvider>
      </MemoryRouter>
    );

    await waitFor(() => expect(screen.getByText('Dataset Alpha')).toBeInTheDocument());
    await waitFor(() => expect(screen.getByText('Dataset Beta')).toBeInTheDocument());
    await waitFor(() => expect(screen.getByText('Gamma Dataset')).toBeInTheDocument());
  });

  it('should render all datasets initially', () => {
    expect(screen.getByText('Dataset Alpha')).toBeInTheDocument();
    expect(screen.getByText('Dataset Beta')).toBeInTheDocument();
    expect(screen.getByText('Gamma Dataset')).toBeInTheDocument();
  });

  it('should filter datasets based on search query', async () => {
    const searchInput = screen.getByPlaceholderText('Search datasets...');
    await userEvent.type(searchInput, 'Alpha');

    await waitFor(() => {
        expect(screen.getByText('Dataset Alpha')).toBeInTheDocument();
        expect(screen.queryByText('Dataset Beta')).not.toBeInTheDocument();
        expect(screen.queryByText('Gamma Dataset')).not.toBeInTheDocument();
    });
  });

  it('should be case-insensitive', async () => {
    const searchInput = screen.getByPlaceholderText('Search datasets...');
    await userEvent.type(searchInput, 'gamma');

    await waitFor(() => {
        expect(screen.queryByText('Dataset Alpha')).not.toBeInTheDocument();
        expect(screen.queryByText('Dataset Beta')).not.toBeInTheDocument();
        expect(screen.getByText('Gamma Dataset')).toBeInTheDocument();
    });
  });

  it('should show no results message if search matches nothing', async () => {
    const searchInput = screen.getByPlaceholderText('Search datasets...');
    await userEvent.type(searchInput, 'NonExistentDataset');

    // The component currently does not implement a specific "no results" message.
    // It would render an empty list.
    await waitFor(() => {
        expect(screen.queryByText('Dataset Alpha')).not.toBeInTheDocument();
        expect(screen.queryByText('Dataset Beta')).not.toBeInTheDocument();
        expect(screen.queryByText('Gamma Dataset')).not.toBeInTheDocument();
    });
    // If a "No datasets found..." message were implemented, we'd assert its presence here.
  });

  it('should show all datasets when search query is cleared', async () => {
    const searchInput = screen.getByPlaceholderText('Search datasets...');
    await userEvent.type(searchInput, 'Alpha');

    await waitFor(() => expect(screen.getByText('Dataset Alpha')).toBeInTheDocument());
    await waitFor(() => expect(screen.queryByText('Dataset Beta')).not.toBeInTheDocument());

    await userEvent.clear(searchInput);

    await waitFor(() => {
        expect(screen.getByText('Dataset Alpha')).toBeInTheDocument();
        expect(screen.getByText('Dataset Beta')).toBeInTheDocument();
        expect(screen.getByText('Gamma Dataset')).toBeInTheDocument();
    });
  });
});

vi.mock('@/components/dialog/dataset/dataset', () => ({
  CreateDatasetDialog: vi.fn(({ isOpen, onClose, onCreate, dataset }: any) => { // eslint-disable-line @typescript-eslint/no-explicit-any
    if (!isOpen) return null;
    return (
      <div data-testid="mock-create-dataset-dialog">
        <button
          data-testid="mock-create-dataset-submit"
          onClick={() => {
            const type = dataset?.type || 'csv';
            onCreate({
              name: dataset?.name || 'New Mocked Dataset',
              description: dataset?.description || 'Mocked description',
              type: type,
              files: type === 'csv' ? [new File([''], 'mock.csv', { type: 'text/csv' })] : undefined,
              options: type === 'list' ? ['opt1'] : undefined,
            });
          }}
        >
          Create/Update
        </button>
        <button data-testid="mock-create-dataset-close" onClick={onClose}>Close</button>
      </div>
    );
  }),
}));
vi.mock('@/components/dialog/dataset/info', () => ({
  DatasetInfoDialog: vi.fn(() => <div data-testid="mock-dataset-info-dialog" />),
}));
vi.mock('@/components/dialog/dataset/preview', () => ({
  DatasetPreviewDialog: vi.fn(() => <div data-testid="mock-dataset-preview-dialog" />),
}));

vi.mock('@/components/ui/common-card', () => ({
  CommonCard: vi.fn(({ name, children, onDelete, onEdit, onClick, badgeText }: {
    name: string;
    children: React.ReactNode;
    onDelete?: () => void;
    onEdit?: () => void;
    onClick: () => void;
    badgeText?: string;
  }) => (
    <div data-testid={`common-card-${name.replace(/\s+/g, "-")}`}>
      <button onClick={onClick} data-testid={`view-${name.replace(/\s+/g, "-")}`}>{name}</button>
      <div>{children}</div>
      {badgeText && <span>{badgeText}</span>}
      {onDelete && <button onClick={onDelete} data-testid={`delete-${name.replace(/\s+/g, "-")}`}>Delete</button>}
      {onEdit && <button onClick={onEdit} data-testid={`edit-${name.replace(/\s+/g, "-")}`}>Edit</button>}
    </div>
  )),
}));


describe('DatasetListPage Create, Delete and Refresh', () => {
  const initialDatasets = [
    { id: '1', name: 'Dataset Alpha', description: 'First one', type: 'csv' as const },
    { id: '2', name: 'Dataset Beta', description: 'Second one', type: 'list' as const },
  ];

  const newDataset = { id: '3', name: 'New Mocked Dataset', description: 'Mocked description', type: 'csv' as const };

  const mockGetDatasets = actions.getDatasets as Mock;
  const mockCreateDataset = actions.createDataset as Mock;
  const mockDeleteDataset = actions.deleteDataset as Mock;

  beforeEach(() => {
    vi.resetAllMocks();
    (useNavigate as Mock).mockReturnValue(vi.fn());
    (useLocation as Mock).mockReturnValue({
        key: 'testKeyCUD',
        pathname: '/datasets',
        search: '',
        hash: '',
        state: null,
    });

    mockCreateDataset.mockResolvedValue({ ...newDataset });
    mockDeleteDataset.mockResolvedValue(undefined);
  });

  it('should refresh the list after creating a new dataset', async () => {
    mockGetDatasets
      .mockResolvedValueOnce({ datasets: initialDatasets })
      .mockResolvedValueOnce({ datasets: [...initialDatasets, newDataset] });

    render(
      <MemoryRouter>
        <TestProvider>
          <DatasetListPage />
        </TestProvider>
      </MemoryRouter>
    );

    await waitFor(() => expect(screen.getByText('Dataset Alpha')).toBeInTheDocument());
    expect(screen.getByText('Dataset Beta')).toBeInTheDocument();

    const addNewButton = screen.getByRole('button', { name: /Add New Dataset/i });
    await userEvent.click(addNewButton);

    await waitFor(() => expect(screen.getByTestId('mock-create-dataset-dialog')).toBeInTheDocument());

    const submitButton = screen.getByTestId('mock-create-dataset-submit');
    await userEvent.click(submitButton);

    await waitFor(() => expect(screen.getByText('New Mocked Dataset')).toBeInTheDocument(), { timeout: 2000 });

    expect(mockCreateDataset).toHaveBeenCalledTimes(1);
    await waitFor(() => expect(mockGetDatasets).toHaveBeenCalledTimes(2)); // Initial load + load after create
  });

  it('should refresh the list after deleting a dataset', async () => {
    mockGetDatasets
      .mockResolvedValueOnce({ datasets: initialDatasets })
      .mockResolvedValueOnce({ datasets: [initialDatasets[1]] }); // Dataset Alpha (id: '1') is removed

    render(
      <MemoryRouter>
        <TestProvider>
          <DatasetListPage />
        </TestProvider>
      </MemoryRouter>
    );

    await waitFor(() => expect(screen.getByText('Dataset Alpha')).toBeInTheDocument());
    expect(screen.getByText('Dataset Beta')).toBeInTheDocument();

    const deleteButtonAlpha = screen.getByTestId('delete-Dataset-Alpha');
    await userEvent.click(deleteButtonAlpha);
    // Note: Mocked CommonCard's onDelete is called directly, no confirmation dialog step here.

    await waitFor(() => expect(screen.queryByText('Dataset Alpha')).not.toBeInTheDocument(), { timeout: 2000 });
    await waitFor(() => expect(screen.getByText('Dataset Beta')).toBeInTheDocument());

    expect(mockDeleteDataset).toHaveBeenCalledWith('1');
    await waitFor(() => expect(mockGetDatasets).toHaveBeenCalledTimes(2)); // Initial load + load after delete
  });
});
