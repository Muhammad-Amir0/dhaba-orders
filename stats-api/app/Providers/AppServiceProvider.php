<?php

namespace App\Providers;

use App\Contracts\StatsRedis;
use App\Services\PredisStatsRedis;
use Illuminate\Support\ServiceProvider;
use Predis\Client;

final class AppServiceProvider extends ServiceProvider
{
    public function register(): void
    {
        $this->app->singleton(StatsRedis::class, function (): StatsRedis {
            $parameters = [
                'scheme' => 'tcp',
                'host' => config('services.stats_redis.host'),
                'port' => config('services.stats_redis.port'),
            ];
            $password = config('services.stats_redis.password');
            if (is_string($password) && $password !== '') {
                $parameters['password'] = $password;
            }

            return new PredisStatsRedis(new Client($parameters));
        });
    }
}
