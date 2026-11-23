import { describe, expect, it } from 'vitest';
import { relevantDoctorChecks } from '@/domain/relevant-doctor-checks';
import { doctorWithFailedCheck, okDoctor } from '../helpers/fixtures';

describe('relevantDoctorChecks', () => {
  it('keeps docker checks regardless of providers', () => {
    const checks = relevantDoctorChecks(okDoctor.checks, []);
    expect(checks.map((check) => check.name)).toEqual([
      'docker',
      'docker_daemon',
    ]);
  });

  it('includes act_image only for github providers', () => {
    const failed = doctorWithFailedCheck('act_image');
    expect(
      relevantDoctorChecks(failed.checks, ['jenkins']).some(
        (check) => check.name === 'act_image',
      ),
    ).toBe(false);
    expect(
      relevantDoctorChecks(failed.checks, ['github']).some(
        (check) => check.name === 'act_image',
      ),
    ).toBe(true);
  });

  it('includes jenkins_image only for jenkins providers', () => {
    const failed = doctorWithFailedCheck('jenkins_image');
    expect(
      relevantDoctorChecks(failed.checks, ['github']).some(
        (check) => check.name === 'jenkins_image',
      ),
    ).toBe(false);
    expect(
      relevantDoctorChecks(failed.checks, ['jenkins']).some(
        (check) => check.name === 'jenkins_image',
      ),
    ).toBe(true);
  });
});
