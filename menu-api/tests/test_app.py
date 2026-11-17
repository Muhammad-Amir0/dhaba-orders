import json

from fastapi.testclient import TestClient

import app as menu_app

client = TestClient(menu_app.app)


class FakeRedis:
    def __init__(self, value=None):
        self.value = value
        self.saved = None

    def get(self, key):
        assert key == "menu:all"
        return self.value

    def set(self, key, value, ex):
        self.saved = (key, value, ex)

    def ping(self):
        return True


def test_menu_cache_hit(monkeypatch):
    expected_items = [{"id": 1, "name": "Chicken Karahi", "price_pkr": 1450, "category": "Karahi"}]
    cache = FakeRedis(json.dumps(expected_items))
    monkeypatch.setattr(menu_app, "redis_client", lambda: cache)
    monkeypatch.setattr(
        menu_app,
        "load_menu_from_db",
        lambda: (_ for _ in ()).throw(AssertionError("database should not be queried")),
    )

    response = client.get("/api/menu")

    assert response.status_code == 200
    assert response.json() == {"source": "cache", "items": expected_items}


def test_menu_cache_miss_reads_database_and_caches(monkeypatch):
    expected_items = [{"id": 2, "name": "Plain Naan", "price_pkr": 70, "category": "Bread"}]
    cache = FakeRedis()
    monkeypatch.setattr(menu_app, "redis_client", lambda: cache)
    monkeypatch.setattr(menu_app, "load_menu_from_db", lambda: expected_items)

    response = client.get("/api/menu")

    assert response.status_code == 200
    assert response.json() == {"source": "db", "items": expected_items}
    assert cache.saved == ("menu:all", json.dumps(expected_items), 60)


def test_health_checks_dependencies_and_reports_version(monkeypatch):
    monkeypatch.setattr(menu_app, "check_dependencies", lambda: None)
    monkeypatch.setenv("APP_VERSION", "test-commit")
    monkeypatch.setenv("HOSTNAME", "menu-pod")

    response = client.get("/api/menu/healthz")

    assert response.status_code == 200
    assert response.json() == {"status": "ok", "version": "test-commit", "pod": "menu-pod"}


def test_health_returns_unavailable_when_dependency_check_fails(monkeypatch):
    def fail_check():
        raise menu_app.redis.ConnectionError("redis unavailable")

    monkeypatch.setattr(menu_app, "check_dependencies", fail_check)

    response = client.get("/api/menu/healthz")

    assert response.status_code == 503


def test_livez_does_not_check_dependencies(monkeypatch):
    monkeypatch.setattr(
        menu_app,
        "check_dependencies",
        lambda: (_ for _ in ()).throw(AssertionError("liveness must not check dependencies")),
    )

    response = client.get("/api/menu/livez")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}
