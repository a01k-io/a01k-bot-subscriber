
export const getUser = async (prisma,id) => {
    const telegramId = id.toString()
    let user = await prisma.user.findUnique({
        where: {telegramId}
    })

    if (!user) {
        user = await prisma.user.create({
            data: {
                username:  'str',
                telegramId
            }
        })
    }
    return user
}