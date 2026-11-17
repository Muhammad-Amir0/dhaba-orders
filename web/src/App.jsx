import React, { useCallback, useEffect, useMemo, useState } from 'react';

const TABS = ['Order', 'Kitchen', 'Stats'];
const ORDER_STATUSES = ['new', 'cooking', 'ready'];
const formatPKR = (amount) => `Rs ${Number(amount).toLocaleString('en-PK')}`;

async function getJSON(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Request failed (${response.status})`);
  }
  return response.json();
}

function usePolling(url, interval) {
  const [data, setData] = useState(null);
  const [error, setError] = useState('');

  const refresh = useCallback(async () => {
    try {
      setData(await getJSON(url));
      setError('');
    } catch (requestError) {
      setError(requestError.message);
    }
  }, [url]);

  useEffect(() => {
    refresh();
    const timer = window.setInterval(refresh, interval);
    return () => window.clearInterval(timer);
  }, [refresh, interval]);

  return { data, error, refresh };
}

function ErrorMessage({ children }) {
  return children ? <p className="error-message" role="alert">{children}</p> : null;
}

function MenuCard({ item, onAdd }) {
  return (
    <article className="menu-card">
      <div className="dish-mark" aria-hidden="true">{item.name.slice(0, 1)}</div>
      <div className="dish-details">
        <p className="eyebrow">{item.category}</p>
        <h3>{item.name}</h3>
        <p className="price">{formatPKR(item.price_pkr)}</p>
      </div>
      <button className="add-button" onClick={() => onAdd(item)} aria-label={`Add ${item.name}`}>
        +
      </button>
    </article>
  );
}

function OrderView({ menu, menuError, onPlaced }) {
  const [cart, setCart] = useState({});
  const [placing, setPlacing] = useState(false);
  const [orderError, setOrderError] = useState('');
  const cartItems = useMemo(
    () => Object.values(cart).filter((entry) => entry.qty > 0),
    [cart],
  );
  const total = cartItems.reduce((sum, entry) => sum + entry.item.price_pkr * entry.qty, 0);

  function addItem(item) {
    setCart((current) => ({
      ...current,
      [item.id]: { item, qty: (current[item.id]?.qty ?? 0) + 1 },
    }));
  }

  function changeQuantity(itemId, delta) {
    setCart((current) => {
      const currentQty = current[itemId]?.qty ?? 0;
      const qty = Math.max(0, currentQty + delta);
      if (qty === 0) {
        const next = { ...current };
        delete next[itemId];
        return next;
      }
      return { ...current, [itemId]: { ...current[itemId], qty } };
    });
  }

  async function placeOrder() {
    setPlacing(true);
    setOrderError('');
    try {
      const response = await fetch('/api/orders', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          items: cartItems.map(({ item, qty }) => ({ menu_item_id: item.id, qty })),
        }),
      });
      if (!response.ok) {
        const result = await response.json().catch(() => ({}));
        throw new Error(result.error ?? `Could not place order (${response.status})`);
      }
      const order = await response.json();
      setCart({});
      onPlaced(order);
    } catch (requestError) {
      setOrderError(requestError.message);
    } finally {
      setPlacing(false);
    }
  }

  return (
    <section className="order-layout">
      <div>
        <div className="section-heading">
          <div>
            <p className="eyebrow">Made fresh, every time</p>
            <h2>Our menu</h2>
          </div>
          <span className="menu-count">{menu.length} dishes</span>
        </div>
        <ErrorMessage>{menuError}</ErrorMessage>
        <div className="menu-grid">
          {menu.map((item) => <MenuCard key={item.id} item={item} onAdd={addItem} />)}
          {menu.length === 0 && !menuError && <p className="empty-state">Loading the menu…</p>}
        </div>
      </div>

      <aside className="cart-panel">
        <p className="eyebrow">Your selection</p>
        <h2>Your order</h2>
        {cartItems.length === 0 ? (
          <p className="empty-state">Add a dish to get started.</p>
        ) : (
          <ul className="cart-list">
            {cartItems.map(({ item, qty }) => (
              <li key={item.id}>
                <div>
                  <strong>{item.name}</strong>
                  <span>{formatPKR(item.price_pkr * qty)}</span>
                </div>
                <div className="quantity-control" aria-label={`${item.name} quantity`}>
                  <button onClick={() => changeQuantity(item.id, -1)} aria-label={`Remove one ${item.name}`}>−</button>
                  <span>{qty}</span>
                  <button onClick={() => changeQuantity(item.id, 1)} aria-label={`Add one ${item.name}`}>+</button>
                </div>
              </li>
            ))}
          </ul>
        )}
        <div className="cart-total">
          <span>Total</span>
          <strong>{formatPKR(total)}</strong>
        </div>
        <ErrorMessage>{orderError}</ErrorMessage>
        <button
          className="primary-button place-order"
          disabled={cartItems.length === 0 || placing}
          onClick={placeOrder}
        >
          {placing ? 'Placing order…' : 'Place order'}
        </button>
      </aside>
    </section>
  );
}

function KitchenView({ orders, error }) {
  return (
    <section>
      <div className="section-heading">
        <div>
          <p className="eyebrow">Live kitchen board</p>
          <h2>Orders in progress</h2>
        </div>
        <span className="live-dot">Auto-refresh · 2 sec</span>
      </div>
      <ErrorMessage>{error}</ErrorMessage>
      <div className="kitchen-board">
        {ORDER_STATUSES.map((status) => {
          const statusOrders = orders.filter((order) => order.status === status);
          return (
            <section className="status-column" key={status} aria-label={`${status} orders`}>
              <header>
                <h3>{status}</h3>
                <span>{statusOrders.length}</span>
              </header>
              {statusOrders.length === 0 && <p className="empty-state">Nothing here yet</p>}
              {statusOrders.map((order) => (
                <article className="order-ticket" key={order.id}>
                  <div className="ticket-heading">
                    <strong>Order #{order.id}</strong>
                    <span>{new Date(order.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                  </div>
                  <ul>
                    {order.items.map((item, index) => (
                      <li key={`${item.menu_item_id}-${index}`}>
                        <span>{item.qty} × {item.name}</span>
                      </li>
                    ))}
                  </ul>
                  <strong className="ticket-total">{formatPKR(order.total_pkr)}</strong>
                </article>
              ))}
            </section>
          );
        })}
      </div>
    </section>
  );
}

function StatsView({ stats, error }) {
  const dishes = stats?.top_dishes ?? [];
  const largestCount = Math.max(1, ...dishes.map((dish) => dish.count));

  return (
    <section>
      <div className="section-heading">
        <div>
          <p className="eyebrow">The numbers behind the spice</p>
          <h2>Today at a glance</h2>
        </div>
        <span className="live-dot">Auto-refresh · 3 sec</span>
      </div>
      <ErrorMessage>{error}</ErrorMessage>
      <div className="stats-layout">
        <article className="orders-stat">
          <p className="eyebrow">Orders this minute</p>
          <strong>{stats?.orders_per_minute ?? '—'}</strong>
          <span>Fresh orders, minute by minute.</span>
        </article>
        <article className="top-dishes">
          <p className="eyebrow">Customer favourites</p>
          <h3>Top five dishes</h3>
          {dishes.length === 0 && <p className="empty-state">No dish stats yet.</p>}
          <ol>
            {dishes.map((dish, index) => (
              <li key={dish.name}>
                <span className="dish-rank">{String(index + 1).padStart(2, '0')}</span>
                <div className="dish-bar">
                  <div>
                    <strong>{dish.name}</strong>
                    <span>{dish.count} ordered</span>
                  </div>
                  <div className="bar-track">
                    <span style={{ width: `${(dish.count / largestCount) * 100}%` }} />
                  </div>
                </div>
              </li>
            ))}
          </ol>
        </article>
      </div>
    </section>
  );
}

export default function App() {
  const [activeTab, setActiveTab] = useState('Order');
  const [placedMessage, setPlacedMessage] = useState('');
  const menuResult = usePolling('/api/menu', 60_000);
  const ordersResult = usePolling('/api/orders', 2_000);
  const statsResult = usePolling('/api/stats', 3_000);
  const menu = menuResult.data?.items ?? [];
  const orders = Array.isArray(ordersResult.data) ? ordersResult.data : [];

  function handlePlaced(order) {
    setPlacedMessage(`Order #${order.id} is in the kitchen.`);
    ordersResult.refresh();
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <a className="brand" href="/" aria-label="Dhaba Orders home">
          <span className="brand-mark">D</span>
          <span>dhaba<span className="brand-dot">.</span><small>ORDERS</small></span>
        </a>
        <span className="location-label"><span aria-hidden="true">●</span> Lahore · Kitchen open</span>
      </header>

      <section className="hero">
        <div className="hero-copy">
          <p className="eyebrow">A taste of home</p>
          <h1>Good food.<br /><em>Good mood.</em></h1>
          <p>From the karahi to your table, made with the kind of love you can taste.</p>
        </div>
        <div className="hero-art" aria-hidden="true">
          <div className="hero-ring" />
          <div className="hero-bowl">✳</div>
          <span className="hero-spice spice-one">✦</span>
          <span className="hero-spice spice-two">✧</span>
        </div>
      </section>

      <nav className="tabs" aria-label="Main views">
        {TABS.map((tab) => (
          <button
            key={tab}
            className={activeTab === tab ? 'tab active' : 'tab'}
            aria-current={activeTab === tab ? 'page' : undefined}
            onClick={() => setActiveTab(tab)}
          >
            {tab}
          </button>
        ))}
      </nav>

      <ErrorMessage>{placedMessage}</ErrorMessage>
      <div className="view-content">
        {activeTab === 'Order' && (
          <OrderView menu={menu} menuError={menuResult.error} onPlaced={handlePlaced} />
        )}
        {activeTab === 'Kitchen' && (
          <KitchenView orders={orders} error={ordersResult.error} />
        )}
        {activeTab === 'Stats' && (
          <StatsView stats={statsResult.data} error={statsResult.error} />
        )}
      </div>
      <footer>
        <span>Good food brings us together.</span>
        <span>Made with care in Lahore <span aria-hidden="true">♥</span></span>
      </footer>
    </main>
  );
}
