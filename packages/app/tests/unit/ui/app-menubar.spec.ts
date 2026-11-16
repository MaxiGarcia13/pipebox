import { flushPromises, mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import AppMenubar from '@/components/app-menubar.vue';
import { createMockPipebox } from '../helpers/mock-pipebox';

describe('app menubar', () => {
  it('calls reportIssue when the help item is clicked', async () => {
    const mock = createMockPipebox();
    window.pipebox = mock.api;

    const wrapper = mount(AppMenubar, {
      global: {
        stubs: {
          Menubar: false,
          MenubarMenu: false,
          MenubarTrigger: true,
          MenubarContent: {
            template: '<div><slot /></div>',
          },
          MenubarItem: {
            inheritAttrs: false,
            template:
              '<button :data-testid="$attrs[\'data-testid\']" @click="$emit(\'click\')"><slot /></button>',
          },
        },
      },
    });

    await wrapper.get('[data-testid="menubar-report-issue"]').trigger('click');
    await flushPromises();

    expect(mock.reportIssueCalls).toHaveBeenCalledOnce();
  });
});
