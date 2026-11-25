// Copies the test cases shared between Go and TypeScript into dist/, so
// they ship in the npm package as @uug-ai/models/dist/fixtures/*.
const fs = require('fs');
const path = require('path');

const fixtures = {
  'redaction-plan-cases.json': '../../pkg/models/testdata/redaction_plan_cases.json',
};

const target = path.join(__dirname, 'dist', 'fixtures');
fs.mkdirSync(target, { recursive: true });
for (const [name, source] of Object.entries(fixtures)) {
  fs.copyFileSync(path.join(__dirname, source), path.join(target, name));
}
