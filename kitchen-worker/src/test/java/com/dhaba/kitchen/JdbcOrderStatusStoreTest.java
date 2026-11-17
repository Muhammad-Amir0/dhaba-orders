package com.dhaba.kitchen;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.junit.jupiter.api.Test;
import org.springframework.jdbc.core.JdbcTemplate;

class JdbcOrderStatusStoreTest {
    @Test
    void updatesStatusForOrder() {
        JdbcTemplate jdbc = org.mockito.Mockito.mock(JdbcTemplate.class);
        when(jdbc.update("UPDATE orders SET status = ? WHERE id = ?", "cooking", 9L))
                .thenReturn(1);
        JdbcOrderStatusStore store = new JdbcOrderStatusStore(jdbc);

        store.updateStatus(9, "cooking");

        verify(jdbc).update("UPDATE orders SET status = ? WHERE id = ?", "cooking", 9L);
    }

    @Test
    void rejectsUnknownOrder() {
        JdbcTemplate jdbc = org.mockito.Mockito.mock(JdbcTemplate.class);
        when(jdbc.update("UPDATE orders SET status = ? WHERE id = ?", "ready", 99L))
                .thenReturn(0);
        JdbcOrderStatusStore store = new JdbcOrderStatusStore(jdbc);

        IllegalStateException exception = assertThrows(
                IllegalStateException.class,
                () -> store.updateStatus(99, "ready"));

        assertEquals("Order 99 does not exist", exception.getMessage());
    }
}
