import type { MockPipeboxHandle, MockPipeboxOptions } from './mock-pipebox';
import { flushPromises, mount } from '@vue/test-utils';
import App from '@/app.vue';
import { createMockPipebox } from './mock-pipebox';

export async function mountApp(options: MockPipeboxOptions = {}) {
  const mock = createMockPipebox(options);
  window.pipebox = mock.api;

  const wrapper = mount(App, {
    attachTo: document.body,
  });
  await flushPromises();

  return { wrapper, mock };
}

export type MountedApp = Awaited<ReturnType<typeof mountApp>>;

export async function openProject(
  mounted: MountedApp,
  mockOverride?: MockPipeboxHandle,
) {
  const mock = mockOverride ?? mounted.mock;
  await mounted.wrapper.get('[data-testid="open-project-empty"]').trigger('click');
  await flushPromises();
  return mock;
}
