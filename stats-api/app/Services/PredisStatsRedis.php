<?php

namespace App\Services;

use App\Contracts\StatsRedis;
use Predis\Client;

final class PredisStatsRedis implements StatsRedis
{
    public function __construct(private readonly Client $client)
    {
    }

    public function ping(): bool
    {
        return (string) $this->client->ping() === 'PONG';
    }

    public function zScore(string $key, string $member): mixed
    {
        return $this->client->zscore($key, $member);
    }

    public function topMembers(string $key, int $limit): array
    {
        /** @var array<string, mixed> $members */
        $members = $this->client->zrevrange($key, 0, $limit - 1, ['withscores' => true]);

        return $members;
    }
}
