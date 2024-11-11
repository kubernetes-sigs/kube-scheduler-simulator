FROM node:16-alpine AS deps

RUN apk update && \
    apk upgrade && \
    apk add --no-cache make gcc g++ py-pip

WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile

FROM node:16-alpine AS builder
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN yarn build && yarn install --production --ignore-scripts --prefer-offline

FROM node:16-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NUXT_TELEMETRY_DISABLED=1
# Currnently, HOST_ENV is used to decide whether to import the TEMPLATE files of the k8s resources.
# ./components/lib/templates/*.yaml
ENV HOST_ENV=production
ENV BASE_URL=http://localhost:1212
ENV KUBE_API_SERVER_URL=http://localhost:3131

RUN addgroup -g 1001 -S nodejs
RUN adduser -S nuxtjs -u 1001

COPY --from=builder /app/nuxt.config.js ./
#COPY --from=builder ./app/server ./server/
COPY --from=builder ./app/package.json ./
COPY --from=builder ./app/node_modules ./node_modules/
COPY --from=builder ./app/.nuxt ./.nuxt/
COPY --from=builder ./app/static ./static/

USER nuxtjs

EXPOSE 3000

CMD ["yarn", "start"]
