import { describe, expect, it, vi } from 'vitest';

import { createMockSimulationStreamClient } from '@/lib/api/mock/realtime';

describe('createMockSimulationStreamClient', () => {
  it('emits deterministic sequence values', () => {
    vi.useFakeTimers();

    const client = createMockSimulationStreamClient();
    const sequences: number[] = [];

    client.subscribe((event) => {
      sequences.push(event.sequence);
    });

    client.connect();
    vi.advanceTimersByTime(250);
    vi.advanceTimersByTime(1_500);
    vi.advanceTimersByTime(1_500);
    vi.advanceTimersByTime(1_500);

    expect(sequences.length).toBeGreaterThanOrEqual(3);
    expect(sequences.slice(0, 3)).toEqual([1, 2, 3]);

    client.disconnect();
    vi.useRealTimers();
  });
});
