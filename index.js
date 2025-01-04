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

    if (!await checkAccess(msg.from.id)) {
        return
    }
    await subscribeCommand(prisma, bot, msg);
});

bot.on('callback_query', async (query) => {
    console.log(query)
    if (!await checkAccess(query.from.id)) {
        return
    }
    await unsubscribeCommand(prisma, bot, query);
});


bot.on('message', async (msg) => {
    if (!await checkAccess(msg.from.id)) {
        return
    }

    await handleMessage(prisma, bot, msg);
});