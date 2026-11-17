import json
import logging
import os
import time

import psycopg
import redis
from fastapi import FastAPI, HTTPException
from psycopg.rows import dict_row

logger = logging.getLogger(__name__)
app = FastAPI(title="Dhaba Menu API")

MENU_CACHE_KEY = "menu:all"
MENU_CACHE_TTL_SECONDS = 60


def postgres_connection() -> psycopg.Connection:
    return psycopg.connect(
        host=os.environ["POSTGRES_HOST"],
        port=int(os.environ["POSTGRES_PORT"]),
        dbname=os.environ["POSTGRES_DB"],
        user=os.environ["POSTGRES_USER"],
        password=os.environ["POSTGRES_PASSWORD"],
        row_factory=dict_row,
    )


def redis_client() -> redis.Redis:
    return redis.Redis(
        host=os.environ["REDIS_HOST"],
        port=int(os.environ["REDIS_PORT"]),
        password=os.getenv("REDIS_PASSWORD") or None,
        decode_responses=True,
    )


def load_menu_from_db() -> list[dict]:
    with postgres_connection() as connection:
        with connection.cursor() as cursor:
            cursor.execute(
                "SELECT id, name, price_pkr, category "
                "FROM menu_items ORDER BY category, name"
            )
            return cursor.fetchall()


def check_dependencies() -> None:
    with postgres_connection() as connection:
        with connection.cursor() as cursor:
            cursor.execute("SELECT 1")
    redis_client().ping()


@app.get("/api/menu")
def get_menu() -> dict:
    cache = redis_client()
    cached_menu = cache.get(MENU_CACHE_KEY)
    if cached_menu is not None:
        return {"source": "cache", "items": json.loads(cached_menu)}

    items = load_menu_from_db()
    cache.set(MENU_CACHE_KEY, json.dumps(items), ex=MENU_CACHE_TTL_SECONDS)
    return {"source": "db", "items": items}


@app.get("/api/menu/healthz")
def healthz() -> dict:
    try:
        check_dependencies()
    except (psycopg.Error, redis.RedisError) as error:
        logger.exception("Menu API dependency readiness check failed")
        raise HTTPException(status_code=503, detail="dependency unavailable") from error

    return {
        "status": "ok",
        "version": os.getenv("APP_VERSION", "dev"),
        "pod": os.environ["HOSTNAME"],
    }


@app.get("/api/menu/livez")
def livez() -> dict:
    return {"status": "ok"}


@app.get("/api/menu/info")
def info() -> dict:
    return {
        "version": os.getenv("APP_VERSION", "dev"),
        "pod": os.environ["HOSTNAME"],
    }
