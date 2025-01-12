import { PrismaClient } from '@prisma/client';
import TelegramBot from 'node-telegram-bot-api';
import { subscribeCommand, unsubscribeCommand } from './commands/index.js';
import {handleMessage} from "./modules/handleMessage.js";
import dotenv from 'dotenv';
import {checkAccess} from "./modules/checkAccess.js";
const prisma = new PrismaClient();
const BOT_TOKEN = process.env.BOT_TOKEN;
const bot = new TelegramBot(BOT_TOKEN, { polling: true });


dotenv.config();
await bot.setMyCommands([
    { command: '/sub', description: 'Подписаться на сообщение пользователя' },
]);

bot.onText(/\/sub/, async (msg) => {

    // if (!await checkAccess(msg.from.id)) {
    //     return
    // }
    await subscribeCommand(prisma, bot, msg);
});

bot.on('callback_query', async (query) => {
    console.log(query)
    if (!await checkAccess(query.from.id)) {
        return
    }
    await unsubscribeCommand(prisma, bot, query);
});


bot.onText(/\/start/, async (msg) => {
    console.log(msg)
    if (!await checkAccess(msg.from.id)) {
        return
    }
    await bot.sendMessage(msg.chat.id, 'Привет! Я бот, который поможет тебе подписаться на сообщения пользователей. Для того, чтобы подписаться на сообщения пользователя, напиши /sub в нужном чате.');
});

bot.on('message', async (msg) => {
    await handleMessage(prisma, bot, msg);
});