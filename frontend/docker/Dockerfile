FROM node:lts-alpine as builder

ARG BUILD_ENV=production

WORKDIR app
COPY ./app/package.json ./app/yarn.lock ./
RUN yarn install
COPY ./app .
RUN yarn build:${BUILD_ENV}

FROM nginx:1.19.0-alpine

RUN apk add --no-cache curl

COPY ./docker/snippets /etc/nginx/snippets
COPY ./docker/nginx.conf /etc/nginx/
COPY ./docker/conf.d /etc/nginx/conf.d
COPY  --from=builder /app/build /app

RUN rm /etc/nginx/conf.d/default.conf

HEALTHCHECK --interval=5s --timeout=3s --start-period=1s CMD curl --fail http://127.0.0.1/health || exit 1