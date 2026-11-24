import type { DoctorEvent } from '@/domain/runner-events';
import { ref } from 'vue';

export function useDoctorSession() {
  const doctor = ref<DoctorEvent | null>(null);

  function applyDoctor(next: DoctorEvent) {
    const version = next.version || doctor.value?.version || '';
    doctor.value = version === next.version
      ? next
      : { ...next, version };
  }

  async function runDoctor(repoPath?: string | null) {
    applyDoctor(await window.pipebox.doctor(repoPath));
  }

  return {
    doctor,
    applyDoctor,
    runDoctor,
  };
}
