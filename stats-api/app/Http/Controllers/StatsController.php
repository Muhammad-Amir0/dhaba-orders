<?php

namespace App\Http\Controllers;

use App\Services\StatsReader;
use App\Services\RedisNotReadyException;
use Illuminate\Http\JsonResponse;
use Illuminate\Support\Facades\Log;
use Predis\Connection\ConnectionException;
use Predis\Response\ServerException;

final class StatsController
{
    public function __construct(private readonly StatsReader $stats)
    {
    }

    public function index(): JsonResponse
    {
        return response()->json($this->stats->read());
    }

    public function health(): JsonResponse
    {
        try {
            $this->stats->checkReady();
        } catch (ConnectionException|ServerException|RedisNotReadyException $exception) {
            Log::error('Stats API Redis readiness check failed', ['exception' => $exception]);

            return response()->json(['status' => 'unavailable'], 503);
        }

        return response()->json(['status' => 'ok']);
    }

    public function live(): JsonResponse
    {
        return response()->json(['status' => 'ok']);
    }

    public function info(): JsonResponse
    {
        return response()->json([
            'version' => config('app.version', 'dev'),
            'pod' => (string) config('app.pod', ''),
        ]);
    }
}
