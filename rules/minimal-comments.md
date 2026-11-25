# Minimal comments

Comment only when the code cannot say it clearly. Prefer better names over narration.
Types and signatures are the documentation — do not duplicate them in comments.

**Never** write:

- `/** ... */` block comments or JSDoc (`@param`, `@returns`, `@example`, etc.)
- Blocks that restate the function, type, or file name
- Input → output examples the reader can infer from the body
- "What this does" prose when the signature and implementation are obvious

**Do** comment (short `//` only):

- Non-obvious constraints, invariants, or trade-offs
- Why a surprising approach was chosen (not what the next line does)
- Cross-cutting intent that would otherwise be missed

```typescript
// BAD — JSDoc / block comment
/**
 * Map act exit codes to pass/fail.
 * @param code — process exit code
 */
export function toStepStatus(code: number): 'passed' | 'failed' {
  if (code === 0) {
    return 'passed';
  }
  return 'failed';
}

// BAD — narrating one-liner
/** Detect failed steps from log lines. */
export function hasFailedStep(lines: string[]): boolean {
  return lines.some((line) => line.includes('failure'));
}

// GOOD — no comment when the name and body are enough
export function stepPassed(code: number): boolean {
  return code === 0;
}

// GOOD — non-obvious constraint as //
export function stripAnsi(line: string, pattern: RegExp): string {
  // act prints ANSI; strip before matching step markers.
  return line.replace(pattern, '');
}
```
