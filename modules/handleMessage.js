import {getUser} from "./getUser.js";
import {checkAccess} from "./checkAccess.js";

export async function handleMessage(prisma, bot, msg) {
    const chatId = msg.chat.id.toString();

    try {
        const user = await getUser(prisma, msg.from);
        const subscriptions = await prisma.subscription.findMany({
            where: {
                targetId: user.id,
                chatId,
            },
            include: {
                subscriber: {
                    select: {
                        telegramId: true,
                        id: true,
                    },
                },
            },
        });

        for (const subscriber of subscriptions) {
            const inline = {
                reply_markup: {
                    inline_keyboard: [
                        [
                            {
                                text: '🔗',
                                url: `https://t.me/c/${msg.chat.username || msg.chat.id.toString().replace('-100', '')}/${msg.message_id}`,
                            },
                            {
                                text: '❌',
                                callback_data: `unsubscribe_${user.id}_${subscriber.subscriber.id}_${chatId}`,
                            },
                        ],
                    ],
                },
            }

            if (!await checkAccess(subscriber.subscriber.telegramId)) {
                continue
            }

            if (msg.text) {
                await bot.sendMessage(subscriber.subscriber.telegramId, `${msg.from.username} (${msg.chat.title}):\n${msg.text}`,inline);
            }
            if (msg.photo) {
                const fileId = msg.photo[msg.photo.length - 1].file_id;
                await bot.sendPhoto(subscriber.subscriber.telegramId, fileId, {
                    caption: `${msg.from.username} (${msg.chat.title}): ${msg.caption || 'отправил фото'}`,
                    inline
                });

            }
        }
    } catch (error) {
            console.error('Ошибка обработки сообщения', chatId, msg);
    }
}
