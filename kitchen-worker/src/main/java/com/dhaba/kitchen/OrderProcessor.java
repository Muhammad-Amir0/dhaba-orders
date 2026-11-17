package com.dhaba.kitchen;

import java.time.Duration;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

@Service
public class OrderProcessor {
    private final OrderStatusStore statusStore;
    private final Duration cookingTime;

    @Autowired
    public OrderProcessor(
            OrderStatusStore statusStore,
            @Value("${kitchen.cooking-time-seconds:7}") long cookingTimeSeconds) {
        if (cookingTimeSeconds < 5 || cookingTimeSeconds > 10) {
            throw new IllegalArgumentException("Cooking time must be between 5 and 10 seconds");
        }
        this.statusStore = statusStore;
        this.cookingTime = Duration.ofSeconds(cookingTimeSeconds);
    }

    OrderProcessor(OrderStatusStore statusStore, Duration cookingTime) {
        this.statusStore = statusStore;
        this.cookingTime = cookingTime;
    }

    public void process(long orderId) throws InterruptedException {
        statusStore.updateStatus(orderId, "cooking");
        Thread.sleep(cookingTime.toMillis());
        statusStore.updateStatus(orderId, "ready");
    }
}
