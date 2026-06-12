# Production client UI for k8s. Build from repo root:
#   docker build -f build/client-frontend.Dockerfile -t dealer-client-ui:latest .
FROM node:20-alpine AS build
WORKDIR /app
COPY frontend/client/package.json frontend/client/package-lock.json ./
RUN npm ci
COPY frontend/client/ ./
RUN npm run build

FROM nginx:1.27-alpine
COPY build/client-frontend.nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 3001
CMD ["nginx", "-g", "daemon off;"]
