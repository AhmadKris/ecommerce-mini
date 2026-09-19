import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 20 }, // Ramp up to 20 virtual users over 10s
    { duration: '30s', target: 50 }, // Stay at 50 virtual users for 30s
    { duration: '10s', target: 0 },  // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<200', 'p(99)<500'], // 95% of requests must finish under 200ms
    http_req_failed: ['rate<0.01'],               // Error rate must be less than 1%
  },
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080/api';

export default function () {
  // Scenario 1: Browse Public Catalog (Cache-aside hit path)
  const catalogRes = http.get(`${BASE_URL}/products?page=1&limit=10`);
  check(catalogRes, {
    'catalog status is 200': (r) => r.status === 200,
    'catalog response time < 100ms': (r) => r.timings.duration < 100,
  });

  sleep(1);

  // Scenario 2: Browse Category List
  const categoryRes = http.get(`${BASE_URL}/categories`);
  check(categoryRes, {
    'categories status is 200': (r) => r.status === 200,
  });

  sleep(1);

  // Scenario 3: Login User
  const loginPayload = JSON.stringify({
    email: 'budi@example.com',
    password: 'Password123!',
  });

  const loginParams = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, loginParams);
  const loginSuccess = check(loginRes, {
    'login status is 200 or 401': (r) => r.status === 200 || r.status === 401,
  });

  if (loginRes.status === 200) {
    const body = JSON.parse(loginRes.body);
    const token = body.data?.access_token;

    if (token) {
      const authParams = {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
      };

      // Scenario 4: Get Cart
      const cartRes = http.get(`${BASE_URL}/cart`, authParams);
      check(cartRes, {
        'get cart status is 200': (r) => r.status === 200,
      });
    }
  }

  sleep(1);
}
