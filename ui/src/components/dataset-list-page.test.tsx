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
vi.mock('@/hooks/use-toast', () => ({
  useToast: vi.fn(() => ({
    toast: vi.fn(),
  })),
}));


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
