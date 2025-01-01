
export const getUser = async (prisma,user) => {

    const telegramId = user.id.toString()
    let userModel = await prisma.user.findUnique({
        where: {telegramId}
    })

    if (!userModel) {
        userModel = await prisma.user.create({
            data: {
                username: user.username ?? null,
                telegramId
            }
        })
    }
    return userModel
}