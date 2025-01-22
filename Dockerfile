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


RUN npm install pm2 -g
ENV PM2_PUBLIC_KEY cwpw2ym276smyww
ENV PM2_SECRET_KEY ene1cub62a8sv8u

COPY --from=builder /usr/src/app ./
COPY --from=builder /usr/src/app/prisma ./prisma
COPY --from=builder /usr/src/app/.env ./

CMD ["pm2-runtime", "index.js"]