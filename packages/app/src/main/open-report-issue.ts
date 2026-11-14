import { shell } from 'electron';

export async function openReportIssue(): Promise<void> {
  await shell.openExternal(__PIPEBOX_REPORT_ISSUE_URL__);
}
