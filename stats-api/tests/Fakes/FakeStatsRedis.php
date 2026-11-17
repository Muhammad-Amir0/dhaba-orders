<?php

namespace Tests\Fakes;

use App\Contracts\StatsRedis;

final class FakeStatsRedis implements StatsRedis
{
    public function __construct(
        public mixed $orders = null,
        public array $dishes = [],
        public bool $available = true,
    ) {
    }

    public function ping(): bool
    {
        return $this->available;
    }

    public function zScore(string $key, string $member): mixed
    {
        return $key === 'stats:orders:minute' ? $this->orders : null;
    }

    public function topMembers(string $key, int $limit): array
    {
        return $key === 'stats:dishes' ? array_slice($this->dishes, 0, $limit, true) : [];
    }
}
