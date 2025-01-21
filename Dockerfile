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

RUN pm2 link ene1cub62a8sv8u cwpw2ym276smyww


COPY --from=builder /usr/src/app ./
COPY --from=builder /usr/src/app/prisma ./prisma
COPY --from=builder /usr/src/app/.env ./

CMD ["pm2", "start", "app.js", "--no-daemon"]