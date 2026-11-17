import type { DoctorEvent } from '@/domain/runner-events';
import { ref } from 'vue';

export function useDoctorSession() {
  const doctor = ref<DoctorEvent | null>(null);

  async function runDoctor() {
    doctor.value = await window.pipebox.doctor();
  }

  return {
    doctor,
    runDoctor,
  };
}
