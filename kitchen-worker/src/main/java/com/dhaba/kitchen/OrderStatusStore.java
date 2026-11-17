package com.dhaba.kitchen;

public interface OrderStatusStore {
    void updateStatus(long orderId, String status);
}
