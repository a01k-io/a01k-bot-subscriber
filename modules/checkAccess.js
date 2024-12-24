import Redis from 'ioredis';
import dayjs from "dayjs";

const redis = new Redis({
    host: process.env.REDIS_HOST,
    port: process.env.REDIS_PORT,
    // username: process.env.REDIS_USERNAME,
    password: process.env.REDIS_PASSWORD,
});

export async function checkAccess(userId) {

    userId = 347283122;
    const cacheKey = `subscription:${userId}`;
    const cachedData = await redis.get(cacheKey);

    let userData;

    if (cachedData) {
        userData = JSON.parse(cachedData);
    } else {


        try {
            const req = await fetch(`https://bot.a01k.io/api/user/${userId}`);
            const response = await req.json();

            if (response.code === 404) {
                userData = {
                    access: false,
                }
            }
            if (response.alpha) {
                userData = {
                    access: true,
                    alpha: response.alpha
                }
            }
        } catch (e) {
            userData = {
                access: false
            }
        }

        await redis.set(cacheKey, JSON.stringify(userData), 'EX', 86000);

    }

    if (userData.access) {
        const now = dayjs();
        const start = dayjs(userData.alpha.date_start_subscription);
        const end = dayjs(userData.alpha.date_end_subscription);

        if (now.isAfter(start) && now.isBefore(end)) {
            return true
        }
    }

    return false;
}
