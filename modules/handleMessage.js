import {getUser} from "./getUser.js";

export async function handleMessage(prisma, bot, msg) {
    const chatId = msg.chat.id.toString();

    try {
        const user = await getUser(prisma, msg.from.id);

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
            await bot.sendMessage(subscriber.subscriber.telegramId, `${msg.from.username} (${msg.chat.title}):\n${msg.text}`, {
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
            });
        }
    } catch (error) {
        console.error('Ошибка обработки сообщения:', error);
    }
}
