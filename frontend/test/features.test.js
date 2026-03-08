import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { FeatureFlags } from '../src/core/services/features.js';

describe('FeatureFlags', () => {
  it('has default flags', () => {
    const flags = new FeatureFlags();
    assert.equal(flags.isEnabled('popularity'), false);
  });

  it('unknown flags return false', () => {
    const flags = new FeatureFlags();
    assert.equal(flags.isEnabled('nonexistent'), false);
  });

  it('siteConfig enables features', () => {
    const flags = new FeatureFlags({ features: { popularity: true } });
    assert.equal(flags.isEnabled('popularity'), true);
  });

  it('siteConfig ignores unknown features', () => {
    const flags = new FeatureFlags({ features: { unknown: true } });
    assert.equal(flags.isEnabled('unknown'), false);
  });

  it('enable() sets flag at runtime', () => {
    const flags = new FeatureFlags();
    assert.equal(flags.isEnabled('popularity'), false);
    flags.enable('popularity');
    assert.equal(flags.isEnabled('popularity'), true);
  });

  it('disable() clears flag at runtime', () => {
    const flags = new FeatureFlags({ features: { popularity: true } });
    flags.disable('popularity');
    assert.equal(flags.isEnabled('popularity'), false);
  });
});
