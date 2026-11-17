<?php

use App\Http\Controllers\StatsController;
use Illuminate\Support\Facades\Route;

Route::get('/stats', [StatsController::class, 'index']);
Route::get('/stats/healthz', [StatsController::class, 'health']);
Route::get('/stats/livez', [StatsController::class, 'live']);
Route::get('/stats/info', [StatsController::class, 'info']);
