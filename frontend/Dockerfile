# syntax=docker/dockerfile:1
# Build context is the repository root so the shared design-system/ tokens are available.
# The build output is static files, identical for every CPU, so the build stage
# runs on the builder's native platform; only the nginx stage is per-platform.
FROM --platform=$BUILDPLATFORM node:24-alpine AS build
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY design-system /app/design-system
COPY frontend ./
RUN npm run build

FROM nginx:1.29-alpine
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/frontend/dist /usr/share/nginx/html
EXPOSE 80
