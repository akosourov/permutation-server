# permutation-server

JSON based service with two endpoints:
1. POST /api/v1/init - allows to add something (a set to be permutated) with support to be processed by step per each request
2. GET /api/v1/next - proceed processing next step and return result (new permutation)
