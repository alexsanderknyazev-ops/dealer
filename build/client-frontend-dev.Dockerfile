# Dev UI for client zone. Build from repo root:
#   docker build -f build/client-frontend-dev.Dockerfile -t dealer-client-ui .
#   ./scripts/run-client-frontend.sh
FROM node:20-alpine
WORKDIR /app
COPY frontend/client/package.json frontend/client/package-lock.json ./
RUN npm ci
COPY frontend/client/ ./
ENV VITE_CLIENT_PUBLIC_GATEWAY=http://127.0.0.1:8091
ENV VITE_CLIENT_PROTECTED_GATEWAY=http://127.0.0.1:8093
EXPOSE 3001
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3001"]
