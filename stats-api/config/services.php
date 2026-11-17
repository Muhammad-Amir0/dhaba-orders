<?php

return [
    'stats_redis' => [
        'host' => env('REDIS_HOST'),
        'port' => (int) env('REDIS_PORT', 6379),
        'password' => env('REDIS_PASSWORD'),
    ],
];
