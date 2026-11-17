package com.dhaba.kitchen;

import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

@Repository
public class JdbcOrderStatusStore implements OrderStatusStore {
    private final JdbcTemplate jdbcTemplate;

    public JdbcOrderStatusStore(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    @Override
    public void updateStatus(long orderId, String status) {
        int changed = jdbcTemplate.update(
                "UPDATE orders SET status = ? WHERE id = ?",
                status,
                orderId);
        if (changed != 1) {
            throw new IllegalStateException("Order " + orderId + " does not exist");
        }
    }
}
