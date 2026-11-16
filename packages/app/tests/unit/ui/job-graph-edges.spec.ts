import type { JobSummary } from '@/domain/run';
import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { buildJobGraphConnector } from '@/components/job-graph-edges';
import JobTimeline from '@/components/job-timeline.vue';

describe('buildJobGraphConnector', () => {
  it('draws a centered bead on a one-to-one link', () => {
    const connector = buildJobGraphConnector(
      [
        { id: 'test', row: 0, needs: [] },
      ],
      [
        {
          id: 'e2e',
          row: 0,
          needs: ['test'],
        },
      ],
    );

    expect(connector).toEqual({
      width: 40,
      height: 40,
      lines: ['M0 20H40'],
      dots: [
        { x: 20, y: 20 },
      ],
    });
  });

  it('merges several jobs into the next one', () => {
    const connector = buildJobGraphConnector(
      [
        { id: 'test', row: 0, needs: [] },
        { id: 'lint', row: 1, needs: [] },
      ],
      [
        {
          id: 'e2e',
          row: 0,
          needs: ['test', 'lint'],
        },
      ],
    );

    expect(connector).toEqual({
      width: 40,
      height: 96,
      lines: [
        'M0 20H20',
        'M0 76H20',
        'M20 48H40',
        'M20 20V76',
      ],
      dots: [
        { x: 20, y: 48 },
      ],
    });
  });

  it('splits one job toward several successors', () => {
    const connector = buildJobGraphConnector(
      [
        { id: 'build', row: 0, needs: [] },
      ],
      [
        {
          id: 'test',
          row: 0,
          needs: ['build'],
        },
        {
          id: 'lint',
          row: 1,
          needs: ['build'],
        },
      ],
    );

    expect(connector.lines).toEqual([
      'M0 48H20',
      'M20 20H40',
      'M20 76H40',
      'M20 20V76',
    ]);
    expect(connector.dots).toEqual([
      { x: 20, y: 48 },
    ]);
  });

  it('keeps independent rows on their own links', () => {
    const connector = buildJobGraphConnector(
      [
        { id: 'test', row: 0, needs: [] },
        { id: 'lint', row: 1, needs: [] },
      ],
      [
        {
          id: 'e2e',
          row: 0,
          needs: ['test'],
        },
        {
          id: 'types',
          row: 1,
          needs: ['lint'],
        },
      ],
    );

    expect(connector.lines).toEqual([
      'M0 20H40',
      'M0 76H40',
    ]);
    expect(connector.dots).toEqual([
      { x: 20, y: 20 },
      { x: 20, y: 76 },
    ]);
  });

  it('offsets buses when links would cross', () => {
    const connector = buildJobGraphConnector(
      [
        { id: 'a', row: 0, needs: [] },
        { id: 'b', row: 1, needs: [] },
      ],
      [
        {
          id: 'c',
          row: 1,
          needs: ['a'],
        },
        {
          id: 'd',
          row: 0,
          needs: ['b'],
        },
      ],
    );

    expect(connector.dots).toEqual([
      { x: 12, y: 48 },
      { x: 28, y: 48 },
    ]);
    expect(connector.lines).toEqual([
      'M0 20H12',
      'M12 76H40',
      'M12 20V76',
      'M0 76H28',
      'M28 20H40',
      'M28 20V76',
    ]);
  });

  it('skips dependencies that are not in the previous column', () => {
    const connector = buildJobGraphConnector(
      [
        { id: 'test', row: 0, needs: [] },
      ],
      [
        {
          id: 'e2e',
          row: 0,
          needs: ['build'],
        },
      ],
    );

    expect(connector.lines).toEqual([]);
    expect(connector.dots).toEqual([]);
    expect(connector.height).toBe(40);
  });
});

describe('job timeline connectors', () => {
  it('draws a merge bead between dependent jobs', () => {
    const jobs: JobSummary[] = [
      {
        id: 'test',
        name: 'ci / test',
        needs: [],
        status: 'passed',
        durationMs: 34000,
      },
      {
        id: 'lint',
        name: 'ci / lint',
        needs: [],
        status: 'passed',
        durationMs: 33000,
      },
      {
        id: 'e2e',
        name: 'ci / e2e',
        needs: ['test', 'lint'],
        status: 'pending',
      },
    ];
    const wrapper = mount(JobTimeline, {
      props: {
        jobs,
        selectedId: 'test',
      },
    });

    const connector = wrapper.get('[data-testid="job-graph-connector"]');
    expect(connector.attributes('width')).toBe('40');
    expect(connector.attributes('height')).toBe('96');
    expect(connector.find('path').attributes('d')).toBe('M0 20H20 M0 76H20 M20 48H40 M20 20V76');
    expect(connector.find('circle').attributes('cy')).toBe('48');
    expect(wrapper.findAll('button')).toHaveLength(3);
  });
});
