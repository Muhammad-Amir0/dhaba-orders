<?php

namespace App\Services;

use App\Contracts\StatsRedis;
use Carbon\CarbonImmutable;

final class StatsReader
{
    public function __construct(private readonly StatsRedis $redis)
    {
    }

    /**
     * @return array{
     *     orders_per_minute: int,
     *     top_dishes: list<array{name: string, count: int}>
     * }
     */
    public function read(): array
    {
        $minute = CarbonImmutable::now('UTC')->format('Y-m-d\TH:i\Z');
        $orders = $this->redis->zScore('stats:orders:minute', $minute);
        $topDishes = [];

        foreach ($this->redis->topMembers('stats:dishes', 5) as $name => $count) {
            $topDishes[] = [
                'name' => $name,
                'count' => (int) $count,
            ];
        }

        return [
            'orders_per_minute' => (int) ($orders ?? 0),
            'top_dishes' => array_slice($topDishes, 0, 5),
        ];
    }

    public function checkReady(): void
    {
        if (! $this->redis->ping()) {
            throw new RedisNotReadyException('Redis did not respond with PONG');
        }
    }
}
