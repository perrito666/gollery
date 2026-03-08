import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { ComponentRegistry } from '../src/ui-contract/interfaces.js';

describe('ComponentRegistry', () => {
  it('register and get', () => {
    const registry = new ComponentRegistry();
    const view = { render() {}, destroy() {} };
    registry.register('home', view);
    assert.equal(registry.get('home'), view);
  });

  it('returns null for unregistered view', () => {
    const registry = new ComponentRegistry();
    assert.equal(registry.get('missing'), null);
  });

  it('override replaces existing', () => {
    const registry = new ComponentRegistry();
    const original = { render() {} };
    const replacement = { render() {} };

    registry.register('home', original);
    registry.override('home', replacement);
    assert.equal(registry.get('home'), replacement);
  });

  it('multiple views coexist', () => {
    const registry = new ComponentRegistry();
    const a = { render() {} };
    const b = { render() {} };
    registry.register('home', a);
    registry.register('album', b);
    assert.equal(registry.get('home'), a);
    assert.equal(registry.get('album'), b);
  });
});
