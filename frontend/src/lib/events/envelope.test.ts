import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

import {
  validateEnvelopeForPublish,
  type EventEnvelope,
} from './envelope';
import type {
  EventName,
  EventNameToPayload,
} from './events';

type FixtureEnvelope = EventEnvelope<EventNameToPayload[EventName]> & { eventName: EventName };

const fixturePath = resolve(fileURLToPath(new URL('../../../../shared/events/__fixtures__/envelopes.json', import.meta.url)));
const fixtures = JSON.parse(readFileSync(fixturePath, 'utf8')) as FixtureEnvelope[];

describe('event envelope contract', () => {
  it('round-trips shared fixture envelopes', () => {
    expect(fixtures).toHaveLength(2);

    for (const fixture of fixtures) {
      const decoded = JSON.parse(JSON.stringify(fixture)) as FixtureEnvelope;

      expect(decoded.eventId).toBe(fixture.eventId);
      expect(decoded.eventName).toBe(fixture.eventName);
      expect(decoded.eventVersion).toBe(fixture.eventVersion);
      expect(decoded.aggregateType).toBe(fixture.aggregateType);
      expect(decoded.aggregateId).toBe(fixture.aggregateId);
      expect(decoded.occurredAt).toBe(fixture.occurredAt);
      expect(decoded.producer).toBe(fixture.producer);
      expect(decoded.traceId).toBe(fixture.traceId);
      expect(decoded.payload).toEqual(fixture.payload);
    }
  });

  it('treats traceId as soft-required for publish', () => {
    const missing = fixtures.find((fixture) => fixture.eventName === 'report.generated');
    expect(missing).toBeDefined();
    expect(validateEnvelopeForPublish(missing!).map((warning) => warning.field)).toEqual(['traceId']);

    const present = fixtures.find((fixture) => fixture.eventName === 'target.import.requested');
    expect(present?.traceId).toBe('trace-target-import-1');
    expect(validateEnvelopeForPublish(present!)).toEqual([]);
  });
});
