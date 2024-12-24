import {getUser} from "../modules/getUser.js";

export async function subscribeCommand(prisma, bot, msg) {
    const chatId = msg.chat.id.toString();

    if (!msg.reply_to_message || !msg.reply_to_message.from) {
        return bot.sendMessage(chatId, "Пожалуйста, перешлите сообщение пользователя, на которого хотите подписаться и укажите /sub");
    }

    try {
        const targetUser = await getUser(prisma, msg.reply_to_message.from.id);
        const fromUser = await getUser(prisma, msg.from.id);


        if (!targetUser) {
            return bot.sendMessage(chatId, "Пользователь не найден.");
        }

        const existingSubscription = await prisma.subscription.findUnique({
            where: {
                subscriberId_targetId: {
                    subscriberId: fromUser.id,
                    targetId: targetUser.id,
                },
            },
        });

        if (existingSubscription) {
            return bot.sendMessage(chatId, "Вы уже подписаны на @" + targetUser.username);
        }


        if (fromUser.id === targetUser.id) {
            return bot.sendMessage(chatId, "Вы не можете подписаться сами на себя");
        }


        await prisma.subscription.create({
            data: {
                subscriberId: Number(fromUser.id),
                targetId: Number(targetUser.id),
                chatId
            },
        });

        const checkMsg = await bot.sendMessage(chatId, '✅', {
            reply_to_message_id: msg.message_id
        });

        setTimeout(async () => {
            try {
                await bot.deleteMessage(chatId, msg.message_id);
                await bot.deleteMessage(chatId, checkMsg.message_id);
            } catch (error) {
                console.error('Не удалось удалить сообщение(я): ', error);
            }
        }, 3000);
    } catch (error) {
        console.error(error);
        await bot.sendMessage(chatId, "Произошла ошибка при обработке запроса.");
    }
}
