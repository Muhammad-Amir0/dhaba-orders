<?php

namespace Tests\Feature;

use App\Contracts\StatsRedis;
use Tests\Fakes\FakeStatsRedis;
use Tests\TestCase;

final class StatsEndpointTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();
        $this->app->instance(
            StatsRedis::class,
            new FakeStatsRedis(7, [
                'Chicken Biryani' => 12,
                'Chicken Karahi' => 9,
                'Plain Naan' => 6,
                'Seekh Kebab' => 4,
                'Mango Lassi' => 2,
                'Chapli Kebab' => 1,
            ]),
        );
    }

    public function test_stats_endpoint_returns_order_rate_and_top_five_dishes(): void
    {
        $response = $this->getJson('/api/stats');

        $response->assertOk()
            ->assertExactJson([
                'orders_per_minute' => 7,
                'top_dishes' => [
                    ['name' => 'Chicken Biryani', 'count' => 12],
                    ['name' => 'Chicken Karahi', 'count' => 9],
                    ['name' => 'Plain Naan', 'count' => 6],
                    ['name' => 'Seekh Kebab', 'count' => 4],
                    ['name' => 'Mango Lassi', 'count' => 2],
                ],
            ]);
    }

    public function test_health_endpoint_reports_ready_when_redis_is_available(): void
    {
        $response = $this->getJson('/api/stats/healthz');

        $response->assertOk()->assertExactJson(['status' => 'ok']);
    }

    public function test_liveness_endpoint_reports_ok_without_reading_stats(): void
    {
        $response = $this->getJson('/api/stats/livez');

        $response->assertOk()->assertExactJson(['status' => 'ok']);
    }

    public function test_info_endpoint_reports_version_and_pod(): void
    {
        config(['app.version' => 'commit-abc', 'app.pod' => 'stats-pod-1']);

        $response = $this->getJson('/api/stats/info');

        $response->assertOk()
        ->assertExactJson(['version' => 'commit-abc', 'pod' => 'stats-pod-1']);
    }

    public function test_health_endpoint_returns_service_unavailable_when_redis_is_not_ready(): void
    {
        $this->app->instance(StatsRedis::class, new FakeStatsRedis(available: false));

        $response = $this->getJson('/api/stats/healthz');

        $response->assertServiceUnavailable()
            ->assertExactJson(['status' => 'unavailable']);
    }

    public function test_stats_endpoint_returns_five_or_fewer_dishes(): void
    {
        $response = $this->getJson('/api/stats');

        $this->assertCount(5, $response->json('top_dishes'));
    }

    public function test_orders_per_minute_expectation(): void
    {
        $response = $this->getJson('/api/stats');

        $this->assertSame(99, $response->json('orders_per_minute'));
    }
}
