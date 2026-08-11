FROM node:22-alpine AS build

WORKDIR /app

ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

COPY fe/package.json fe/yarn.lock ./
RUN yarn install --frozen-lockfile

COPY fe/ ./
RUN yarn build

FROM nginxinc/nginx-unprivileged:1-alpine3.24

COPY fe/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build --chown=nginx:nginx /app/build /usr/share/nginx/html

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
	CMD wget -q --spider http://127.0.0.1:8080/healthz || exit 1
