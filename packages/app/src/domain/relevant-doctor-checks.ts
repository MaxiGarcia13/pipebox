import type { DoctorCheck, WorkflowProvider } from '@/domain/runner-events';

export function relevantDoctorChecks(
  checks: DoctorCheck[],
  providers: readonly WorkflowProvider[],
): DoctorCheck[] {
  const hasGithub = providers.includes('github');
  const hasJenkins = providers.includes('jenkins');
  return checks.filter((check) => {
    if (check.name === 'act_image') {
      return hasGithub;
    }
    if (check.name === 'jenkins_image') {
      return hasJenkins;
    }
    return true;
  });
}
