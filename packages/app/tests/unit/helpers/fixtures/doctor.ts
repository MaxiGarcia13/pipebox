import type { DoctorEvent } from '@/domain/runner-events';

export const okDoctor: DoctorEvent = {
  type: 'doctor',
  ok: true,
  version: '0.7.0',
  checks: [
    { name: 'docker', ok: true, path: '/usr/bin/docker' },
    { name: 'docker_daemon', ok: true, detail: 'reachable' },
    { name: 'act_image', ok: true, detail: 'pipebox-act:latest' },
    { name: 'jenkins_image', ok: true, detail: 'pipebox-jenkins:latest' },
  ],
};

export function doctorWithFailedCheck(
  name: 'docker' | 'docker_daemon' | 'act_image' | 'jenkins_image',
): DoctorEvent {
  const checks = okDoctor.checks.map((check) =>
    check.name === name
      ? { ...check, ok: false, detail: `${name} failed` }
      : check,
  );
  return {
    type: 'doctor',
    ok: false,
    version: '0.7.0',
    checks,
  };
}
