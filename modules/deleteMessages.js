export async function deleteMessages(bot, chatId, messageIds) {
    setTimeout(async () => {
        try {
            for (const messageId of messageIds) {
                await bot.deleteMessage(chatId, messageId);
            }
        } catch (error) {
            console.error('Не удалось удалить сообщения: ', error);
        }
    }, 3000);
}