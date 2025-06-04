import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach, MockedFunction } from 'vitest'; // Added MockedFunction, removed duplicate vi
import { GenerateOptionsDialog } from './generate-options-dialog';
import type { GenerateOptionsDialogProps } from './generate-options-dialog';
// Removed duplicate: import { vi } from 'vitest';


// Mock actions
vi.mock('../../actions', async (importOriginal) => { // Corrected path
  const actual = await importOriginal<typeof import('../../actions')>();
  return {
    ...actual,
    generateOptions: vi.fn(),
  };
});

// Mock react-hot-toast
vi.mock('react-hot-toast', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock ModelSelector
vi.mock('../model-selector', () => ({
  ModelSelector: vi.fn(({ selectModel, value, id, className }) => ( // Expect selectModel, optionally a value if needed by mock's logic
    <input
      data-testid="model-selector"
      id={id}
      className={className}
      value={value || ''} // Use 'value' if the mock needs to display a passed-in value
      onChange={(e) => { if (selectModel) selectModel(e.target.value); }} // Call selectModel
      placeholder="Mock Model Selector"
    />
  )),
}));

// Import after mocks
import * as actions from '../../actions'; // Corrected path
import { toast } from 'react-hot-toast';

const mockGenerateOptions = actions.generateOptions as MockedFunction<typeof actions.generateOptions>; // Changed to MockedFunction

const defaultProps: GenerateOptionsDialogProps = {
  isOpen: true,
  onClose: vi.fn(),
  onGenerationComplete: vi.fn(),
  datasetName: 'MyDataset',
  datasetDescription: 'A test dataset.',
};

const renderDialog = (props?: Partial<GenerateOptionsDialogProps>) => {
  return render(<GenerateOptionsDialog {...defaultProps} {...props} />);
};

describe('GenerateOptionsDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGenerateOptions.mockResolvedValue(['Option 1 from AI', 'Option 2 from AI']);
  });

  describe('Rendering and Initial State', () => {
    it('renders with the correct title', () => {
      renderDialog();
      expect(screen.getByText('Generate AI Options')).toBeInTheDocument();
    });

    it('renders model selector, prompt textarea, and buttons', () => {
      renderDialog();
      expect(screen.getByTestId('model-selector')).toBeInTheDocument();
      expect(screen.getByLabelText('Prompt')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Generate' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
    });

    it('pre-fills prompt with dataset name and description', () => {
      renderDialog();
      const promptTextarea = screen.getByLabelText('Prompt') as HTMLTextAreaElement;
      expect(promptTextarea.value).toContain('MyDataset');
      expect(promptTextarea.value).toContain('A test dataset.');
    });

    it('pre-fills prompt with a default message if name and description are absent', () => {
      renderDialog({ datasetName: undefined, datasetDescription: undefined });
      const promptTextarea = screen.getByLabelText('Prompt') as HTMLTextAreaElement;
      expect(promptTextarea.value).toContain('Based on a dataset, generate a list of relevant options.');
    });
     it('pre-fills prompt with dataset name if description is absent', () => {
      renderDialog({ datasetDescription: undefined });
      const promptTextarea = screen.getByLabelText('Prompt') as HTMLTextAreaElement;
      expect(promptTextarea.value).toContain('MyDataset');
      expect(promptTextarea.value).not.toContain('Description:');
    });
  });

  describe('User Interactions and Action Calls', () => {
    it('updates selectedModel when model selector changes', async () => {
      renderDialog();
      const modelSelectorInput = screen.getByTestId('model-selector');
      await userEvent.type(modelSelectorInput, 'test-model');
      // This test relies on the mock implementation of ModelSelector correctly calling onModelChange
      // The state itself is internal, so we check its effect on the "Generate" button enablement
      expect(screen.getByRole('button', { name: 'Generate' })).not.toBeDisabled();
    });

    it('updates prompt when textarea value changes', async () => {
      renderDialog();
      const promptTextarea = screen.getByLabelText('Prompt');
      await userEvent.clear(promptTextarea);
      await userEvent.type(promptTextarea, 'New custom prompt.');
      expect(promptTextarea).toHaveValue('New custom prompt.');
    });

    it('shows error toast if generate is clicked with no model selected', async () => {
      renderDialog(); // selectedModel is initially ""
      const generateButton = screen.getByRole('button', { name: 'Generate' });
      // Prompt is pre-filled, so only model selection is the issue initially
      // Manually clear prompt to ensure button is enabled if model was selected
      const promptTextarea = screen.getByLabelText('Prompt');
      await userEvent.clear(promptTextarea);
      await userEvent.type(promptTextarea, "test prompt");


      await userEvent.click(generateButton);
      expect(toast.error).toHaveBeenCalledWith('Please select a model.');
      expect(mockGenerateOptions).not.toHaveBeenCalled();
    });

    it('shows error toast if generate is clicked with empty prompt', async () => {
      renderDialog();
      const modelSelectorInput = screen.getByTestId('model-selector');
      await userEvent.type(modelSelectorInput, 'test-model'); // Select a model

      const promptTextarea = screen.getByLabelText('Prompt');
      await userEvent.clear(promptTextarea); // Clear the prompt

      const generateButton = screen.getByRole('button', { name: 'Generate' });
      await userEvent.click(generateButton);

      expect(toast.error).toHaveBeenCalledWith('Prompt cannot be empty.');
      expect(mockGenerateOptions).not.toHaveBeenCalled();
    });

    it('calls generateOptions, onGenerationComplete, onClose, and shows success toast on successful generation', async () => {
      const mockSelectedModel = 'gpt-4';
      const mockPrompt = 'Generate some good options for MyDataset (A test dataset).';
      mockGenerateOptions.mockResolvedValueOnce(['Generated Option X', 'Generated Option Y']);

      renderDialog();

      const modelSelectorInput = screen.getByTestId('model-selector');
      await userEvent.type(modelSelectorInput, mockSelectedModel);

      const promptTextarea = screen.getByLabelText('Prompt');
      await userEvent.clear(promptTextarea);
      await userEvent.type(promptTextarea, mockPrompt);

      const generateButton = screen.getByRole('button', { name: 'Generate' });
      await userEvent.click(generateButton);

      expect(mockGenerateOptions).toHaveBeenCalledWith({ model: mockSelectedModel, prompt: mockPrompt });
      await waitFor(() => {
        expect(defaultProps.onGenerationComplete).toHaveBeenCalledWith(['Generated Option X', 'Generated Option Y']);
      });
      await waitFor(() => {
        expect(toast.success).toHaveBeenCalledWith('Options generated successfully!');
      });
      await waitFor(() => {
        expect(defaultProps.onClose).toHaveBeenCalled();
      });
    });

    it('shows error toast and does not call onGenerationComplete or onClose if generateOptions throws an error', async () => {
      const errorMessage = 'Simulated API Error';
      mockGenerateOptions.mockRejectedValueOnce(new Error(errorMessage));

      renderDialog();

      const modelSelectorInput = screen.getByTestId('model-selector');
      await userEvent.type(modelSelectorInput, 'test-model');
      const promptTextarea = screen.getByLabelText('Prompt');
      // Ensure prompt is not empty
      if (! (promptTextarea as HTMLTextAreaElement).value.trim()) {
        await userEvent.type(promptTextarea, 'A valid prompt');
      }

      const generateButton = screen.getByRole('button', { name: 'Generate' });
      await userEvent.click(generateButton);

      await waitFor(() => {
        expect(toast.error).toHaveBeenCalledWith(errorMessage);
      });
      expect(defaultProps.onGenerationComplete).not.toHaveBeenCalled();
      expect(defaultProps.onClose).not.toHaveBeenCalled(); // Dialog should stay open on error
    });

    it('disables Generate button during loading and re-enables after completion', async () => {
      // Make the mock take some time
      mockGenerateOptions.mockImplementationOnce(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
        return ['Delayed Option'];
      });

      renderDialog();
      const modelSelectorInput = screen.getByTestId('model-selector');
      await userEvent.type(modelSelectorInput, 'test-model');
      const promptTextarea = screen.getByLabelText('Prompt');
       if (! (promptTextarea as HTMLTextAreaElement).value.trim()) {
        await userEvent.type(promptTextarea, 'A valid prompt for loading test');
      }

      const generateButton = screen.getByRole('button', { name: 'Generate' });

      // fireEvent.click is synchronous, userEvent.click is asynchronous
      // Use userEvent for more realistic interaction simulation
      const clickPromise = userEvent.click(generateButton);

      // Immediately after click (before mock promise resolves), button should be disabled
      await waitFor(() => expect(generateButton).toBeDisabled());
      expect(screen.getByText('Generating...')).toBeInTheDocument(); // Check for loading text

      // Wait for the click promise to complete (which includes the mock resolving)
      await clickPromise;

      // After mock promise resolves, button should be re-enabled (because dialog closes)
      // Or, if we didn't close on success, it would be enabled.
      // Since it closes, we can check if onClose was called.
      await waitFor(() => expect(defaultProps.onClose).toHaveBeenCalled());
      // If the dialog were to stay open, we'd check:
      // await waitFor(() => expect(generateButton).not.toBeDisabled());
    });

    it('disables Generate button if prompt is empty or only whitespace', async () => {
      renderDialog();
      const modelSelectorInput = screen.getByTestId('model-selector');
      await userEvent.type(modelSelectorInput, 'test-model');

      const promptTextarea = screen.getByLabelText('Prompt');
      const generateButton = screen.getByRole('button', { name: 'Generate' });

      await userEvent.clear(promptTextarea);
      expect(generateButton).toBeDisabled();

      await userEvent.type(promptTextarea, '   ');
      expect(generateButton).toBeDisabled();

      await userEvent.type(promptTextarea, 'not empty');
      expect(generateButton).not.toBeDisabled();
    });


    it('calls onClose prop when Cancel button is clicked', async () => {
      renderDialog();
      const cancelButton = screen.getByRole('button', { name: 'Cancel' });
      await userEvent.click(cancelButton);
      expect(defaultProps.onClose).toHaveBeenCalled();
    });

    it('does not render if isOpen is false', () => {
      renderDialog({ isOpen: false });
      expect(screen.queryByText('Generate AI Options')).not.toBeInTheDocument();
    });
  });
});
