FROM node:20.14-alpine AS builder

WORKDIR /usr/src/app

COPY package*.json ./
RUN npm install

COPY . .
RUN npx prisma generate

FROM node:20.14-alpine AS production

WORKDIR /usr/src/app

COPY package*.json ./
RUN npm install --omit=dev

COPY --from=builder /usr/src/app ./
COPY --from=builder /usr/src/app/prisma ./prisma
COPY --from=builder /usr/src/app/.env ./

CMD ["node", "index.js"]