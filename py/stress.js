const http = require('http');

const URL = 'http://localhost:8080/web';
const CONCURRENCY = 1000;
let success = 0;
let fail = 0;
let count = 0;

console.log('开始 1000 并发压测...');
console.time('压测时间');

for (let i = 0; i < CONCURRENCY; i++) {
  const req = http.get(URL, (res) => {
    if (res.statusCode === 200) success++;
    else fail++;
    finish();
  }).on('error', () => {
    fail++;
    finish();
  });
}

function finish() {
  count++;
  if (count === CONCURRENCY) {
    console.timeEnd('压测时间');
    console.log('总并发：' + CONCURRENCY);
    console.log('成功：' + success);
    console.log('失败：' + fail);
  }
}