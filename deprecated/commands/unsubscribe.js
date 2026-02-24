export async function unsubscribeCommand(prisma, bot, query) {
    const data = query.data;
    
    if (data.startsWith('unsubscribe_')) {
        const [, targetId, subscriberId, chatId] = data.split('_');
        console.log(chatId)

        await prisma.subscription.deleteMany({
            where: {
                targetId: parseInt(targetId),
                subscriberId: parseInt(subscriberId),
                chatId,
            }
        });

        await bot.answerCallbackQuery(query.id, { text: 'Вы успешно отписались от пользователя!' });
    }
}
