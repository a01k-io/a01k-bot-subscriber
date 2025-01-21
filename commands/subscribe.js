import {getUser} from "../modules/getUser.js";
import {deleteMessages} from "../modules/deleteMessages.js";

export async function subscribeCommand(prisma, bot, msg) {

    const chatId = msg.chat.id.toString();

    if (!msg.reply_to_message || !msg.reply_to_message.from) {
        const errorMsg = await  bot.sendMessage(chatId, "Пожалуйста, ответьте на сообщение пользователя, на которого хотите подписаться и укажите /sub");
        await deleteMessages(bot,chatId,[msg.message_id,errorMsg.message_id])
        return;

    }

    try {
        const targetUser = await getUser(prisma, msg.reply_to_message.from);
        const fromUser = await getUser(prisma, msg.from);


        if (!targetUser) {
            const errorMsg = await bot.sendMessage(chatId, "Пользователь не найден.");
            await deleteMessages(bot,chatId,[msg.message_id,errorMsg.message_id])
            return;
        }

        const existingSubscription = await prisma.subscription.findFirst({
            where: {
                    chatId,
                    subscriberId: fromUser.id,
                    targetId: targetUser.id,
            },
        });

        if (existingSubscription) {
            const errorMsg = await bot.sendMessage(chatId, "Вы уже подписаны на @" + targetUser.username);
            await deleteMessages(bot,chatId,[msg.message_id,errorMsg.message_id])
            return;
        }

        if (fromUser.id === targetUser.id) {
            const errorMsg = await bot.sendMessage(chatId, "Вы не можете подписаться сами на себя");
            await deleteMessages(bot,chatId,[msg.message_id,errorMsg.message_id])
            return;
        }

        try {
            await bot.sendMessage(fromUser.id, 'Вы подписались на @' + targetUser.username);
        } catch (error) {
            await bot.sendMessage(chatId, 'Чтобы подписаться, вам нужно перейти в личные сообщения с ботом @a01k_sub_bot и запустить его!');
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

        await deleteMessages(bot,chatId,[msg.message_id,checkMsg.message_id])

    } catch (error) {
        console.error(error);
        await bot.sendMessage(chatId, "Произошла ошибка при обработке запроса.");
    }
}
