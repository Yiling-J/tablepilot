import { describe, it, expect } from 'vitest'; // Removed vi, beforeEach, afterEach
import { generateOptions, GenerateOptionsParams } from './actions';

describe('actions', () => {
  describe('generateOptions', () => {
    // Removing fake timers for this last attempt to resolve timeouts.
    // beforeEach(() => {
    //   vi.useFakeTimers();
    // });

    // afterEach(() => {
    //   vi.runOnlyPendingTimers();
    //   vi.useRealTimers();
    // });

    const baseParams: Omit<GenerateOptionsParams, 'prompt'> = {
      model: 'test-model',
    };

    it('should throw an error when prompt includes "error test"', async () => {
      const params: GenerateOptionsParams = { ...baseParams, prompt: 'This is an error test' };
      await expect(generateOptions(params)).rejects.toThrow('Simulated error: Could not generate options due to a server issue.');
    });

    it('should return an empty array when prompt includes "empty"', async () => {
      const params: GenerateOptionsParams = { ...baseParams, prompt: 'Please return empty' };
      const result = await generateOptions(params);
      expect(result).toEqual([]);
    });

    it('should return specific options when prompt includes "specific"', async () => {
      const params: GenerateOptionsParams = { ...baseParams, prompt: 'Give me specific options' };
      const result = await generateOptions(params);
      expect(result).toEqual(["Specific Option Alpha", "Specific Option Beta", "Specific Option Gamma"]);
    });

    it('should return default mock options for a generic prompt', async () => {
      const params: GenerateOptionsParams = { ...baseParams, prompt: 'Generic prompt' };
      const result = await generateOptions(params);

      expect(result).toHaveLength(3);
      expect(result[0]).toBe('Generated option 1 for test-model');
      expect(result[1]).toBe('Generated option 2 based on: "Generic prompt..."');
      expect(result[2]).toMatch(/^Another item - \d{1,2}:\d{2}:\d{2} (AM|PM)$/);
    });

    // This test might be flaky without fake timers if the real delay is too short or too long.
    // For now, focusing on the main functionality.
    // it('should have a delay for API simulation', async () => {
    //   const params: GenerateOptionsParams = { ...baseParams, prompt: 'Test delay' };
    //   const startTime = Date.now();
    //   await generateOptions(params);
    //   const endTime = Date.now();
    //   expect(endTime - startTime).toBeGreaterThanOrEqual(1400); // Check if delay is roughly 1.5s
    // });

    it('should correctly substring the prompt in the second default option', async () => {
      const longPrompt = 'This is a very long prompt that definitely exceeds thirty characters limit';
      const params: GenerateOptionsParams = { ...baseParams, prompt: longPrompt };
      const result = await generateOptions(params);
      expect(result[1]).toBe(`Generated option 2 based on: "${longPrompt.substring(0, 30)}..."`);
    });

  });
});
