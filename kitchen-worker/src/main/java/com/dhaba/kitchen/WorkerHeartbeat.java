package com.dhaba.kitchen;

import java.time.Instant;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.scheduling.annotation.EnableScheduling;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

@Component
@EnableScheduling
public class WorkerHeartbeat {
    private final StringRedisTemplate redis;

    public WorkerHeartbeat(StringRedisTemplate redis) {
        this.redis = redis;
    }

    @Scheduled(fixedRate = 10_000)
    public void writeHeartbeat() {
        redis.opsForValue().set("worker:heartbeat", Long.toString(Instant.now().getEpochSecond()));
    }
}
