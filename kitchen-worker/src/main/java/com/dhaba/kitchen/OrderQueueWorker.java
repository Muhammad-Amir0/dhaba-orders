package com.dhaba.kitchen;

import java.time.Duration;
import java.util.concurrent.atomic.AtomicBoolean;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Component;

@Component
public class OrderQueueWorker implements org.springframework.context.SmartLifecycle {
    private static final Logger LOGGER = LoggerFactory.getLogger(OrderQueueWorker.class);
    private static final Duration POP_TIMEOUT = Duration.ofSeconds(1);

    private final StringRedisTemplate redis;
    private final OrderProcessor processor;
    private final AtomicBoolean running = new AtomicBoolean();
    private Thread workerThread;

    public OrderQueueWorker(StringRedisTemplate redis, OrderProcessor processor) {
        this.redis = redis;
        this.processor = processor;
    }

    @Override
    public synchronized void start() {
        if (running.compareAndSet(false, true)) {
            workerThread = new Thread(this::pollQueue, "order-queue-worker");
            workerThread.start();
        }
    }

    private void pollQueue() {
        while (running.get()) {
            try {
                String job = redis.opsForList().rightPop("queue:orders", POP_TIMEOUT);
                if (job != null) {
                    processor.process(parseOrderId(job));
                }
            } catch (InterruptedException exception) {
                if (running.get()) {
                    LOGGER.warn("Order worker interrupted while running", exception);
                }
                Thread.currentThread().interrupt();
                return;
            } catch (RuntimeException exception) {
                LOGGER.error("Order job failed", exception);
            }
        }
    }

    private long parseOrderId(String job) {
        try {
            long orderId = Long.parseLong(job);
            if (orderId < 1) {
                throw new NumberFormatException("order id must be positive");
            }
            return orderId;
        } catch (NumberFormatException exception) {
            throw new IllegalArgumentException("Invalid order id in queue: " + job, exception);
        }
    }

    @Override
    public synchronized void stop() {
        running.set(false);
        if (workerThread != null) {
            try {
                workerThread.join();
            } catch (InterruptedException exception) {
                Thread.currentThread().interrupt();
                throw new IllegalStateException("Interrupted while waiting for order worker", exception);
            }
        }
    }

    @Override
    public void stop(Runnable callback) {
        stop();
        callback.run();
    }

    @Override
    public boolean isRunning() {
        return running.get();
    }

    @Override
    public boolean isAutoStartup() {
        return true;
    }

    @Override
    public int getPhase() {
        return Integer.MAX_VALUE;
    }
}
