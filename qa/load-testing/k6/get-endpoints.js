/**
 * k6 load test for all employee-gateway GET endpoints.
 *
 * Prerequisites: k6 installed, endpoints manifest from probe script.
 *
 *   BASE_URL=http://192.168.0.27:9080 \
 *   LOGIN_EMAIL=qa.master@test.local LOGIN_PASSWORD='Test1234!' \
 *   k6 run -e PROFILE=smoke qa/load-testing/k6/get-endpoints.js
 *
 * Profiles (env PROFILE):
 *   smoke  — 5 VUs, 30s
 *   load   — 50 VUs, 5m
 *   stress — 100 VUs, 10m
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = (__ENV.BASE_URL || 'http://192.168.0.27:9080').replace(/\/$/, '');
const LOGIN_EMAIL = __ENV.LOGIN_EMAIL || 'qa.master@test.local';
const LOGIN_PASSWORD = __ENV.LOGIN_PASSWORD || 'Test1234!';
const MANIFEST_PATH = __ENV.MANIFEST_PATH || '../results/latest-endpoints.json';

const profiles = {
  smoke: { vus: 5, duration: '30s' },
  load: { vus: 50, duration: '5m' },
  stress: { vus: 100, duration: '10m' },
};
const profile = profiles[__ENV.PROFILE || 'smoke'] || profiles.smoke;

export const options = {
  vus: profile.vus,
  duration: profile.duration,
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<2000'],
  },
};

// Load manifest at init time (k6 open() is relative to script file)
const manifest = JSON.parse(open(MANIFEST_PATH));
const endpoints = manifest.endpoints || [];

// Weighted pool: repeat entries by weight for random selection
const weightedPool = [];
for (const ep of endpoints) {
  const w = ep.weight || 1;
  for (let i = 0; i < w; i++) weightedPool.push(ep);
}

export function setup() {
  const loginRes = http.post(
    `${BASE_URL}/api/login`,
    JSON.stringify({ email: LOGIN_EMAIL, password: LOGIN_PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  check(loginRes, { 'login status 200': (r) => r.status === 200 });
  const body = loginRes.json();
  if (!body || !body.access_token) {
    throw new Error(`login failed: ${loginRes.status} ${loginRes.body}`);
  }
  return { token: body.access_token };
}

export default function (data) {
  const ep = randomItem(weightedPool);
  const headers = { Accept: 'application/json' };
  if (ep.auth) {
    headers['Authorization'] = `Bearer ${data.token}`;
  }
  const res = http.get(`${BASE_URL}${ep.path}`, { headers, tags: { endpoint: ep.id } });
  check(res, {
    [`${ep.id} status 200`]: (r) => r.status === 200,
  });
  sleep(0.1 + Math.random() * 0.3);
}
