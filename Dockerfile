FROM node:20.14-alpine

RUN apk add --no-cache openssl

WORKDIR /app

COPY package*.json ./
COPY ./prisma prisma
RUN npm install

COPY . .

RUN npx prisma generate


CMD ["node", "index.js"]
