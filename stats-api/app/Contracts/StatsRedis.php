<?php

namespace App\Contracts;

interface StatsRedis
{
    public function ping(): bool;

    public function zScore(string $key, string $member): mixed;

    /**
     * @return array<string, mixed>
     */
    public function topMembers(string $key, int $limit): array;
}
