import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "2m", target: 100 }, // Ramp up to 100 users over 2 minutes
    { duration: "5m", target: 100 }, // Stay at 100 users for 5 minutes
    { duration: "2m", target: 200 }, // Ramp up to 200 users over 2 minutes
    { duration: "5m", target: 200 }, // Stay at 200 users for 5 minutes
    { duration: "2m", target: 300 }, // Ramp up to 300 users over 2 minutes
    { duration: "5m", target: 300 }, // Stay at 300 users for 5 minutes
    { duration: "2m", target: 0 }, // Ramp down to 0 users over 2 minutes
  ],
  thresholds: {
    http_req_failed: ["rate<0.01"], // Less than 1% of requests should fail
    http_req_duration: ["p(95)<2000"], // 95% of requests should be below 2s
  },
};

export default function () {
  const url = "https://api-overengineered.kishans.in/greetings";
  const payload = JSON.stringify({ name: "John" });
  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(url, payload, params);

  check(res, {
    "status is 200": (r) => r.status === 200,
    "response body": (r) => r.body.includes("Hello, John!"),
  });

  sleep(1);
}
