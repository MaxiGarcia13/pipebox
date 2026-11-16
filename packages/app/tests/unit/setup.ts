import { config } from '@vue/test-utils';
import { afterEach } from 'vitest';

config.global.renderStubDefaultSlot = true;

afterEach(() => {
  document.body.innerHTML = '';
});
