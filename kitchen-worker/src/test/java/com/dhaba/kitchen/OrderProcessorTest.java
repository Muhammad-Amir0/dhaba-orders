package com.dhaba.kitchen;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.inOrder;
import static org.mockito.Mockito.mock;

import java.time.Duration;
import org.mockito.InOrder;
import org.junit.jupiter.api.Test;

class OrderProcessorTest {
    @Test
    void transitionsOrderFromCookingToReady() throws InterruptedException {
        OrderStatusStore store = mock(OrderStatusStore.class);
        OrderProcessor processor = new OrderProcessor(store, Duration.ZERO);

        processor.process(42);

        InOrder calls = inOrder(store);
        calls.verify(store).updateStatus(42, "cooking");
        calls.verify(store).updateStatus(42, "ready");
    }

    @Test
    void configuredCookingTimeMustBeWithinSupportedRange() {
        OrderStatusStore store = (orderId, status) -> { };

        assertThrows(IllegalArgumentException.class, () -> new OrderProcessor(store, 4));
        assertThrows(IllegalArgumentException.class, () -> new OrderProcessor(store, 11));
    }
}
