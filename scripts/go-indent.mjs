import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';

const mode = process.argv[2];
const root = process.argv[3] ?? '.';

function* goFiles(dir) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.name === 'vendor')
      continue;
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory())
      yield* goFiles(fullPath);
    else if (entry.name.endsWith('.go'))
      yield fullPath;
  }
}

let failed = false;
for (const file of goFiles(root)) {
  const text = fs.readFileSync(file, 'utf8');
  if (!text.includes('\t'))
    continue;
  if (mode === '--fix') {
    fs.writeFileSync(file, text.replaceAll('\t', '  '));
    continue;
  }
  failed = true;
  for (const [i, line] of text.split(/\r?\n/).entries()) {
    if (line.includes('\t'))
      console.error(`${file}:${i + 1}`);
  }
}

if (failed) {
  console.error(
    'error: Go files must use 2-space indent (no tabs; see .editorconfig)',
  );
  process.exit(1);
}
