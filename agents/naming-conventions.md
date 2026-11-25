# Naming conventions

- **Files and folders**: kebab-case (`workflow-list.vue`, `step-log.ts`)
- **Functions and local/const variables**: camelCase (`getWorkflowById`, `isRunning`)
- **Global constants**: UPPER_SNAKE_CASE (`DEFAULT_WORKFLOW_DIR`, `MAX_LOG_LINES`)
- **Names must be descriptive**: prefer clear intent over abbreviations or vague names (`data`, `temp`, `handle`, `fn`)

```typescript
// ❌ BAD
const MaxLogLines = 3;
const d = fetch();
function Get_Workflow() {}
function doStuff(x) {}
// file: WorkflowList.vue

// ✅ GOOD
const MAX_LOG_LINES = 5000;
const workflowList = fetch();
function getWorkflow() {}
function parseStepStatus(line) {}
// file: workflow-list.vue
```
