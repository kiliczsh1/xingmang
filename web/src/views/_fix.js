const fs = require('fs');
const f = 'ExperienceShares.vue';
let s = fs.readFileSync(f, 'utf8');
const target = s.split('[^\w一-龥]').join('[^\w\u4e00-\u9fa5]');
fs.writeFileSync(f, target);
const lines = fs.readFileSync(f, 'utf8').split('\n');
console.log('LINE:', JSON.stringify(lines[1165]));
